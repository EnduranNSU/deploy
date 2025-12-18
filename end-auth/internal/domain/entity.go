package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	IsBlocked    bool
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RefreshSession struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash []byte
	UserAgent *string
	IP        *string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

type PasswordReset struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	OTPHash   string
	ExpiresAt time.Time
	UsedAt    *time.Time
}
