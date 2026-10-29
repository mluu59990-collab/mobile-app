package handler

import (
	"errors"
	"net/http"
	"time"

	"expense-app/backend/internal/apierror"
	"expense-app/backend/internal/auth"
	"expense-app/backend/internal/middleware"
	"expense-app/backend/internal/model"
	"expense-app/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Expense struct{ service *service.ExpenseService }

type expenseResponse struct {
	ID        string    `json:"id"`
	Amount    int64     `json:"amount"`
	Note      *string   `json:"note"`
	Category  *string   `json:"category"`
	ImageURL  *string   `json:"image_url"`
	SpentAt   time.Time `json:"spent_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type createExpenseRequest struct {
	Amount   int64      `json:"amount" binding:"required,gt=0"`
	Note     *string    `json:"note" binding:"omitempty,max=2000"`
	Category *string    `json:"category" binding:"omitempty,max=100"`
	SpentAt  *time.Time `json:"spent_at"`
	ImageKey *string    `json:"image_key"`
}

type updateExpenseRequest struct {
	Amount   *int64     `json:"amount" binding:"omitempty,gt=0"`
	Note     *string    `json:"note" binding:"omitempty,max=2000"`
	Category *string    `json:"category" binding:"omitempty,max=100"`
	SpentAt  *time.Time `json:"spent_at"`
	ImageKey *string    `json:"image_key"`
}

func RegisterExpenseRoutes(router *gin.Engine, svc *service.ExpenseService, tokens auth.Tokens) {
	h := &Expense{service: svc}
	group := router.Group("/api/v1/expenses")
	group.Use(middleware.JWT(tokens))
	group.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32*1024)
		c.Next()
	})
	group.POST("", h.create)
	group.GET("", h.list)
	group.GET("/:id", h.get)
	group.PATCH("/:id", h.update)
	group.DELETE("/:id", h.delete)
}

func (h *Expense) create(c *gin.Context) {
	var req createExpenseRequest
	if !bind(c, &req) {
		return
	}
	expense, err := h.service.Create(c.Request.Context(), c.GetString(middleware.UserIDKey), service.CreateExpenseInput{
		Amount: req.Amount, Note: req.Note, Category: req.Category, SpentAt: req.SpentAt, ImageKey: req.ImageKey,
	})
	if err != nil {
		expenseError(c, err)
		return
	}
	view, err := h.expenseView(c, expense)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, view)
}

func (h *Expense) list(c *gin.Context) {
	expenses, err := h.service.List(c.Request.Context(), c.GetString(middleware.UserIDKey), c.Query("month"), c.Query("tz"))
	if err != nil {
		expenseError(c, err)
		return
	}
	result := make([]expenseResponse, 0, len(expenses))
	for _, expense := range expenses {
		view, err := h.expenseView(c, expense)
		if err != nil {
			respondError(c, err)
			return
		}
		result = append(result, view)
	}
	c.JSON(http.StatusOK, gin.H{"expenses": result})
}

func (h *Expense) get(c *gin.Context) {
	id, ok := expenseID(c)
	if !ok {
		return
	}
	expense, err := h.service.Get(c.Request.Context(), c.GetString(middleware.UserIDKey), id)
	if err != nil {
		expenseError(c, err)
		return
	}
	view, err := h.expenseView(c, expense)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *Expense) update(c *gin.Context) {
	id, ok := expenseID(c)
	if !ok {
		return
	}
	var req updateExpenseRequest
	if !bind(c, &req) {
		return
	}
	expense, err := h.service.Update(c.Request.Context(), c.GetString(middleware.UserIDKey), id, service.UpdateExpenseInput{
		Amount: req.Amount, Note: req.Note, Category: req.Category, SpentAt: req.SpentAt, ImageKey: req.ImageKey,
	})
	if err != nil {
		expenseError(c, err)
		return
	}
	view, err := h.expenseView(c, expense)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *Expense) delete(c *gin.Context) {
	id, ok := expenseID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), c.GetString(middleware.UserIDKey), id); err != nil {
		expenseError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func expenseID(c *gin.Context) (string, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apierror.Write(c, http.StatusBadRequest, "invalid_expense_id", "Expense ID must be a UUID")
		return "", false
	}
	return id.String(), true
}

func (h *Expense) expenseView(c *gin.Context, expense model.Expense) (expenseResponse, error) {
	imageURL, err := h.service.SignedImageURL(c.Request.Context(), expense.ImageKey)
	if err != nil {
		return expenseResponse{}, err
	}
	return expenseResponse{
		ID: expense.ID, Amount: expense.Amount, Note: expense.Note, Category: expense.Category,
		ImageURL: imageURL, SpentAt: expense.SpentAt, CreatedAt: expense.CreatedAt, UpdatedAt: expense.UpdatedAt,
	}, nil
}

func expenseError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput), errors.Is(err, service.ErrInvalidMonth):
		apierror.Write(c, http.StatusBadRequest, "invalid_request", "Invalid expense fields or month; month must use YYYY-MM")
	case errors.Is(err, service.ErrInvalidTimezone):
		apierror.Write(c, http.StatusBadRequest, "invalid_timezone", "Query tz must be a valid IANA time zone")
	case errors.Is(err, service.ErrInvalidImage), errors.Is(err, service.ErrInvalidImageKey):
		apierror.Write(c, http.StatusBadRequest, "invalid_image", "Image upload is invalid or unavailable")
	case errors.Is(err, service.ErrImageStorageUnavailable):
		apierror.Write(c, http.StatusServiceUnavailable, "image_storage_unavailable", "Image storage is not configured")
	case errors.Is(err, model.ErrExpenseNotFound):
		apierror.Write(c, http.StatusNotFound, "expense_not_found", "Expense not found")
	default:
		respondError(c, err)
	}
}
