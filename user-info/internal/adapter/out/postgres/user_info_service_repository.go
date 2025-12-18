package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/EnduranNSU/end-user-info/internal/adapter/out/postgres/gen"
	"github.com/EnduranNSU/end-user-info/internal/domain"
	"github.com/EnduranNSU/end-user-info/internal/logging"
)

type UserInfoRepositoryImpl struct {
	db      *sql.DB
	queries *gen.Queries
}

func NewUserInfoRepository(db *sql.DB) domain.UserInfoRepository {
	return &UserInfoRepositoryImpl{
		db:      db,
		queries: gen.New(db),
	}
}

func (r *UserInfoRepositoryImpl) CreateUserInfo(ctx context.Context, info *domain.UserInfo) error {
	params := gen.CreateUserInfoParams{
		Weight: info.Weight.String(),
		Height: info.Height,
		Date:   info.Date,
		Age:    info.Age,
		UserID: info.UserID,
	}

	err := r.queries.CreateUserInfo(ctx, params)
	if err != nil {
		jsonData := logging.MarshalLogData(info)
		logging.Error(err, "CreateUserInfo", jsonData, "failed to create user info")
		return err
	}

	jsonData := logging.MarshalLogData(info)
	logging.Debug("CreateUserInfo", jsonData, "successfully created user info")

	return nil
}

func (r *UserInfoRepositoryImpl) GetLatestUserInfoByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserInfo, error) {
	info, err := r.queries.GetLatestUserInfoByUserID(ctx, userID)
	if err == sql.ErrNoRows {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"user_id": userID.String(),
		})
		logging.Warn("GetLatestUserInfoByUserID", jsonData, "user info not found")
		return nil, fmt.Errorf("user_info not found for user_id: %s", userID)
	}

	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"user_id": userID.String(),
		})
		logging.Error(err, "GetLatestUserInfoByUserID", jsonData, "failed to get latest user info")
		return nil, err
	}

	weight, _ := decimal.NewFromString(info.Weight)
	domainInfo := &domain.UserInfo{
		ID:     info.ID,
		Weight: weight,
		Height: info.Height,
		Date:   info.Date,
		Age:    info.Age,
		UserID: info.UserID,
	}

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"user_id": userID.String(),
		"info_id": info.ID,
		"date":    info.Date,
	})
	logging.Debug("GetLatestUserInfoByUserID", jsonData, "successfully retrieved latest user info")

	return domainInfo, nil
}

func (r *UserInfoRepositoryImpl) GetAllUserInfoByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.UserInfo, error) {
	infos, err := r.queries.GetAllUserInfoByUserID(ctx, userID)
	if err != nil {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"user_id": userID.String(),
		})
		logging.Error(err, "GetAllUserInfoByUserID", jsonData, "failed to query user info")
		return nil, err
	}

	if len(infos) == 0 {
		jsonData := logging.MarshalLogData(map[string]interface{}{
			"user_id": userID.String(),
		})
		logging.Warn("GetAllUserInfoByUserID", jsonData, "no user info records found")
		return nil, fmt.Errorf("no user_info records found for user_id: %s", userID)
	}

	domainInfos := make([]*domain.UserInfo, len(infos))
	for i, info := range infos {
		weight, _ := decimal.NewFromString(info.Weight)
		domainInfos[i] = &domain.UserInfo{
			ID:     info.ID,
			Weight: weight,
			Height: info.Height,
			Date:   info.Date,
			Age:    info.Age,
			UserID: info.UserID,
		}
	}

	jsonData := logging.MarshalLogData(map[string]interface{}{
		"user_id":       userID.String(),
		"records_count": len(domainInfos),
	})
	logging.Debug("GetAllUserInfoByUserID", jsonData, "successfully retrieved user info records")

	return domainInfos, nil
}