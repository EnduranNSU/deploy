package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type UserInfo struct {
	ID     int64           `db:"id" json:"id"`
	Weight decimal.Decimal `db:"weight" json:"weight"`
	Height int32           `db:"height" json:"height"`
	Date   time.Time       `db:"date" json:"date"`
	Age    int32           `db:"age" json:"age"`
	UserID uuid.UUID       `db:"user_id" json:"user_id"`
}
