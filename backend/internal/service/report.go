package service

import (
	"context"
	"time"

	"expense-app/backend/internal/model"
)

type ReportRepository interface {
	MonthlyExpenseReport(context.Context, string, time.Time, time.Time, string) (model.MonthlyExpenseReport, error)
}

type MonthlyReportService struct{ repo ReportRepository }

func NewMonthlyReportService(repo ReportRepository) *MonthlyReportService {
	return &MonthlyReportService{repo: repo}
}

func (s *MonthlyReportService) Monthly(ctx context.Context, userID, month, timezone string) (model.MonthlyExpenseReport, error) {
	start, end, location, err := monthRange(month, timezone)
	if err != nil {
		return model.MonthlyExpenseReport{}, err
	}
	report, err := s.repo.MonthlyExpenseReport(ctx, userID, start, end, location.String())
	if err != nil {
		return model.MonthlyExpenseReport{}, err
	}
	if report.ByCategory == nil {
		report.ByCategory = []model.CategoryExpenseTotal{}
	}
	if report.ByDay == nil {
		report.ByDay = []model.DailyExpenseTotal{}
	}
	return report, nil
}
