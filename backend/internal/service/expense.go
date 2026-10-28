package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"expense-app/backend/internal/model"
	"github.com/google/uuid"
)

var ErrInvalidMonth = errors.New("invalid month filter")

type ExpenseRepository interface {
	CreateExpense(context.Context, *model.Expense) error
	ListExpenses(context.Context, string, *time.Time, *time.Time) ([]model.Expense, error)
	GetExpense(context.Context, string, string) (model.Expense, error)
	UpdateExpense(context.Context, string, string, map[string]any) (model.Expense, error)
	DeleteExpense(context.Context, string, string) error
}

type ExpenseService struct{ repo ExpenseRepository }

type CreateExpenseInput struct {
	Amount   int64
	Note     *string
	Category *string
	SpentAt  *time.Time
}

type UpdateExpenseInput struct {
	Amount   *int64
	Note     *string
	Category *string
	SpentAt  *time.Time
}

func NewExpenseService(repo ExpenseRepository) *ExpenseService {
	return &ExpenseService{repo: repo}
}

func (s *ExpenseService) Create(ctx context.Context, userID string, input CreateExpenseInput) (model.Expense, error) {
	if input.Amount <= 0 {
		return model.Expense{}, ErrInvalidInput
	}
	spentAt := time.Now().UTC()
	if input.SpentAt != nil {
		spentAt = input.SpentAt.UTC()
	}
	expense := model.Expense{
		ID: uuid.NewString(), UserID: userID, Amount: input.Amount,
		Note: normalizeText(input.Note), Category: normalizeText(input.Category), SpentAt: spentAt,
	}
	if err := s.repo.CreateExpense(ctx, &expense); err != nil {
		return model.Expense{}, err
	}
	return expense, nil
}

func (s *ExpenseService) List(ctx context.Context, userID, month string) ([]model.Expense, error) {
	var start, end *time.Time
	if month != "" {
		monthStart, monthEnd, err := monthRange(month)
		if err != nil {
			return nil, ErrInvalidMonth
		}
		start, end = &monthStart, &monthEnd
	}
	return s.repo.ListExpenses(ctx, userID, start, end)
}

func (s *ExpenseService) Get(ctx context.Context, userID, id string) (model.Expense, error) {
	return s.repo.GetExpense(ctx, userID, id)
}

func (s *ExpenseService) Update(ctx context.Context, userID, id string, input UpdateExpenseInput) (model.Expense, error) {
	updates := make(map[string]any, 4)
	if input.Amount != nil {
		if *input.Amount <= 0 {
			return model.Expense{}, ErrInvalidInput
		}
		updates["amount"] = *input.Amount
	}
	if input.Note != nil {
		updates["note"] = normalizeText(input.Note)
	}
	if input.Category != nil {
		updates["category"] = normalizeText(input.Category)
	}
	if input.SpentAt != nil {
		updates["spent_at"] = input.SpentAt.UTC()
	}
	if len(updates) == 0 {
		return model.Expense{}, ErrInvalidInput
	}
	return s.repo.UpdateExpense(ctx, userID, id, updates)
}

func (s *ExpenseService) Delete(ctx context.Context, userID, id string) error {
	return s.repo.DeleteExpense(ctx, userID, id)
}

func normalizeText(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
