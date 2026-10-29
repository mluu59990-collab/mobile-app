package handler

import (
	"errors"
	"net/http"

	"expense-app/backend/internal/apierror"
	"expense-app/backend/internal/auth"
	"expense-app/backend/internal/middleware"
	"expense-app/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type Upload struct{ expenses *service.ExpenseService }

type presignImageRequest struct {
	ContentType string `json:"content_type" binding:"required"`
	Size        int64  `json:"size" binding:"required,gt=0,lte=10485760"`
}

func RegisterUploadRoutes(router *gin.Engine, expenses *service.ExpenseService, tokens auth.Tokens) {
	h := &Upload{expenses: expenses}
	router.Group("/api/v1/uploads", middleware.JWT(tokens)).POST("/presign", h.presign)
}

func (h *Upload) presign(c *gin.Context) {
	var request presignImageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		apierror.Write(c, http.StatusBadRequest, "invalid_upload", "Provide an allowed image content_type and size between 1 byte and 10 MiB")
		return
	}
	result, err := h.expenses.PresignImageUpload(c.Request.Context(), c.GetString(middleware.UserIDKey), request.ContentType, request.Size)
	if err != nil {
		writeImageError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func writeImageError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidImage), errors.Is(err, service.ErrInvalidImageKey):
		apierror.Write(c, http.StatusBadRequest, "invalid_upload", "Unsupported image content type or size")
	case err == service.ErrImageStorageUnavailable:
		apierror.Write(c, http.StatusServiceUnavailable, "image_storage_unavailable", "Image storage is not configured")
	default:
		respondError(c, err)
	}
}
