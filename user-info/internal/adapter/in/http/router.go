// @title Enduran User Info API
// @version 1.0
// @description Сервис информации о пользователе (вес, рост, возраст и т.д.)
// @BasePath /api/v1

package httpin

import (
	"github.com/gin-gonic/gin"

	_ "github.com/EnduranNSU/end-user-info/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewGinRouter создает новый Gin router
// @title Enduran User Info API
// @version 1.0
// @description Сервис информации о пользователе (вес, рост, возраст и т.д.)
// @BasePath /api/v1
func NewGinRouter(h *UserInfoHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.StaticFile("/openapi.yaml", "docs/swagger.yaml")

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		// CreateUserInfo godoc
		// @Summary Создать пользовательскую информацию
		// @Description Создает новую запись с информацией о пользователе (вес, рост, возраст)
		// @Tags user-info
		// @Accept json
		// @Produce json
		// @Param request body dto.CreateUserInfoRequest true "Данные пользователя"
		// @Success 200 {object} dto.UserInfoResponse
		// @Failure 400 {object} dto.ErrorResponse
		// @Failure 500 {object} dto.ErrorResponse
		// @Router /user-info [post]
		api.POST("/user-info", h.Create)

		// GetLatestUserInfo godoc
		// @Summary Получить последнюю запись
		// @Description Возвращает последнюю запись информации о пользователе
		// @Tags user-info
		// @Produce json
		// @Param user_id query string true "User ID"
		// @Success 200 {object} dto.UserInfoResponse
		// @Failure 400 {object} dto.ErrorResponse
		// @Failure 404 {object} dto.ErrorResponse
		// @Router /user-info/latest [get]
		api.GET("/user-info/latest", h.GetLatest)

		// ListUserInfo godoc
		// @Summary Получить все записи
		// @Description Возвращает все записи информации о пользователе
		// @Tags user-info
		// @Produce json
		// @Param user_id query string true "User ID"
		// @Success 200 {array} dto.UserInfoResponse
		// @Failure 400 {object} dto.ErrorResponse
		// @Failure 404 {object} dto.ErrorResponse
		// @Router /user-info [get]
		api.GET("/user-info", h.List)
	}

	return r
}