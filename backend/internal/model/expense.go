package model

import (
	"encoding/json"
	"time"
)

type Expense struct {
	ID          string `gorm:"type:uuid;primaryKey"`
	UserID      string `gorm:"type:uuid;not null"`
	Amount      int64
	Note        *string
	Category    *string
	ImageKey    *string
	ImageURL    *string         `gorm:"-"`
	Source      string          `gorm:"not null;default:manual"`
	RawAIResult json.RawMessage `gorm:"type:jsonb;serializer:json"`
	SpentAt     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Expense) TableName() string { return "expenses" }
