package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
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

// Use a migrated, disposable PostgreSQL database. Never migrate/drop a caller's DB here.
func TestAuthPostgres(t *testing.T) {
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
	tokens := auth.Tokens{Config: config.Auth{Secret: strings.Repeat("t", 32), Issuer: "test", Audience: "mobile", AccessTTL: time.Minute, RefreshTTL: time.Hour}}
	svc, err := service.NewAuth(repository.NewAuth(db), tokens)
	if err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterAuth(router, svc, tokens)
	call := func(path string, body any, bearer string) *httptest.ResponseRecorder {
		data, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		method := "POST"
		if path == "/me" {
			method = "GET"
		}
		req := httptest.NewRequest(method, "/api/v1/auth"+path, strings.NewReader(string(data)))
		req.Header.Set("Content-Type", "application/json")
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		return res
	}
	expect := func(res *httptest.ResponseRecorder, status int) {
		t.Helper()
		if res.Code != status {
			t.Fatalf("got status %d want %d: %s", res.Code, status, res.Body.String())
		}
	}
	decode := func(res *httptest.ResponseRecorder) service.Session {
		t.Helper()
		var s service.Session
		if err := json.Unmarshal(res.Body.Bytes(), &s); err != nil {
			t.Fatal(err)
		}
		return s
	}
	email := uuid.NewString() + "@example.com"
	creds := map[string]string{"email": strings.ToUpper(email), "password": "test-password-123"}
	res := call("/register", creds, "")
	expect(res, 201)
	original := decode(res)
	if original.User.Email == nil || *original.User.Email != email || original.AccessToken == "" || original.RefreshToken == "" {
		t.Fatal("incomplete register response")
	}
	if strings.Contains(res.Body.String(), "hash") || strings.Contains(res.Body.String(), "password") {
		t.Fatal("credential hash exposed")
	}
	if res.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("auth response cacheable")
	}
	expect(call("/register", creds, ""), 409)
	var count int64
	if err := db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("duplicate register left orphan user")
	}
	var stored model.RefreshToken
	if err := db.Where("token_hash = ?", auth.Hash(original.RefreshToken)).Take(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if stored.TokenHash == original.RefreshToken || stored.UserID != original.User.ID {
		t.Fatal("unsafe refresh storage")
	}
	expect(call("/me", nil, ""), 401)
	expect(call("/me", nil, original.AccessToken), 200)
	expect(call("/login", map[string]string{"email": email, "password": "wrong-password"}, ""), 401)
	expect(call("/login", map[string]string{"email": "missing-" + email, "password": "wrong-password"}, ""), 401)
	expect(call("/register", map[string]string{"email": "bad", "password": "short"}, ""), 400)
	expect(call("/login", map[string]string{}, ""), 400)
	expect(call("/login", map[string]string{"email": email, "password": strings.Repeat("x", 5000)}, ""), 400)
	res = call("/login", creds, "")
	expect(res, 200)
	login := decode(res)
	payload := func(raw string) map[string]string { return map[string]string{"refresh_token": raw} }
	// Both callers race for the same refresh row; exactly one must commit.
	var wg sync.WaitGroup
	results := make(chan *httptest.ResponseRecorder, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); results <- call("/refresh", payload(login.RefreshToken), "") }()
	}
	wg.Wait()
	close(results)
	successes, denied := 0, 0
	var rotated service.Session
	for res := range results {
		switch res.Code {
		case 200:
			successes++
			rotated = decode(res)
		case 401:
			denied++
		default:
			t.Fatalf("concurrent refresh: %d", res.Code)
		}
	}
	if successes != 1 || denied != 1 {
		t.Fatalf("refresh race: successes=%d denied=%d", successes, denied)
	}
	if rotated.RefreshToken == login.RefreshToken {
		t.Fatal("refresh not rotated")
	}
	expect(call("/refresh", payload(login.RefreshToken), ""), 401)
	expect(call("/me", nil, rotated.AccessToken), 200)
	expect(call("/logout", payload(rotated.RefreshToken), ""), 204)
	expect(call("/logout", payload(rotated.RefreshToken), ""), 204)
	expect(call("/refresh", payload(rotated.RefreshToken), ""), 401)
	expect(call("/me", nil, rotated.AccessToken), 200) // logout does not revoke existing access JWTs
	expect(call("/refresh", payload("malformed"), ""), 401)
	unknown, err := auth.NewRefresh()
	if err != nil {
		t.Fatal(err)
	}
	expect(call("/refresh", payload(unknown), ""), 401)
	expect(call("/logout", payload(unknown), ""), 204)
	// One device's logout must not revoke another device's refresh token.
	res = call("/refresh", payload(original.RefreshToken), "")
	expect(res, 200)
	other := decode(res)
	if err := db.Model(&model.RefreshToken{}).Where("token_hash = ?", auth.Hash(other.RefreshToken)).Update("expires_at", time.Now().Add(-time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	expect(call("/refresh", payload(other.RefreshToken), ""), 401)
	// A failed rotation insert must roll back the old token's revocation.
	res = call("/login", creds, "")
	expect(res, 200)
	rollback := decode(res)
	colliding := model.RefreshToken{ID: uuid.NewString(), TokenHash: stored.TokenHash, ExpiresAt: time.Now().Add(time.Hour)}
	if _, err := repository.NewAuth(db).Rotate(ctx, auth.Hash(rollback.RefreshToken), &colliding); err == nil {
		t.Fatal("expected unique constraint failure")
	}
	expect(call("/refresh", payload(rollback.RefreshToken), ""), 200)
}
