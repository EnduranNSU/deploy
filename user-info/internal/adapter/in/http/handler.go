package httpin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/EnduranNSU/end-user-info/internal/adapter/in/http/dto"
	svcuserinfo "github.com/EnduranNSU/end-user-info/internal/domain"
)

type UserInfoHandler struct {
	svc svcuserinfo.Service
}

func NewUserInfoHandler(svc svcuserinfo.Service) *UserInfoHandler {
	return &UserInfoHandler{svc: svc}
}

// Create создает новую запись пользовательской информации
// @Summary      Создать пользовательскую информацию
// @Description  Создает новую запись с информацией о пользователе (вес, рост, возраст)
// @Tags         user-info
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateUserInfoRequest true "Данные пользователя"
// @Success      200  {object}  dto.UserInfoResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /user-info [post]
func (h *UserInfoHandler) Create(c *gin.Context) {
	var req dto.CreateUserInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "bad json"})
		return
	}

	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid user_id"})
		return
	}

	cmd := svcuserinfo.CreateUserInfoCmd{
		UserID: uid,
		Weight: req.Weight,
		Height: req.Height,
		Age:    req.Age,
		Date:   req.Date,
	}

	m, err := h.svc.Create(c.Request.Context(), cmd)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "failed to create"})
		return
	}

	c.JSON(http.StatusOK, dto.UserInfoResponse{
		Weight: m.Weight,
		Height: m.Height,
		Age:    m.Age,
		Date:   m.Date.Format("2006-01-02"),
	})
}

// GetLatest получает последнюю запись пользовательской информации
// @Summary      Получить последнюю запись
// @Description  Возвращает последнюю запись информации о пользователе
// @Tags         user-info
// @Produce      json
// @Param        user_id query string true "User ID"
// @Success      200  {object}  dto.UserInfoResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Router       /user-info/latest [get]
func (h *UserInfoHandler) GetLatest(c *gin.Context) {
	uidStr := c.Query("user_id")
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid user_id"})
		return
	}
	m, err := h.svc.GetLatest(c.Request.Context(), uid)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{Error: "not found"})
		return
	}
	c.JSON(http.StatusOK, dto.UserInfoResponse{
		Weight: m.Weight,
		Height: m.Height,
		Age:    m.Age,
		Date:   m.Date.Format("2006-01-02"),
	})
}

// List получает все записи пользовательской информации
// @Summary      Получить все записи
// @Description  Возвращает все записи информации о пользователе
// @Tags         user-info
// @Produce      json
// @Param        user_id query string true "User ID"
// @Success      200  {array}   dto.UserInfoResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Router       /user-info [get]
func (h *UserInfoHandler) List(c *gin.Context) {
	uidStr := c.Query("user_id")
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid user_id"})
		return
	}
	items, err := h.svc.List(c.Request.Context(), uid)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{Error: "not found"})
		return
	}
	resp := make([]dto.UserInfoResponse, 0, len(items))
	for _, m := range items {
		resp = append(resp, dto.UserInfoResponse{
			Weight: m.Weight,
			Height: m.Height,
			Age:    m.Age,
			Date:   m.Date.Format("2006-01-02"),
		})
	}
	c.JSON(http.StatusOK, resp)
}
