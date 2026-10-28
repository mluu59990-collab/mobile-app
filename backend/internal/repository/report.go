package repository

import (
	"context"
	"database/sql"
	"time"

	"expense-app/backend/internal/model"
	"gorm.io/gorm"
)

type Reports struct{ db *gorm.DB }

func NewReports(db *gorm.DB) *Reports { return &Reports{db: db} }

func (r *Reports) MonthlyExpenseReport(ctx context.Context, userID string, start, end time.Time) (model.MonthlyExpenseReport, error) {
	var report model.MonthlyExpenseReport
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		baseQuery := func() *gorm.DB {
			return tx.Table("expenses").Where("user_id = ? AND spent_at >= ? AND spent_at < ?", userID, start, end)
		}
		if err := baseQuery().Select("COALESCE(SUM(amount), 0)").Scan(&report.TotalAmount).Error; err != nil {
			return err
		}
		if err := baseQuery().Select("category, SUM(amount) AS total_amount").Group("category").Order("category ASC NULLS LAST").Scan(&report.ByCategory).Error; err != nil {
			return err
		}
		return baseQuery().Select("to_char(spent_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS date, SUM(amount) AS total_amount").
			Group("to_char(spent_at AT TIME ZONE 'UTC', 'YYYY-MM-DD')").Order("date ASC").Scan(&report.ByDay).Error
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	return report, err
}
