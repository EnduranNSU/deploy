package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/EnduranNSU/end-user-info/internal/domain"
	"github.com/google/uuid"
)

var ErrInvalidDate = errors.New("invalid date format, want YYYY-MM-DD")

func New(repo domain.UserInfoRepository) domain.Service {
	return &service{repo: repo}
}

type service struct {
	repo domain.UserInfoRepository
}

func (s *service) Create(ctx context.Context, cmd domain.CreateUserInfoCmd) (*domain.UserInfo, error) {
	var dt time.Time
	dateStr := strings.TrimSpace(cmd.Date)

	if dateStr == "" {
		dt = time.Now().UTC()
	} else {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, ErrInvalidDate
		}
		dt = parsed
	}

	m := &domain.UserInfo{
		UserID: cmd.UserID,
		Weight: cmd.Weight,
		Height: cmd.Height,
		Age:    cmd.Age,
		Date:   dt,
	}

	if err := s.repo.CreateUserInfo(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *service) GetLatest(ctx context.Context, userID uuid.UUID) (*domain.UserInfo, error) {
	return s.repo.GetLatestUserInfoByUserID(ctx, userID)
}

func (s *service) List(ctx context.Context, userID uuid.UUID) ([]*domain.UserInfo, error) {
	return s.repo.GetAllUserInfoByUserID(ctx, userID)
}
