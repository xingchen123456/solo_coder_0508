package dto

import "time"

type AssignRolePermissionsRequest struct {
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

type RoleWithPermissionsResponse struct {
	ID          uint                 `json:"id" example:"1"`
	Name        string               `json:"name" example:"管理员"`
	Code        string               `json:"code" example:"admin"`
	Description string               `json:"description" example:"系统管理员"`
	Status      int                  `json:"status" example:"1"`
	Permissions []PermissionResponse `json:"permissions,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}
