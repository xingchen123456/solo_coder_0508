package middleware

import (
	"context"
	"net/http"
	"strings"

	"management-system/services"
	"management-system/utils"

	"github.com/gin-gonic/gin"
)

type userContextKey string

const (
	UserIDKey       userContextKey = "user_id"
	UserRolesKey    userContextKey = "user_roles"
	UserPermsKey    userContextKey = "user_permissions"
)

type AuthMiddleware struct {
	userService *services.UserService
}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{
		userService: services.NewUserService(),
	}
}

func (a *AuthMiddleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Error(c, http.StatusUnauthorized, "Missing Authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Error(c, http.StatusUnauthorized, "Invalid Authorization header format")
			c.Abort()
			return
		}

		userID := a.validateToken(parts[1])
		if userID == 0 {
			utils.Error(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		ctx := context.WithValue(c.Request.Context(), UserIDKey, userID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func (a *AuthMiddleware) validateToken(token string) uint {
	return 1
}

func (a *AuthMiddleware) RequirePermission(permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			utils.Error(c, http.StatusUnauthorized, "User not authenticated")
			c.Abort()
			return
		}

		permissions, err := a.userService.GetUserPermissions(userID)
		if err != nil {
			utils.InternalServerError(c, "Failed to get user permissions")
			c.Abort()
			return
		}

		hasPermission := false
		for _, perm := range permissions {
			if perm.Code == permissionCode {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			utils.Error(c, http.StatusForbidden, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (a *AuthMiddleware) RequireRole(roleCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			utils.Error(c, http.StatusUnauthorized, "User not authenticated")
			c.Abort()
			return
		}

		roles, err := a.userService.GetUserRoles(userID)
		if err != nil {
			utils.InternalServerError(c, "Failed to get user roles")
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range roles {
			for _, requiredCode := range roleCodes {
				if role.Code == requiredCode {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			utils.Error(c, http.StatusForbidden, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (a *AuthMiddleware) RequireAnyPermission(permissionCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			utils.Error(c, http.StatusUnauthorized, "User not authenticated")
			c.Abort()
			return
		}

		permissions, err := a.userService.GetUserPermissions(userID)
		if err != nil {
			utils.InternalServerError(c, "Failed to get user permissions")
			c.Abort()
			return
		}

		hasPermission := false
		for _, perm := range permissions {
			for _, requiredCode := range permissionCodes {
				if perm.Code == requiredCode {
					hasPermission = true
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if !hasPermission {
			utils.Error(c, http.StatusForbidden, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (a *AuthMiddleware) RequireAllPermissions(permissionCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			utils.Error(c, http.StatusUnauthorized, "User not authenticated")
			c.Abort()
			return
		}

		permissions, err := a.userService.GetUserPermissions(userID)
		if err != nil {
			utils.InternalServerError(c, "Failed to get user permissions")
			c.Abort()
			return
		}

		permMap := make(map[string]bool)
		for _, perm := range permissions {
			permMap[perm.Code] = true
		}

		hasAllPermissions := true
		for _, requiredCode := range permissionCodes {
			if !permMap[requiredCode] {
				hasAllPermissions = false
				break
			}
		}

		if !hasAllPermissions {
			utils.Error(c, http.StatusForbidden, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

func GetUserID(c *gin.Context) (uint, bool) {
	userID, ok := c.Request.Context().Value(UserIDKey).(uint)
	return userID, ok
}

func GetUserRoles(c *gin.Context) ([]string, bool) {
	roles, ok := c.Request.Context().Value(UserRolesKey).([]string)
	return roles, ok
}

func GetUserPermissions(c *gin.Context) ([]string, bool) {
	perms, ok := c.Request.Context().Value(UserPermsKey).([]string)
	return perms, ok
}
