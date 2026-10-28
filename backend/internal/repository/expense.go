package repository

import (
	"context"
	"errors"
	"time"

	"expense-app/backend/internal/model"
	"gorm.io/gorm"
)

type Expenses struct{ db *gorm.DB }

func NewExpenses(db *gorm.DB) *Expenses { return &Expenses{db: db} }

func (r *Expenses) CreateExpense(ctx context.Context, expense *model.Expense) error {
	return r.db.WithContext(ctx).Create(expense).Error
}

func (r *Expenses) ListExpenses(ctx context.Context, userID string, start, end *time.Time) ([]model.Expense, error) {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if start != nil && end != nil {
		query = query.Where("spent_at >= ? AND spent_at < ?", *start, *end)
	}
	var expenses []model.Expense
	err := query.Order("spent_at DESC").Order("id DESC").Find(&expenses).Error
	return expenses, err
}

func (r *Expenses) GetExpense(ctx context.Context, userID, id string) (model.Expense, error) {
	var expense model.Expense
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Take(&expense).Error
	if err != nil {
		return model.Expense{}, expenseNotFound(err)
	}
	return expense, nil
}

func (r *Expenses) UpdateExpense(ctx context.Context, userID, id string, updates map[string]any) (model.Expense, error) {
	result := r.db.WithContext(ctx).Model(&model.Expense{}).Where("id = ? AND user_id = ?", id, userID).Updates(updates)
	if result.Error != nil {
		return model.Expense{}, result.Error
	}
	if result.RowsAffected == 0 {
		return model.Expense{}, model.ErrExpenseNotFound
	}
	return r.GetExpense(ctx, userID, id)
}

func (r *Expenses) DeleteExpense(ctx context.Context, userID, id string) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.Expense{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return model.ErrExpenseNotFound
	}
	return nil
}

func expenseNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.ErrExpenseNotFound
	}
	return err
}
