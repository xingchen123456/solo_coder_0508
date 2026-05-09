package controllers

import (
	"management-system/dto"
	"management-system/services"
	"management-system/utils"

	"github.com/gin-gonic/gin"
)

type PermissionController struct {
	permissionService *services.PermissionService
}

func NewPermissionController() *PermissionController {
	return &PermissionController{
		permissionService: services.NewPermissionService(),
	}
}

// CreatePermission godoc
// @Summary 创建权限
// @Description 创建一个新权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param request body dto.CreatePermissionRequest true "权限信息"
// @Success 200 {object} utils.Response{data=dto.PermissionResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/permissions [post]
func (p *PermissionController) Create(c *gin.Context) {
	var req dto.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request data")
		return
	}

	data, err := p.permissionService.Create(&req)
	if err != nil {
		utils.InternalServerError(c, "Failed to create permission")
		return
	}

	utils.Success(c, data)
}

// GetPermission godoc
// @Summary 获取权限详情
// @Description 根据ID获取权限详情
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param id path int true "权限ID"
// @Success 200 {object} utils.Response{data=dto.PermissionResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "权限不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/permissions/{id} [get]
func (p *PermissionController) Get(c *gin.Context) {
	var req dto.GetPermissionByIDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.BadRequest(c, "Invalid permission ID")
		return
	}

	data, err := p.permissionService.GetByID(req.ID)
	if err != nil {
		if err.Error() == "permission not found" {
			utils.NotFound(c, "Permission not found")
		} else {
			utils.InternalServerError(c, "Failed to get permission")
		}
		return
	}

	utils.Success(c, data)
}

// ListPermissions godoc
// @Summary 获取权限列表
// @Description 获取权限列表，支持分页和条件查询
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param name query string false "权限名称（模糊查询）"
// @Param code query string false "权限编码（模糊查询）"
// @Param type query string false "权限类型（api/menu/button）"
// @Param resource query string false "资源名称"
// @Param action query string false "操作类型"
// @Param status query int false "状态（1启用 0禁用）"
// @Param page query int false "页码（默认1）"
// @Param page_size query int false "每页数量（默认10）"
// @Success 200 {object} utils.Response{data=dto.PermissionListResponse} "成功"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/permissions [get]
func (p *PermissionController) List(c *gin.Context) {
	var req dto.PermissionListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(c, "Invalid request parameters")
		return
	}

	data, err := p.permissionService.List(&req)
	if err != nil {
		utils.InternalServerError(c, "Failed to get permissions")
		return
	}

	utils.Success(c, data)
}

// UpdatePermission godoc
// @Summary 更新权限
// @Description 根据ID更新权限信息
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param id path int true "权限ID"
// @Param request body dto.UpdatePermissionRequest true "权限信息"
// @Success 200 {object} utils.Response{data=dto.PermissionResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "权限不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/permissions/{id} [put]
func (p *PermissionController) Update(c *gin.Context) {
	var uriReq dto.GetPermissionByIDRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		utils.BadRequest(c, "Invalid permission ID")
		return
	}

	var req dto.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request data")
		return
	}

	data, err := p.permissionService.Update(uriReq.ID, &req)
	if err != nil {
		if err.Error() == "permission not found" {
			utils.NotFound(c, "Permission not found")
		} else {
			utils.InternalServerError(c, "Failed to update permission")
		}
		return
	}

	utils.Success(c, data)
}

// DeletePermission godoc
// @Summary 删除权限
// @Description 根据ID删除权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Param id path int true "权限ID"
// @Success 200 {object} utils.Response "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "权限不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/permissions/{id} [delete]
func (p *PermissionController) Delete(c *gin.Context) {
	var req dto.DeletePermissionRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.BadRequest(c, "Invalid permission ID")
		return
	}

	if err := p.permissionService.Delete(req.ID); err != nil {
		if err.Error() == "permission not found" {
			utils.NotFound(c, "Permission not found")
		} else {
			utils.InternalServerError(c, "Failed to delete permission")
		}
		return
	}

	utils.Success(c, nil)
}
