package httpin

import (
	"net/http"
	"strings"

	"auth/internal/adapter/in/http/dto"
	"auth/internal/domain"
	"auth/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *service.Service
}

func NewAuthHandler(svc *service.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func bearer(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if len(h) < 7 {
		return ""
	}
	if strings.ToLower(h[:7]) != "bearer " {
		return ""
	}
	return strings.TrimSpace(h[7:])
}

// Register регистрирует нового пользователя
// @Summary      Регистрация пользователя
// @Description  Создаёт нового пользователя и возвращает пару access/refresh токенов
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RegisterRequest  true  "Учётные данные пользователя"
// @Success      201      {object}  dto.TokenResponse
// @Failure      400      {object}  dto.ErrorResponse   "Неверный формат запроса"
// @Failure      409      {object}  dto.ErrorResponse   "Пользователь с таким email уже существует"
// @Failure      500      {object}  dto.ErrorResponse   "Внутренняя ошибка сервера"
// @Router       /register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request"})
		return
	}

	tp, err := h.svc.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == domain.ErrAlreadyExists {
			c.AbortWithStatusJSON(http.StatusConflict, dto.ErrorResponse{Error: "email_exists"})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal"})
		return
	}

	c.JSON(http.StatusCreated, dto.TokenResponse{
		AccessToken:  tp.AccessToken,
		RefreshToken: tp.RefreshToken,
	})
}

// Login аутентифицирует пользователя по email и паролю
// @Summary      Логин пользователя
// @Description  Проверяет email/пароль и возвращает пару access/refresh токенов
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.LoginRequest  true  "Учётные данные пользователя"
// @Success      200      {object}  dto.TokenResponse
// @Failure      400      {object}  dto.ErrorResponse   "Неверный формат запроса"
// @Failure      401      {object}  dto.ErrorResponse   "Неверные учётные данные"
// @Failure      403      {object}  dto.ErrorResponse   "Пользователь заблокирован"
// @Failure      500      {object}  dto.ErrorResponse   "Внутренняя ошибка сервера"
// @Router       /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request"})
		return
	}

	tp, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case domain.ErrInvalidCreds:
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid_credentials"})
		case domain.ErrBlockedUser:
			c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{Error: "blocked"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  tp.AccessToken,
		RefreshToken: tp.RefreshToken,
	})
}

// Refresh обновляет пару токенов по refresh-токену
// @Summary      Обновление токенов
// @Description  Принимает refresh-токен и возвращает новую пару access/refresh (rotation)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RefreshRequest  true  "Refresh токен"
// @Success      200      {object}  dto.TokenResponse
// @Failure      400      {object}  dto.ErrorResponse   "Неверный формат запроса"
// @Failure      401      {object}  dto.ErrorResponse   "Невалидный или просроченный refresh-токен"
// @Failure      500      {object}  dto.ErrorResponse   "Внутренняя ошибка сервера"
// @Router       /refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request"})
		return
	}

	tp, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid_refresh"})
		return
	}

	c.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  tp.AccessToken,
		RefreshToken: tp.RefreshToken,
	})
}

// Logout инвалидирует один refresh-токен (выход с текущего устройства)
// @Summary      Логаут с одного устройства
// @Description  Помечает переданный refresh-токен как отозванный. Идемпотентен.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RefreshRequest  false  "Refresh токен (опционально)"
// @Success      204      {string}  string              "Успешный логаут, тело отсутствует"
// @Failure      400      {object}  dto.ErrorResponse   "Неверный формат запроса"
// @Router       /logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.RefreshRequest
	_ = c.ShouldBindJSON(&req)
	_ = h.svc.Logout(c.Request.Context(), req.RefreshToken)
	c.Status(http.StatusNoContent)
}

// StartReset начинает процесс сброса пароля (отправка OTP-кода)
// @Summary      Начало сброса пароля
// @Description  Генерирует одноразовый код для сброса пароля и сохраняет его. В dev-режиме возвращает код в ответе.
// @Tags         password-reset
// @Accept       json
// @Produce      json
// @Param        request  body      dto.StartResetRequest       true  "Email пользователя"
// @Success      200      {object}  dto.StartResetDevResponse   "Dev-режим: OTP-код в ответе"
// @Success      204      {string}  string                      "В проде: всегда 204, даже если email не найден"
// @Failure      400      {object}  dto.ErrorResponse           "Неверный формат запроса"
// @Router       /password/reset/start [post]
func (h *AuthHandler) StartReset(c *gin.Context) {
	var req dto.StartResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request"})
		return
	}

	code, err := h.svc.StartPasswordResetOTP(c.Request.Context(), req.Email)
	if err != nil {
		c.Status(http.StatusNoContent)
		return
	}
	// lol, for testing purposes
	c.JSON(http.StatusOK, gin.H{"dev_code": code})
}

// ConfirmReset подтверждает сброс пароля по OTP-коду
// @Summary      Подтверждение сброса пароля
// @Description  Проверяет OTP-код, устанавливает новый пароль и инвалидирует все refresh-токены пользователя
// @Tags         password-reset
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ConfirmResetRequest  true  "Email, OTP-код и новый пароль"
// @Success      204      {string}  string                   "Пароль успешно изменён, тело отсутствует"
// @Failure      400      {object}  dto.ErrorResponse        "Неверный код или некорректные данные"
// @Router       /password/reset/confirm [post]
func (h *AuthHandler) ConfirmReset(c *gin.Context) {
	var req dto.ConfirmResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad_request"})
		return
	}

	if err := h.svc.ConfirmPasswordResetOTP(c.Request.Context(), req.Email, req.Code, req.NewPassword); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_code"})
		return
	}
	c.Status(http.StatusNoContent)
}

// Validate проверяет валидность access-токена
// @Summary      Валидация access-токена
// @Description  Проверяет access-токен, убеждается что пользователь существует и не заблокирован, и возвращает его ID.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer access токен"  default(Bearer <token>)
// @Success      200            {object}  dto.ValidateResponse
// @Failure      401            {object}  dto.ErrorResponse  "Нет токена или он невалиден"
// @Security     BearerAuth
// @Router       /validate [get]
func (h *AuthHandler) Validate(c *gin.Context) {
	access := bearer(c)
	if access == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "no_bearer"})
		return
	}

	userID, err := h.svc.ValidateAccess(c.Request.Context(), access)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid_token"})
		return
	}

	c.JSON(http.StatusOK, dto.ValidateResponse{
		UserID: userID.String(),
	})
}
