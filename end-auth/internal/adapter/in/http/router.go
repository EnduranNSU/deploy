// @title Training API
// @version 1.0
// @description Сервис авторизации
// @BasePath /api/v1
package httpin

import (
	_ "auth/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewGinRouter создает новый Gin router
// @title Enduran Training API
// @version 1.0
// @description Сервис авторизации
// @BasePath /api/v1
func NewGinRouter(h *AuthHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	a := r.Group("/api/v1")
	{
		a.POST("/register", h.Register)
		a.POST("/login", h.Login)
		a.POST("/refresh", h.Refresh)
		a.POST("/logout", h.Logout)

		pr := a.Group("/password/reset")
		{
			pr.POST("/start", h.StartReset)
			pr.POST("/confirm", h.ConfirmReset)
		}

		a.GET("/validate", h.Validate)
	}

	return r
}
