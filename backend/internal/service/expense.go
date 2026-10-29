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

type ExpenseService struct {
	repo        ExpenseRepository
	images      ImageStorage
	imageConfig ExpenseImageConfig
}

type CreateExpenseInput struct {
	Amount   int64
	Note     *string
	Category *string
	SpentAt  *time.Time
	ImageKey *string
}

type UpdateExpenseInput struct {
	Amount   *int64
	Note     *string
	Category *string
	SpentAt  *time.Time
	ImageKey *string
}

func NewExpenseService(repo ExpenseRepository) *ExpenseService {
	return &ExpenseService{repo: repo, imageConfig: ExpenseImageConfig{PutURLTTL: 10 * time.Minute, GetURLTTL: 15 * time.Minute}}
}

func NewExpenseServiceWithImages(repo ExpenseRepository, images ImageStorage, imageConfig ExpenseImageConfig) *ExpenseService {
	return &ExpenseService{repo: repo, images: images, imageConfig: imageConfig}
}

func (s *ExpenseService) Create(ctx context.Context, userID string, input CreateExpenseInput) (model.Expense, error) {
	if input.Amount <= 0 {
		return model.Expense{}, ErrInvalidInput
	}
	spentAt := time.Now().UTC()
	if input.SpentAt != nil {
		spentAt = input.SpentAt.UTC()
	}
	var imageKey *string
	if input.ImageKey != nil {
		finalized, err := s.FinalizeImage(ctx, userID, *input.ImageKey)
		if err != nil {
			return model.Expense{}, err
		}
		imageKey = &finalized
	}
	expense := model.Expense{
		ID: uuid.NewString(), UserID: userID, Amount: input.Amount,
		Note: normalizeText(input.Note), Category: normalizeText(input.Category), SpentAt: spentAt,
		ImageKey: imageKey,
	}
	if err := s.repo.CreateExpense(ctx, &expense); err != nil {
		if imageKey != nil {
			s.cleanupImage(ctx, "remove_unattached_image", *imageKey)
		}
		return model.Expense{}, err
	}
	return expense, nil
}

func (s *ExpenseService) List(ctx context.Context, userID, month, timezone string) ([]model.Expense, error) {
	location, err := timezoneLocation(timezone)
	if err != nil {
		return nil, err
	}
	var start, end *time.Time
	if month != "" {
		monthStart, monthEnd, _, err := monthRange(month, location.String())
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
	var oldImageKey *string
	var newImageKey *string
	if input.ImageKey != nil {
		oldExpense, err := s.repo.GetExpense(ctx, userID, id)
		if err != nil {
			return model.Expense{}, err
		}
		oldImageKey = oldExpense.ImageKey
		finalized, err := s.FinalizeImage(ctx, userID, *input.ImageKey)
		if err != nil {
			return model.Expense{}, err
		}
		newImageKey = &finalized
		updates["image_key"] = finalized
	}
	if len(updates) == 0 {
		return model.Expense{}, ErrInvalidInput
	}
	updated, err := s.repo.UpdateExpense(ctx, userID, id, updates)
	if err != nil {
		if newImageKey != nil {
			s.cleanupImage(ctx, "remove_unattached_image", *newImageKey)
		}
		return model.Expense{}, err
	}
	if oldImageKey != nil && (newImageKey == nil || *oldImageKey != *newImageKey) {
		s.cleanupImage(ctx, "remove_replaced_image", *oldImageKey)
	}
	return updated, nil
}

func (s *ExpenseService) Delete(ctx context.Context, userID, id string) error {
	expense, err := s.repo.GetExpense(ctx, userID, id)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteExpense(ctx, userID, id); err != nil {
		return err
	}
	if expense.ImageKey != nil {
		s.cleanupImage(ctx, "remove_deleted_expense_image", *expense.ImageKey)
	}
	return nil
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
