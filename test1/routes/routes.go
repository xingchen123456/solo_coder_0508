package routes

import (
	"management-system/controllers"
	"management-system/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupRoutes(r *gin.Engine, logger *zap.Logger) {
	r.Use(middleware.LogMiddleware(logger))
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.CORSMiddleware())

	healthController := controllers.NewHealthController()
	userController := controllers.NewUserController()

	api := r.Group("/api")
	{
		api.GET("/health", healthController.Check)

		users := api.Group("/users")
		{
			users.POST("", userController.Create)
			users.GET("", userController.List)
			users.GET("/:id", userController.Get)
			users.PUT("/:id", userController.Update)
			users.DELETE("/:id", userController.Delete)
		}
	}
}
