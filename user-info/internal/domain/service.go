package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Service interface {
	Create(ctx context.Context, cmd CreateUserInfoCmd) (*UserInfo, error)
	GetLatest(ctx context.Context, userID uuid.UUID) (*UserInfo, error)
	List(ctx context.Context, userID uuid.UUID) ([]*UserInfo, error)
}

type CreateUserInfoCmd struct {
	UserID uuid.UUID
	Weight decimal.Decimal
	Height int32
	Age    int32
	Date   string
}
