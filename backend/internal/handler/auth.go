package handler

import (
	"errors"
	"log"
	"net/http"

	"expense-app/backend/internal/apierror"
	"expense-app/backend/internal/auth"
	"expense-app/backend/internal/middleware"
	"expense-app/backend/internal/model"
	"expense-app/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type Auth struct{ service *service.Auth }
type credentialsRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,max=512"`
}
type googleRequest struct {
	IDToken string `json:"id_token" binding:"required,max=8192"`
}
type facebookRequest struct {
	AccessToken string `json:"access_token" binding:"required,max=4096"`
}

func RegisterAuth(router *gin.Engine, svc *service.Auth, tokens auth.Tokens) {
	h := &Auth{service: svc}
	group := router.Group("/api/v1/auth")
	group.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 12*1024)
		c.Next()
	})
	group.POST("/register", h.register)
	group.POST("/login", h.login)
	group.POST("/google", h.google)
	group.POST("/facebook", h.facebook)
	group.POST("/refresh", h.refresh)
	group.POST("/logout", h.logout)
	group.GET("/me", middleware.JWT(tokens), h.me)
}
func bind(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		apierror.Write(c, 400, "invalid_request", "Invalid JSON request body")
		return false
	}
	return true
}
func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		apierror.Write(c, 400, "invalid_request", "Provide a valid email and password of 8–72 bytes, or a valid refresh token")
	case errors.Is(err, model.ErrEmailTaken):
		apierror.Write(c, 409, "email_taken", "Email already registered")
	case errors.Is(err, service.ErrCredentials):
		apierror.Write(c, 401, "invalid_credentials", "Invalid email or password")
	case errors.Is(err, service.ErrGoogleToken):
		apierror.Write(c, 401, "invalid_google_token", "Invalid Google ID token")
	case errors.Is(err, service.ErrGoogleConfig):
		apierror.Write(c, 503, "google_login_unavailable", "Google login is not configured")
	case errors.Is(err, service.ErrFacebookToken):
		apierror.Write(c, 401, "invalid_facebook_token", "Invalid Facebook access token")
	case errors.Is(err, service.ErrFacebookConfig):
		apierror.Write(c, 503, "facebook_login_unavailable", "Facebook login is not configured")
	case errors.Is(err, service.ErrFacebookUnavailable):
		apierror.Write(c, 503, "facebook_provider_unavailable", "Facebook verification is temporarily unavailable")
	case errors.Is(err, model.ErrRefreshInvalid):
		apierror.Write(c, 401, "invalid_refresh_token", "Invalid or expired refresh token")
	case errors.Is(err, model.ErrNotFound):
		apierror.Write(c, 401, "invalid_access_token", "User is no longer available")
	default:
		// Avoid logging driver errors that may include password or token hashes.
		log.Printf("auth operation failed (%T)", err)
		apierror.Write(c, 500, "internal_error", "An internal error occurred")
	}
}
func (h *Auth) google(c *gin.Context) {
	var req googleRequest
	if !bind(c, &req) {
		return
	}
	session, err := h.service.GoogleLogin(c.Request.Context(), req.IDToken)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}
func (h *Auth) facebook(c *gin.Context) {
	var req facebookRequest
	if !bind(c, &req) {
		return
	}
	session, err := h.service.FacebookLogin(c.Request.Context(), req.AccessToken)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}
func (h *Auth) register(c *gin.Context) {
	var req credentialsRequest
	if !bind(c, &req) {
		return
	}
	session, err := h.service.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(201, session)
}
func (h *Auth) login(c *gin.Context) {
	var req credentialsRequest
	if !bind(c, &req) {
		return
	}
	session, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, session)
}
func (h *Auth) refresh(c *gin.Context) {
	var req refreshRequest
	if !bind(c, &req) {
		return
	}
	session, err := h.service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, session)
}
func (h *Auth) logout(c *gin.Context) {
	var req refreshRequest
	if !bind(c, &req) {
		return
	}
	if err := h.service.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		respondError(c, err)
		return
	}
	c.Status(204)
}
func (h *Auth) me(c *gin.Context) {
	user, err := h.service.Me(c.Request.Context(), c.GetString(middleware.UserIDKey))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(200, gin.H{"user": user})
}
