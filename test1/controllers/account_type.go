package controllers

import (
	"management-system/dto"
	"management-system/services"
	"management-system/utils"

	"github.com/gin-gonic/gin"
)

type AccountTypeController struct {
	accountTypeService *services.AccountTypeService
}

func NewAccountTypeController() *AccountTypeController {
	return &AccountTypeController{
		accountTypeService: services.NewAccountTypeService(),
	}
}

// CreateAccountType godoc
// @Summary 创建账户类型
// @Description 创建一个新账户类型
// @Tags 账户类型管理
// @Accept json
// @Produce json
// @Param request body dto.CreateAccountTypeRequest true "账户类型信息"
// @Success 200 {object} utils.Response{data=dto.AccountTypeResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/account-types [post]
func (a *AccountTypeController) Create(c *gin.Context) {
	var req dto.CreateAccountTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request data")
		return
	}

	data, err := a.accountTypeService.Create(&req)
	if err != nil {
		utils.InternalServerError(c, "Failed to create account type")
		return
	}

	utils.Success(c, data)
}

// GetAccountType godoc
// @Summary 获取账户类型详情
// @Description 根据ID获取账户类型详情
// @Tags 账户类型管理
// @Accept json
// @Produce json
// @Param id path int true "账户类型ID"
// @Success 200 {object} utils.Response{data=dto.AccountTypeResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "账户类型不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/account-types/{id} [get]
func (a *AccountTypeController) Get(c *gin.Context) {
	var req dto.GetAccountTypeByIDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.BadRequest(c, "Invalid account type ID")
		return
	}

	data, err := a.accountTypeService.GetByID(req.ID)
	if err != nil {
		if err.Error() == "account type not found" {
			utils.NotFound(c, "Account type not found")
		} else {
			utils.InternalServerError(c, "Failed to get account type")
		}
		return
	}

	utils.Success(c, data)
}

// ListAccountTypes godoc
// @Summary 获取账户类型列表
// @Description 获取账户类型列表，支持分页和条件查询
// @Tags 账户类型管理
// @Accept json
// @Produce json
// @Param name query string false "账户类型名称（模糊查询）"
// @Param code query string false "账户类型编码（模糊查询）"
// @Param purpose query string false "账户用途（模糊查询）"
// @Param role_code query string false "所属角色编码"
// @Param status query int false "状态（1启用 0禁用）"
// @Param page query int false "页码（默认1）"
// @Param page_size query int false "每页数量（默认10）"
// @Success 200 {object} utils.Response{data=dto.AccountTypeListResponse} "成功"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/account-types [get]
func (a *AccountTypeController) List(c *gin.Context) {
	var req dto.AccountTypeListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(c, "Invalid request parameters")
		return
	}

	data, err := a.accountTypeService.List(&req)
	if err != nil {
		utils.InternalServerError(c, "Failed to get account types")
		return
	}

	utils.Success(c, data)
}

// UpdateAccountType godoc
// @Summary 更新账户类型
// @Description 根据ID更新账户类型信息
// @Tags 账户类型管理
// @Accept json
// @Produce json
// @Param id path int true "账户类型ID"
// @Param request body dto.UpdateAccountTypeRequest true "账户类型信息"
// @Success 200 {object} utils.Response{data=dto.AccountTypeResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "账户类型不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/account-types/{id} [put]
func (a *AccountTypeController) Update(c *gin.Context) {
	var uriReq dto.GetAccountTypeByIDRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		utils.BadRequest(c, "Invalid account type ID")
		return
	}

	var req dto.UpdateAccountTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request data")
		return
	}

	data, err := a.accountTypeService.Update(uriReq.ID, &req)
	if err != nil {
		if err.Error() == "account type not found" {
			utils.NotFound(c, "Account type not found")
		} else {
			utils.InternalServerError(c, "Failed to update account type")
		}
		return
	}

	utils.Success(c, data)
}

// DeleteAccountType godoc
// @Summary 删除账户类型
// @Description 根据ID删除账户类型
// @Tags 账户类型管理
// @Accept json
// @Produce json
// @Param id path int true "账户类型ID"
// @Success 200 {object} utils.Response "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "账户类型不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/account-types/{id} [delete]
func (a *AccountTypeController) Delete(c *gin.Context) {
	var req dto.DeleteAccountTypeRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.BadRequest(c, "Invalid account type ID")
		return
	}

	if err := a.accountTypeService.Delete(req.ID); err != nil {
		if err.Error() == "account type not found" {
			utils.NotFound(c, "Account type not found")
		} else {
			utils.InternalServerError(c, "Failed to delete account type")
		}
		return
	}

	utils.Success(c, nil)
}
