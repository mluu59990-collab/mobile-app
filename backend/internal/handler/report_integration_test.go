package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"expense-app/backend/internal/auth"
	"expense-app/backend/internal/config"
	"expense-app/backend/internal/model"
	"expense-app/backend/internal/repository"
	"expense-app/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestMonthlyReportIsUserScopedAndAggregatedPostgres(t *testing.T) {
	dsn := os.Getenv("AUTH_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set AUTH_TEST_DATABASE_URL to a migrated disposable PostgreSQL DB")
	}
	ctx := context.Background()
	db, err := repository.OpenDatabase(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Error(err)
		}
	})

	users := []model.User{{ID: uuid.NewString()}, {ID: uuid.NewString()}}
	for i := range users {
		email := uuid.NewString() + "@example.com"
		users[i].Email = &email
		if err := db.WithContext(ctx).Create(&users[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	food := "food"
	expenses := []model.Expense{
		{ID: uuid.NewString(), UserID: users[0].ID, Amount: 100, Category: &food, SpentAt: time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.NewString(), UserID: users[0].ID, Amount: 500, Category: &food, SpentAt: time.Date(2025, 5, 1, 23, 59, 59, 0, time.UTC)},
		{ID: uuid.NewString(), UserID: users[0].ID, Amount: 200, SpentAt: time.Date(2025, 5, 2, 12, 0, 0, 0, time.UTC)},
		{ID: uuid.NewString(), UserID: users[0].ID, Amount: 900, SpentAt: time.Date(2025, 4, 30, 23, 59, 59, 0, time.UTC)},
		{ID: uuid.NewString(), UserID: users[0].ID, Amount: 700, SpentAt: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)},
		{ID: uuid.NewString(), UserID: users[1].ID, Amount: 10000, Category: &food, SpentAt: time.Date(2025, 5, 15, 10, 0, 0, 0, time.UTC)},
		{ID: uuid.NewString(), UserID: users[0].ID, Amount: 1234, SpentAt: time.Date(2026, 10, 31, 17, 30, 0, 0, time.UTC)},
	}
	for i := range expenses {
		if err := db.WithContext(ctx).Create(&expenses[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	gin.SetMode(gin.TestMode)
	tokens := auth.Tokens{Config: config.Auth{Secret: strings.Repeat("monthly-report-integration-test-key-", 2), Issuer: "test", Audience: "test", AccessTTL: time.Hour}}
	router := gin.New()
	RegisterReportRoutes(router, service.NewMonthlyReportService(repository.NewReports(db)), tokens)
	access := func(userID string) string {
		token, err := tokens.Access(userID)
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	request := func(userID, month, timezone string) *httptest.ResponseRecorder {
		path := "/api/v1/reports/monthly?month=" + month
		if timezone != "" {
			path += "&tz=" + timezone
		}
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if userID != "" {
			req.Header.Set("Authorization", "Bearer "+access(userID))
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response
	}

	userA := request(users[0].ID, "2025-05", "")
	if userA.Code != http.StatusOK {
		t.Fatalf("user A report returned %d: %s", userA.Code, userA.Body.String())
	}
	var report struct {
		Month       string `json:"month"`
		TotalAmount int64  `json:"total_amount"`
		ByCategory  []struct {
			Category    *string `json:"category"`
			TotalAmount int64   `json:"total_amount"`
		} `json:"by_category"`
		ByDay []struct {
			Date        string `json:"date"`
			TotalAmount int64  `json:"total_amount"`
		} `json:"by_day"`
	}
	if err := json.Unmarshal(userA.Body.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Month != "2025-05" || report.TotalAmount != 800 {
		t.Fatalf("wrong month total or boundary handling: %+v", report)
	}
	if len(report.ByCategory) != 2 || report.ByCategory[0].Category == nil || *report.ByCategory[0].Category != food || report.ByCategory[0].TotalAmount != 600 || report.ByCategory[1].Category != nil || report.ByCategory[1].TotalAmount != 200 {
		t.Fatalf("wrong per-category totals: %+v", report.ByCategory)
	}
	if len(report.ByDay) != 2 || report.ByDay[0].Date != "2025-05-01" || report.ByDay[0].TotalAmount != 600 || report.ByDay[1].Date != "2025-05-02" || report.ByDay[1].TotalAmount != 200 {
		t.Fatalf("wrong per-day totals: %+v", report.ByDay)
	}

	userB := request(users[1].ID, "2025-05", "")
	if userB.Code != http.StatusOK {
		t.Fatalf("user B report returned %d: %s", userB.Code, userB.Body.String())
	}
	var bReport struct {
		TotalAmount int64 `json:"total_amount"`
	}
	if err := json.Unmarshal(userB.Body.Bytes(), &bReport); err != nil {
		t.Fatal(err)
	}
	if bReport.TotalAmount != 10000 || bReport.TotalAmount == report.TotalAmount {
		t.Fatalf("user B report is mixed with user A data: %+v", bReport)
	}
	if got := request(users[0].ID, "2025-13", ""); got.Code != http.StatusBadRequest {
		t.Fatalf("invalid month returned %d: %s", got.Code, got.Body.String())
	}
	if got := request("", "2025-05", ""); got.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated report returned %d", got.Code)
	}
	checkBoundary := func(month, timezone, wantDay string) {
		t.Helper()
		got := request(users[0].ID, month, timezone)
		if got.Code != http.StatusOK {
			t.Fatalf("boundary report returned %d: %s", got.Code, got.Body.String())
		}
		var result struct {
			TotalAmount int64 `json:"total_amount"`
			ByDay       []struct {
				Date        string `json:"date"`
				TotalAmount int64  `json:"total_amount"`
			} `json:"by_day"`
		}
		if err := json.Unmarshal(got.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.TotalAmount != 1234 || len(result.ByDay) != 1 || result.ByDay[0].Date != wantDay || result.ByDay[0].TotalAmount != 1234 {
			t.Fatalf("unexpected timezone boundary aggregation for month=%s tz=%s: %+v", month, timezone, result)
		}
	}
	checkBoundary("2026-11", "Asia%2FHo_Chi_Minh", "2026-11-01")
	checkBoundary("2026-10", "", "2026-10-31")
	if got := request(users[0].ID, "2026-11", "Not%2FA_Timezone"); got.Code != http.StatusBadRequest || !strings.Contains(got.Body.String(), `"error"`) {
		t.Fatalf("invalid timezone should return error envelope 400, got %d: %s", got.Code, got.Body.String())
	}
}
