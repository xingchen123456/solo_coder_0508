package dto

import "time"

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required" example:"管理员"`
	Code        string `json:"code" binding:"required" example:"admin"`
	Description string `json:"description" example:"系统管理员"`
	Status      int    `json:"status" example:"1"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name" example:"管理员"`
	Code        string `json:"code" example:"admin"`
	Description string `json:"description" example:"系统管理员"`
	Status      int    `json:"status" example:"1"`
}

type GetRoleByIDRequest struct {
	ID uint `uri:"id" binding:"required"`
}

type DeleteRoleRequest struct {
	ID uint `uri:"id" binding:"required"`
}

type RoleListRequest struct {
	Name     string `form:"name"`
	Code     string `form:"code"`
	Status   *int   `form:"status"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=10"`
}

type RoleResponse struct {
	ID          uint      `json:"id" example:"1"`
	Name        string    `json:"name" example:"管理员"`
	Code        string    `json:"code" example:"admin"`
	Description string    `json:"description" example:"系统管理员"`
	Status      int       `json:"status" example:"1"`
	CreatedAt   time.Time `json:"created_at" example:"2026-05-09T12:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2026-05-09T12:00:00Z"`
}

type RoleListResponse struct {
	Items    []RoleResponse `json:"items"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}
