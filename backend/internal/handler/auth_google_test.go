package handler

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"expense-app/backend/internal/auth"
	"expense-app/backend/internal/config"
	"expense-app/backend/internal/model"
	"expense-app/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type googleRouteRepo struct {
	service.AuthRepository
	user model.User
}

func (r *googleRouteRepo) GoogleUser(context.Context, string, *string, bool) (model.User, error) {
	return r.user, nil
}
func (r *googleRouteRepo) FacebookUser(_ context.Context, _ string, email *string) (model.User, error) {
	r.user.Email = email
	return r.user, nil
}
func (r *googleRouteRepo) SaveRefresh(context.Context, *model.RefreshToken) error { return nil }

type googleRouteVerifier struct{ err error }

func (v googleRouteVerifier) Verify(_ context.Context, raw, audience string) (service.GoogleIdentity, error) {
	if v.err != nil {
		return service.GoogleIdentity{}, v.err
	}
	if raw != "fake-id-token" || audience != "client-id" {
		return service.GoogleIdentity{}, errors.New("unexpected verification input")
	}
	return service.GoogleIdentity{Subject: "google-subject", Email: "user@gmail.com", EmailVerified: true}, nil
}

type facebookRouteVerifier struct{ err error }

func (v facebookRouteVerifier) Verify(_ context.Context, raw string) (service.FacebookIdentity, error) {
	if v.err != nil {
		return service.FacebookIdentity{}, v.err
	}
	if raw != "fake-facebook-token" {
		return service.FacebookIdentity{}, errors.New("unexpected token")
	}
	return service.FacebookIdentity{Subject: "facebook-subject"}, nil
}

func TestGoogleLoginRouteUsesVerifierAndReturnsAppTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	email := "user@gmail.com"
	repo := &googleRouteRepo{user: model.User{ID: "user-id", Email: &email}}
	tokens := auth.Tokens{Config: config.Auth{Secret: strings.Repeat("s", 32), Issuer: "test", Audience: "mobile", GoogleClientID: "client-id", AccessTTL: time.Minute, RefreshTTL: time.Hour}}
	svc, err := service.NewAuth(repo, tokens, googleRouteVerifier{})
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	RegisterAuth(router, svc, tokens)
	request := httptest.NewRequest("POST", "/api/v1/auth/google", strings.NewReader(`{"id_token":"fake-id-token"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != 200 {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"access_token"`) || !strings.Contains(response.Body.String(), `"refresh_token"`) || strings.Contains(response.Body.String(), `"provider_token"`) {
		t.Fatalf("expected application token pair, got %s", response.Body.String())
	}
}

func TestGoogleLoginRouteRejectsInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &googleRouteRepo{}
	tokens := auth.Tokens{Config: config.Auth{Secret: strings.Repeat("s", 32), GoogleClientID: "client-id", AccessTTL: time.Minute, RefreshTTL: time.Hour}}
	svc, err := service.NewAuth(repo, tokens, googleRouteVerifier{err: errors.New("bad signature")})
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	RegisterAuth(router, svc, tokens)
	request := httptest.NewRequest("POST", "/api/v1/auth/google", strings.NewReader(`{"id_token":"fake-id-token"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != 401 || !strings.Contains(response.Body.String(), `"code":"invalid_google_token"`) {
		t.Fatalf("invalid token not rejected: %d %s", response.Code, response.Body.String())
	}
}

func TestFacebookLoginRouteAllowsNullEmailAndReturnsAppTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &googleRouteRepo{user: model.User{ID: "facebook-user"}}
	tokens := auth.Tokens{Config: config.Auth{Secret: strings.Repeat("s", 32), FacebookAppID: "app-id", FacebookAppSecret: "app-secret", AccessTTL: time.Minute, RefreshTTL: time.Hour}}
	svc, err := service.NewAuth(repo, tokens)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetFacebookVerifier(facebookRouteVerifier{})
	router := gin.New()
	RegisterAuth(router, svc, tokens)
	request := httptest.NewRequest("POST", "/api/v1/auth/facebook", strings.NewReader(`{"access_token":"fake-facebook-token"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != 200 {
		t.Fatalf("status %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"email":null`) || !strings.Contains(response.Body.String(), `"access_token"`) || !strings.Contains(response.Body.String(), `"refresh_token"`) {
		t.Fatalf("expected nullable email and app session, got %s", response.Body.String())
	}
}

func TestFacebookLoginRouteRejectsInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &googleRouteRepo{}
	tokens := auth.Tokens{Config: config.Auth{Secret: strings.Repeat("s", 32), FacebookAppID: "app-id", FacebookAppSecret: "app-secret", AccessTTL: time.Minute, RefreshTTL: time.Hour}}
	svc, err := service.NewAuth(repo, tokens)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetFacebookVerifier(facebookRouteVerifier{err: errors.New("invalid token")})
	router := gin.New()
	RegisterAuth(router, svc, tokens)
	request := httptest.NewRequest("POST", "/api/v1/auth/facebook", strings.NewReader(`{"access_token":"fake-facebook-token"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != 401 || !strings.Contains(response.Body.String(), `"code":"invalid_facebook_token"`) {
		t.Fatalf("invalid token not rejected: %d %s", response.Code, response.Body.String())
	}
}

func TestFacebookLoginRouteReturnsUnavailableWhenGraphAPIIsDown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &googleRouteRepo{}
	tokens := auth.Tokens{Config: config.Auth{Secret: strings.Repeat("s", 32), FacebookAppID: "app-id", FacebookAppSecret: "app-secret", AccessTTL: time.Minute, RefreshTTL: time.Hour}}
	svc, err := service.NewAuth(repo, tokens)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetFacebookVerifier(facebookRouteVerifier{err: service.ErrFacebookUnavailable})
	router := gin.New()
	RegisterAuth(router, svc, tokens)
	request := httptest.NewRequest("POST", "/api/v1/auth/facebook", strings.NewReader(`{"access_token":"fake-facebook-token"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != 503 || !strings.Contains(response.Body.String(), `"code":"facebook_provider_unavailable"`) {
		t.Fatalf("provider outage not reported as unavailable: %d %s", response.Code, response.Body.String())
	}
}
