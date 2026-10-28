package model

import "time"

type User struct {
	ID           string `gorm:"type:uuid;primaryKey"`
	Email        *string
	PasswordHash *string `json:"-"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type AuthIdentity struct {
	ID             string `gorm:"type:uuid;primaryKey"`
	UserID         string
	Provider       string
	ProviderUserID string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type RefreshToken struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	UserID    string
	TokenHash string `json:"-"`
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}
