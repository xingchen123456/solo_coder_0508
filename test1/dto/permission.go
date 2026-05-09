package dto

import "time"

type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required" example:"用户创建"`
	Code        string `json:"code" binding:"required" example:"user:create"`
	Type        string `json:"type" binding:"required" example:"api"`
	Resource    string `json:"resource" binding:"required" example:"user"`
	Action      string `json:"action" binding:"required" example:"create"`
	Path        string `json:"path" example:"/api/users"`
	Method      string `json:"method" example:"POST"`
	Description string `json:"description" example:"创建用户权限"`
	Status      int    `json:"status" example:"1"`
}

type UpdatePermissionRequest struct {
	Name        string `json:"name" example:"用户创建"`
	Code        string `json:"code" example:"user:create"`
	Type        string `json:"type" example:"api"`
	Resource    string `json:"resource" example:"user"`
	Action      string `json:"action" example:"create"`
	Path        string `json:"path" example:"/api/users"`
	Method      string `json:"method" example:"POST"`
	Description string `json:"description" example:"创建用户权限"`
	Status      int    `json:"status" example:"1"`
}

type GetPermissionByIDRequest struct {
	ID uint `uri:"id" binding:"required"`
}

type DeletePermissionRequest struct {
	ID uint `uri:"id" binding:"required"`
}

type PermissionListRequest struct {
	Name     string `form:"name"`
	Code     string `form:"code"`
	Type     string `form:"type"`
	Resource string `form:"resource"`
	Action   string `form:"action"`
	Status   *int   `form:"status"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=10"`
}

type PermissionResponse struct {
	ID          uint      `json:"id" example:"1"`
	Name        string    `json:"name" example:"用户创建"`
	Code        string    `json:"code" example:"user:create"`
	Type        string    `json:"type" example:"api"`
	Resource    string    `json:"resource" example:"user"`
	Action      string    `json:"action" example:"create"`
	Path        string    `json:"path" example:"/api/users"`
	Method      string    `json:"method" example:"POST"`
	Description string    `json:"description" example:"创建用户权限"`
	Status      int       `json:"status" example:"1"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PermissionListResponse struct {
	Items    []PermissionResponse `json:"items"`
	Total    int                  `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}
