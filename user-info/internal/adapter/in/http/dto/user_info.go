package dto

import "github.com/shopspring/decimal"

// CreateUserInfoRequest представляет запрос на создание пользовательской информации
type CreateUserInfoRequest struct {
	Weight decimal.Decimal `json:"weight" binding:"required" example:"70.5" minimum:"0.1" maximum:"300" description:"Вес пользователя в килограммах"`
	Height int32           `json:"height" binding:"required,gt=0,lte=300" example:"175" minimum:"1" maximum:"300" description:"Рост пользователя в сантиметрах"`
	Date   string          `json:"date,omitempty" example:"2023-10-05" pattern:"^\\d{4}-\\d{2}-\\d{2}$" description:"Дата в формате YYYY-MM-DD (опционально, по умолчанию текущая дата)"`
	Age    int32           `json:"age" binding:"required,gt=0,lte=150" example:"25" minimum:"1" maximum:"150" description:"Возраст пользователя в годах"`
}

// UserInfoResponse представляет ответ с пользовательской информацией
type UserInfoResponse struct {
	Weight decimal.Decimal `json:"weight" example:"70.5" description:"Вес пользователя в килограммах"`
	Height int32           `json:"height" example:"175" description:"Рост пользователя в сантиметрах"`
	Date   string          `json:"date" example:"2023-10-05" description:"Дата в формате YYYY-MM-DD"`
	Age    int32           `json:"age" example:"25" description:"Возраст пользователя в годах"`
}

// ErrorResponse представляет ответ об ошибке
type ErrorResponse struct {
	Error string `json:"error" example:"error message" description:"Описание ошибки"`
}
