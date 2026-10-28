package middleware

import (
	"expense-app/backend/internal/auth"
	"expense-app/backend/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestJWTMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := auth.Tokens{Config: config.Auth{Secret: strings.Repeat("x", 32), Issuer: "test", Audience: "test", AccessTTL: time.Minute}}
	id := uuid.NewString()
	raw, err := tokens.Access(id)
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.GET("/private", JWT(tokens), func(c *gin.Context) { c.String(200, c.GetString(UserIDKey)) })
	for _, header := range []string{"", "Bearer", "Basic abc", "Bearer broken", "Bearer " + raw + " extra", "Bearer " + raw} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/private", nil)
		req.Header.Set("Authorization", header)
		r.ServeHTTP(w, req)
		want := 401
		if header == "Bearer "+raw {
			want = 200
		}
		if w.Code != want {
			t.Fatalf("got %d want %d", w.Code, want)
		}
		if want == 200 && w.Body.String() != id {
			t.Fatal("user context missing")
		}
		if want == 401 && !strings.Contains(w.Body.String(), `"error":{"code":"invalid_access_token"`) {
			t.Fatal("error envelope missing")
		}
	}
}
