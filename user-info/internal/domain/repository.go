package domain

import (
	"context"

	"github.com/google/uuid"
)

type UserInfoRepository interface {
	CreateUserInfo(ctx context.Context, info *UserInfo) error
	GetLatestUserInfoByUserID(ctx context.Context, userID uuid.UUID) (*UserInfo, error)
	GetAllUserInfoByUserID(ctx context.Context, userID uuid.UUID) ([]*UserInfo, error)
}
