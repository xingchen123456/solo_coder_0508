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
	roleController := controllers.NewRoleController()
	permissionController := controllers.NewPermissionController()

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
			users.PUT("/:id/roles", userController.AssignRoles)
			users.GET("/:id/roles", userController.GetUserRoles)
			users.GET("/:id/permissions", userController.GetUserPermissions)
		}

		roles := api.Group("/roles")
		{
			roles.POST("", roleController.Create)
			roles.GET("", roleController.List)
			roles.GET("/:id", roleController.Get)
			roles.PUT("/:id", roleController.Update)
			roles.DELETE("/:id", roleController.Delete)
			roles.PUT("/:id/permissions", roleController.AssignPermissions)
			roles.GET("/:id/permissions", roleController.GetWithPermissions)
		}

		permissions := api.Group("/permissions")
		{
			permissions.POST("", permissionController.Create)
			permissions.GET("", permissionController.List)
			permissions.GET("/:id", permissionController.Get)
			permissions.PUT("/:id", permissionController.Update)
			permissions.DELETE("/:id", permissionController.Delete)
		}
	}
}
