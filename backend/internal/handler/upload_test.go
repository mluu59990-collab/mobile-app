package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"expense-app/backend/internal/auth"
	"expense-app/backend/internal/config"
	"expense-app/backend/internal/service"
	"github.com/gin-gonic/gin"
)

func TestPresignRejectsInvalidTypeAndSizeWithErrorEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := auth.Tokens{Config: config.Auth{Secret: strings.Repeat("upload-test-secret-", 3), Issuer: "test", Audience: "test", AccessTTL: time.Hour}}
	token, err := tokens.Access("00000000-0000-4000-8000-000000000001")
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	svc := service.NewExpenseServiceWithImages(nil, nil, service.ExpenseImageConfig{PutURLTTL: 10 * time.Minute, GetURLTTL: 15 * time.Minute})
	RegisterUploadRoutes(router, svc, tokens)

	for _, body := range []string{
		`{"content_type":"image/gif","size":123}`,
		`{"content_type":"image/png","size":10485761}`,
	} {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/uploads/presign", strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer "+token)
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400: %s", response.Code, response.Body.String())
		}
		var envelope struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		if envelope.Error.Code != "invalid_upload" || envelope.Error.Message == "" {
			t.Fatalf("bad error envelope: %+v", envelope)
		}
	}
}
