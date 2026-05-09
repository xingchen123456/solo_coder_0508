package controllers

import (
	"management-system/dto"
	"management-system/services"
	"management-system/utils"

	"github.com/gin-gonic/gin"
)

type RoleController struct {
	roleService *services.RoleService
}

func NewRoleController() *RoleController {
	return &RoleController{
		roleService: services.NewRoleService(),
	}
}

// CreateRole godoc
// @Summary 创建角色
// @Description 创建一个新角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param request body dto.CreateRoleRequest true "角色信息"
// @Success 200 {object} utils.Response{data=dto.RoleResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/roles [post]
func (r *RoleController) Create(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request data")
		return
	}

	data, err := r.roleService.Create(&req)
	if err != nil {
		utils.InternalServerError(c, "Failed to create role")
		return
	}

	utils.Success(c, data)
}

// GetRole godoc
// @Summary 获取角色详情
// @Description 根据ID获取角色详情
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Success 200 {object} utils.Response{data=dto.RoleResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "角色不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/roles/{id} [get]
func (r *RoleController) Get(c *gin.Context) {
	var req dto.GetRoleByIDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.BadRequest(c, "Invalid role ID")
		return
	}

	data, err := r.roleService.GetByID(req.ID)
	if err != nil {
		if err.Error() == "role not found" {
			utils.NotFound(c, "Role not found")
		} else {
			utils.InternalServerError(c, "Failed to get role")
		}
		return
	}

	utils.Success(c, data)
}

// GetRoleWithPermissions godoc
// @Summary 获取角色详情（包含权限）
// @Description 根据ID获取角色详情，包含该角色的所有权限
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Success 200 {object} utils.Response{data=dto.RoleWithPermissionsResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "角色不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/roles/{id}/permissions [get]
func (r *RoleController) GetWithPermissions(c *gin.Context) {
	var req dto.GetRoleByIDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.BadRequest(c, "Invalid role ID")
		return
	}

	data, err := r.roleService.GetWithPermissions(req.ID)
	if err != nil {
		if err.Error() == "role not found" {
			utils.NotFound(c, "Role not found")
		} else {
			utils.InternalServerError(c, "Failed to get role with permissions")
		}
		return
	}

	utils.Success(c, data)
}

// ListRoles godoc
// @Summary 获取角色列表
// @Description 获取角色列表，支持分页和条件查询
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param name query string false "角色名称（模糊查询）"
// @Param code query string false "角色编码（模糊查询）"
// @Param status query int false "状态（1启用 0禁用）"
// @Param page query int false "页码（默认1）"
// @Param page_size query int false "每页数量（默认10）"
// @Success 200 {object} utils.Response{data=dto.RoleListResponse} "成功"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/roles [get]
func (r *RoleController) List(c *gin.Context) {
	var req dto.RoleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(c, "Invalid request parameters")
		return
	}

	data, err := r.roleService.List(&req)
	if err != nil {
		utils.InternalServerError(c, "Failed to get roles")
		return
	}

	utils.Success(c, data)
}

// UpdateRole godoc
// @Summary 更新角色
// @Description 根据ID更新角色信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Param request body dto.UpdateRoleRequest true "角色信息"
// @Success 200 {object} utils.Response{data=dto.RoleResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "角色不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/roles/{id} [put]
func (r *RoleController) Update(c *gin.Context) {
	var uriReq dto.GetRoleByIDRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		utils.BadRequest(c, "Invalid role ID")
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request data")
		return
	}

	data, err := r.roleService.Update(uriReq.ID, &req)
	if err != nil {
		if err.Error() == "role not found" {
			utils.NotFound(c, "Role not found")
		} else {
			utils.InternalServerError(c, "Failed to update role")
		}
		return
	}

	utils.Success(c, data)
}

// DeleteRole godoc
// @Summary 删除角色
// @Description 根据ID删除角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Success 200 {object} utils.Response "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "角色不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/roles/{id} [delete]
func (r *RoleController) Delete(c *gin.Context) {
	var req dto.DeleteRoleRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.BadRequest(c, "Invalid role ID")
		return
	}

	if err := r.roleService.Delete(req.ID); err != nil {
		if err.Error() == "role not found" {
			utils.NotFound(c, "Role not found")
		} else {
			utils.InternalServerError(c, "Failed to delete role")
		}
		return
	}

	utils.Success(c, nil)
}

// AssignPermissions godoc
// @Summary 为角色分配权限
// @Description 为指定角色分配多个权限（覆盖原有权限）
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Param permissions body dto.AssignRolePermissionsRequest true "权限ID列表"
// @Success 200 {object} utils.Response{data=dto.RoleWithPermissionsResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "角色不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/roles/{id}/permissions [put]
func (r *RoleController) AssignPermissions(c *gin.Context) {
	var uriReq dto.GetRoleByIDRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		utils.BadRequest(c, "Invalid role ID")
		return
	}

	var req dto.AssignRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request data")
		return
	}

	data, err := r.roleService.AssignPermissions(uriReq.ID, req.PermissionIDs)
	if err != nil {
		if err.Error() == "role not found" {
			utils.NotFound(c, "Role not found")
		} else if err.Error() == "some permissions not found" {
			utils.BadRequest(c, "Some permissions not found")
		} else {
			utils.InternalServerError(c, "Failed to assign permissions")
		}
		return
	}

	utils.Success(c, data)
}

// GetRolePermissions godoc
// @Summary 获取角色的权限列表
// @Description 获取指定角色的所有权限
// @Tags 角色管理
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Success 200 {object} utils.Response{data=[]dto.PermissionResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "角色不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/roles/{id}/permissions [get]
func (r *RoleController) GetRolePermissions(c *gin.Context) {
	var req dto.GetRoleByIDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.BadRequest(c, "Invalid role ID")
		return
	}

	data, err := r.roleService.GetRolePermissions(req.ID)
	if err != nil {
		if err.Error() == "role not found" {
			utils.NotFound(c, "Role not found")
		} else {
			utils.InternalServerError(c, "Failed to get role permissions")
		}
		return
	}

	utils.Success(c, data)
}
