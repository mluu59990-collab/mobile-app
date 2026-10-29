package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
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

func TestExpenseOwnershipAndCRUDPostgres(t *testing.T) {
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

	gin.SetMode(gin.TestMode)
	tokens := auth.Tokens{Config: config.Auth{Secret: strings.Repeat("expense-integration-test-key-", 2), Issuer: "test", Audience: "test", AccessTTL: time.Hour}}
	router := gin.New()
	RegisterExpenseRoutes(router, service.NewExpenseService(repository.NewExpenses(db)), tokens)

	request := func(method, path, token, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response
	}
	access := func(userID string) string {
		value, err := tokens.Access(userID)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	checkStatus := func(response *httptest.ResponseRecorder, want int) {
		t.Helper()
		if response.Code != want {
			t.Fatalf("got status %d, want %d: %s", response.Code, want, response.Body.String())
		}
	}
	userAToken, userBToken := access(users[0].ID), access(users[1].ID)

	mayExpense := request(http.MethodPost, "/api/v1/expenses", userAToken, `{"amount":25000,"note":" coffee ","category":"food","spent_at":"2025-05-31T23:59:59Z"}`)
	checkStatus(mayExpense, http.StatusCreated)
	var created expenseResponse
	if err := json.Unmarshal(mayExpense.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.Amount != 25000 || created.Note == nil || *created.Note != "coffee" {
		t.Fatalf("unexpected created expense: %+v", created)
	}
	var manualExpense model.Expense
	if err := db.WithContext(ctx).Where("id = ?", created.ID).Take(&manualExpense).Error; err != nil {
		t.Fatal(err)
	}
	if manualExpense.Source != "manual" {
		t.Fatalf("default source = %q, want manual", manualExpense.Source)
	}

	imageKey := "users/" + users[0].ID + "/legacy/receipt.png"
	rawAIResult := json.RawMessage(`{"merchant":"Test Store","total":25000}`)
	metadataExpense := model.Expense{
		ID: uuid.NewString(), UserID: users[0].ID, Amount: 25000,
		ImageKey: &imageKey, Source: "receipt_ai", RawAIResult: rawAIResult,
		SpentAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := db.WithContext(ctx).Create(&metadataExpense).Error; err != nil {
		t.Fatalf("create expense with migration metadata: %v", err)
	}
	legacyImageURL := "https://legacy.example.test/should-not-be-read.png"
	if err := db.WithContext(ctx).Exec("UPDATE expenses SET image_url = ? WHERE id = ?", legacyImageURL, metadataExpense.ID).Error; err != nil {
		t.Fatalf("set legacy image_url column for ignored-field assertion: %v", err)
	}
	var loadedMetadata model.Expense
	if err := db.WithContext(ctx).Where("id = ?", metadataExpense.ID).Take(&loadedMetadata).Error; err != nil {
		t.Fatalf("load expense with migration metadata: %v", err)
	}
	var gotAIResult, wantAIResult any
	if err := json.Unmarshal(loadedMetadata.RawAIResult, &gotAIResult); err != nil {
		t.Fatalf("decode stored raw AI result: %v", err)
	}
	if err := json.Unmarshal(rawAIResult, &wantAIResult); err != nil {
		t.Fatalf("decode expected raw AI result: %v", err)
	}
	if loadedMetadata.ImageKey == nil || *loadedMetadata.ImageKey != imageKey || loadedMetadata.ImageURL != nil || loadedMetadata.Source != "receipt_ai" || !reflect.DeepEqual(gotAIResult, wantAIResult) {
		t.Fatalf("expense migration metadata did not round-trip: %+v", loadedMetadata)
	}

	userBExpense := request(http.MethodPost, "/api/v1/expenses", userBToken, `{"amount":99000,"spent_at":"2025-06-01T00:00:00Z"}`)
	checkStatus(userBExpense, http.StatusCreated)

	checkStatus(request(http.MethodGet, "/api/v1/expenses?month=2025-05", userAToken, ""), http.StatusOK)
	var mayList struct {
		Expenses []expenseResponse `json:"expenses"`
	}
	if err := json.Unmarshal(request(http.MethodGet, "/api/v1/expenses?month=2025-05", userAToken, "").Body.Bytes(), &mayList); err != nil {
		t.Fatal(err)
	}
	if len(mayList.Expenses) != 1 || mayList.Expenses[0].ID != created.ID {
		t.Fatalf("user A month list leaked or missed data: %+v", mayList)
	}
	var bMayList struct {
		Expenses []expenseResponse `json:"expenses"`
	}
	if err := json.Unmarshal(request(http.MethodGet, "/api/v1/expenses?month=2025-05", userBToken, "").Body.Bytes(), &bMayList); err != nil {
		t.Fatal(err)
	}
	if len(bMayList.Expenses) != 0 {
		t.Fatalf("user B saw user A expense: %+v", bMayList)
	}
	checkStatus(request(http.MethodGet, "/api/v1/expenses/"+created.ID, userAToken, ""), http.StatusOK)
	checkStatus(request(http.MethodGet, "/api/v1/expenses/"+created.ID, userBToken, ""), http.StatusNotFound)

	checkStatus(request(http.MethodPatch, "/api/v1/expenses/"+created.ID, userBToken, `{"amount":1}`), http.StatusNotFound)
	var unchanged expenseResponse
	if err := json.Unmarshal(request(http.MethodGet, "/api/v1/expenses/"+created.ID, userAToken, "").Body.Bytes(), &unchanged); err != nil {
		t.Fatal(err)
	}
	if unchanged.Amount != 25000 {
		t.Fatalf("user B changed user A expense: %+v", unchanged)
	}

	updated := request(http.MethodPatch, "/api/v1/expenses/"+created.ID, userAToken, `{"amount":27000,"note":""}`)
	checkStatus(updated, http.StatusOK)
	var updatedExpense expenseResponse
	if err := json.Unmarshal(updated.Body.Bytes(), &updatedExpense); err != nil {
		t.Fatal(err)
	}
	if updatedExpense.Amount != 27000 || updatedExpense.Note != nil {
		t.Fatalf("update did not apply fields: %+v", updatedExpense)
	}
	checkStatus(request(http.MethodGet, "/api/v1/expenses?month=2025-06", userAToken, ""), http.StatusOK)
	var juneList struct {
		Expenses []expenseResponse `json:"expenses"`
	}
	if err := json.Unmarshal(request(http.MethodGet, "/api/v1/expenses?month=2025-06", userAToken, "").Body.Bytes(), &juneList); err != nil {
		t.Fatal(err)
	}
	if len(juneList.Expenses) != 0 {
		t.Fatalf("month boundary included prior month expense: %+v", juneList)
	}
	checkStatus(request(http.MethodGet, "/api/v1/expenses?month=2025-13", userAToken, ""), http.StatusBadRequest)
	boundaryAt := "2026-10-31T17:30:00Z"
	checkStatus(request(http.MethodPost, "/api/v1/expenses", userAToken, `{"amount":1234,"spent_at":"`+boundaryAt+`"}`), http.StatusCreated)
	for _, path := range []string{"/api/v1/expenses?month=2026-11&tz=Asia%2FHo_Chi_Minh", "/api/v1/expenses?month=2026-10"} {
		got := request(http.MethodGet, path, userAToken, "")
		checkStatus(got, http.StatusOK)
		var result struct {
			Expenses []expenseResponse `json:"expenses"`
		}
		if err := json.Unmarshal(got.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if len(result.Expenses) != 1 || result.Expenses[0].SpentAt.UTC().Format(time.RFC3339) != boundaryAt {
			t.Fatalf("timezone month %s did not include expected boundary expense: %+v", path, result.Expenses)
		}
	}
	badTimezone := request(http.MethodGet, "/api/v1/expenses?month=2026-11&tz=Not%2FA_Timezone", userAToken, "")
	if badTimezone.Code != http.StatusBadRequest || !strings.Contains(badTimezone.Body.String(), `"error"`) {
		t.Fatalf("invalid timezone should return error envelope 400, got %d: %s", badTimezone.Code, badTimezone.Body.String())
	}
	checkStatus(request(http.MethodGet, "/api/v1/expenses/not-a-uuid", userAToken, ""), http.StatusBadRequest)
	checkStatus(request(http.MethodGet, "/api/v1/expenses", "", ""), http.StatusUnauthorized)

	checkStatus(request(http.MethodDelete, "/api/v1/expenses/"+created.ID, userBToken, ""), http.StatusNotFound)
	checkStatus(request(http.MethodGet, "/api/v1/expenses/"+created.ID, userAToken, ""), http.StatusOK)
	checkStatus(request(http.MethodDelete, "/api/v1/expenses/"+created.ID, userAToken, ""), http.StatusNoContent)
	checkStatus(request(http.MethodGet, "/api/v1/expenses/"+created.ID, userAToken, ""), http.StatusNotFound)
}
