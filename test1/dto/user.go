package dto

import "time"

type CreateUserRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"123456"`
	Nickname string `json:"nickname" example:"管理员"`
	Email    string `json:"email" example:"admin@example.com"`
	Status   int    `json:"status" example:"1"`
	RoleIDs  []uint `json:"role_ids"`
}

type UpdateUserRequest struct {
	Username string `json:"username" example:"admin"`
	Password string `json:"password" example:"123456"`
	Nickname string `json:"nickname" example:"管理员"`
	Email    string `json:"email" example:"admin@example.com"`
	Status   int    `json:"status" example:"1"`
	RoleIDs  []uint `json:"role_ids"`
}

type GetUserByIDRequest struct {
	ID uint `uri:"id" binding:"required"`
}

type DeleteUserRequest struct {
	ID uint `uri:"id" binding:"required"`
}

type UserListRequest struct {
	Username string `form:"username"`
	Nickname string `form:"nickname"`
	Email    string `form:"email"`
	Status   *int   `form:"status"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=10"`
}

type AssignUserRolesRequest struct {
	RoleIDs []uint `json:"role_ids" binding:"required"`
}

type UserResponse struct {
	ID        uint           `json:"id" example:"1"`
	Username  string         `json:"username" example:"admin"`
	Nickname  string         `json:"nickname" example:"管理员"`
	Email     string         `json:"email" example:"admin@example.com"`
	Status    int            `json:"status" example:"1"`
	Roles     []RoleResponse `json:"roles,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type UserListResponse struct {
	Items    []UserResponse `json:"items"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}
