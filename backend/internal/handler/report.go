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

type Report struct{ service *service.MonthlyReportService }

type monthlyReportQuery struct {
	Month    string `form:"month" binding:"required"`
	Timezone string `form:"tz"`
}

type categoryReportItem struct {
	Category    *string `json:"category"`
	TotalAmount int64   `json:"total_amount"`
}

type dailyReportItem struct {
	Date        string `json:"date"`
	TotalAmount int64  `json:"total_amount"`
}

type monthlyReportResponse struct {
	Month       string               `json:"month"`
	TotalAmount int64                `json:"total_amount"`
	ByCategory  []categoryReportItem `json:"by_category"`
	ByDay       []dailyReportItem    `json:"by_day"`
}

func RegisterReportRoutes(router *gin.Engine, svc *service.MonthlyReportService, tokens auth.Tokens) {
	h := &Report{service: svc}
	group := router.Group("/api/v1/reports")
	group.Use(middleware.JWT(tokens))
	group.GET("/monthly", h.monthly)
}

func (h *Report) monthly(c *gin.Context) {
	var query monthlyReportQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		apierror.Write(c, http.StatusBadRequest, "invalid_request", "Query month is required in YYYY-MM format")
		return
	}
	report, err := h.service.Monthly(c.Request.Context(), c.GetString(middleware.UserIDKey), query.Month, query.Timezone)
	if errors.Is(err, service.ErrInvalidMonth) {
		apierror.Write(c, http.StatusBadRequest, "invalid_request", "Query month must be a valid YYYY-MM month")
		return
	}
	if errors.Is(err, service.ErrInvalidTimezone) {
		apierror.Write(c, http.StatusBadRequest, "invalid_timezone", "Query tz must be a valid IANA time zone")
		return
	}
	if err != nil {
		respondError(c, err)
		return
	}
	response := monthlyReportResponse{
		Month: query.Month, TotalAmount: report.TotalAmount,
		ByCategory: make([]categoryReportItem, 0, len(report.ByCategory)),
		ByDay:      make([]dailyReportItem, 0, len(report.ByDay)),
	}
	for _, item := range report.ByCategory {
		response.ByCategory = append(response.ByCategory, categoryReportItem{Category: item.Category, TotalAmount: item.TotalAmount})
	}
	for _, item := range report.ByDay {
		response.ByDay = append(response.ByDay, dailyReportItem{Date: item.Date, TotalAmount: item.TotalAmount})
	}
	c.JSON(http.StatusOK, response)
}
