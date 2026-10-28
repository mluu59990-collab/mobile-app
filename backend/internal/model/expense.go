package model

import "time"

type Expense struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	UserID    string `gorm:"type:uuid;not null"`
	Amount    int64
	Note      *string
	Category  *string
	SpentAt   time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Expense) TableName() string { return "expenses" }
