package dto

import "time"

type CreateAccountTypeRequest struct {
	Name        string `json:"name" binding:"required" example:"我的佣金（平台）"`
	Code        string `json:"code" binding:"required" example:"platform_commission"`
	Purpose     string `json:"purpose" binding:"required" example:"收取佣金"`
	Description string `json:"description" example:"结算后平台应得的佣金部分"`
	RoleCode    string `json:"role_code" binding:"required" example:"platform"`
	Status      int    `json:"status" example:"1"`
}

type UpdateAccountTypeRequest struct {
	Name        string `json:"name" example:"我的佣金（平台）"`
	Code        string `json:"code" example:"platform_commission"`
	Purpose     string `json:"purpose" example:"收取佣金"`
	Description string `json:"description" example:"结算后平台应得的佣金部分"`
	RoleCode    string `json:"role_code" example:"platform"`
	Status      int    `json:"status" example:"1"`
}

type GetAccountTypeByIDRequest struct {
	ID uint `uri:"id" binding:"required"`
}

type DeleteAccountTypeRequest struct {
	ID uint `uri:"id" binding:"required"`
}

type AccountTypeListRequest struct {
	Name     string `form:"name"`
	Code     string `form:"code"`
	Purpose  string `form:"purpose"`
	RoleCode string `form:"role_code"`
	Status   *int   `form:"status"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=10"`
}

type AccountTypeResponse struct {
	ID          uint      `json:"id" example:"1"`
	Name        string    `json:"name" example:"我的佣金（平台）"`
	Code        string    `json:"code" example:"platform_commission"`
	Purpose     string    `json:"purpose" example:"收取佣金"`
	Description string    `json:"description" example:"结算后平台应得的佣金部分"`
	RoleCode    string    `json:"role_code" example:"platform"`
	Status      int       `json:"status" example:"1"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AccountTypeListResponse struct {
	Items    []AccountTypeResponse `json:"items"`
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}
