package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"expense-app/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

func TestRouterFormatsNotFoundAndMethodNotAllowedAsJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, err := newRouter(nil, nil, nil, auth.Tokens{})
	if err != nil {
		t.Fatalf("create router: %v", err)
	}

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantCode   string
	}{
		{name: "route not found", method: http.MethodGet, path: "/api/v1/missing", wantStatus: http.StatusNotFound, wantCode: "route_not_found"},
		{name: "method not allowed", method: http.MethodPost, path: "/api/v1/health", wantStatus: http.StatusMethodNotAllowed, wantCode: "method_not_allowed"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.path, nil)
			router.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", response.Code, test.wantStatus, response.Body.String())
			}
			if contentType := response.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
				t.Fatalf("Content-Type = %q, want JSON", contentType)
			}

			var body struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode error envelope: %v", err)
			}
			if body.Error.Code != test.wantCode || body.Error.Message == "" {
				t.Fatalf("error = %+v, want code %q and non-empty message", body.Error, test.wantCode)
			}
		})
	}
}
