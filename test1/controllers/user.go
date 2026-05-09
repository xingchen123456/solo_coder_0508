package controllers

import (
	"management-system/dto"
	"management-system/services"
	"management-system/utils"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController() *UserController {
	return &UserController{
		userService: services.NewUserService(),
	}
}

// Create godoc
// @Summary 创建用户
// @Description 创建一个新用户，支持同时分配角色
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "用户信息"
// @Success 200 {object} utils.Response{data=dto.UserResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/users [post]
func (u *UserController) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request data")
		return
	}

	data, err := u.userService.Create(&req)
	if err != nil {
		if err.Error() == "some roles not found" {
			utils.BadRequest(c, "Some roles not found")
		} else {
			utils.InternalServerError(c, "Failed to create user")
		}
		return
	}

	utils.Success(c, data)
}

// Get godoc
// @Summary 获取用户详情
// @Description 根据ID获取用户详情，包含角色信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} utils.Response{data=dto.UserResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "用户不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/users/{id} [get]
func (u *UserController) Get(c *gin.Context) {
	var req dto.GetUserByIDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.BadRequest(c, "Invalid user ID")
		return
	}

	data, err := u.userService.GetByID(req.ID)
	if err != nil {
		if err.Error() == "user not found" {
			utils.NotFound(c, "User not found")
		} else {
			utils.InternalServerError(c, "Failed to get user")
		}
		return
	}

	utils.Success(c, data)
}

// List godoc
// @Summary 获取用户列表
// @Description 获取用户列表，支持分页和条件查询
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param username query string false "用户名（模糊查询）"
// @Param nickname query string false "昵称（模糊查询）"
// @Param email query string false "邮箱（模糊查询）"
// @Param status query int false "状态（1启用 0禁用）"
// @Param page query int false "页码（默认1）"
// @Param page_size query int false "每页数量（默认10）"
// @Success 200 {object} utils.Response{data=dto.UserListResponse} "成功"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/users [get]
func (u *UserController) List(c *gin.Context) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(c, "Invalid request parameters")
		return
	}

	data, err := u.userService.List(&req)
	if err != nil {
		utils.InternalServerError(c, "Failed to get users")
		return
	}

	utils.Success(c, data)
}

// Update godoc
// @Summary 更新用户
// @Description 根据ID更新用户信息，支持同时更新角色
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param user body dto.UpdateUserRequest true "用户信息"
// @Success 200 {object} utils.Response{data=dto.UserResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "用户不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/users/{id} [put]
func (u *UserController) Update(c *gin.Context) {
	var uriReq dto.GetUserByIDRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		utils.BadRequest(c, "Invalid user ID")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request data")
		return
	}

	data, err := u.userService.Update(uriReq.ID, &req)
	if err != nil {
		if err.Error() == "user not found" {
			utils.NotFound(c, "User not found")
		} else if err.Error() == "some roles not found" {
			utils.BadRequest(c, "Some roles not found")
		} else {
			utils.InternalServerError(c, "Failed to update user")
		}
		return
	}

	utils.Success(c, data)
}

// Delete godoc
// @Summary 删除用户
// @Description 根据ID删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} utils.Response "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "用户不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/users/{id} [delete]
func (u *UserController) Delete(c *gin.Context) {
	var req dto.DeleteUserRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.BadRequest(c, "Invalid user ID")
		return
	}

	if err := u.userService.Delete(req.ID); err != nil {
		if err.Error() == "user not found" {
			utils.NotFound(c, "User not found")
		} else {
			utils.InternalServerError(c, "Failed to delete user")
		}
		return
	}

	utils.Success(c, nil)
}

// AssignRoles godoc
// @Summary 为用户分配角色
// @Description 为指定用户分配多个角色（覆盖原有角色）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param roles body dto.AssignUserRolesRequest true "角色ID列表"
// @Success 200 {object} utils.Response{data=dto.UserResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "用户不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/users/{id}/roles [put]
func (u *UserController) AssignRoles(c *gin.Context) {
	var uriReq dto.GetUserByIDRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		utils.BadRequest(c, "Invalid user ID")
		return
	}

	var req dto.AssignUserRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request data")
		return
	}

	data, err := u.userService.AssignRoles(uriReq.ID, req.RoleIDs)
	if err != nil {
		if err.Error() == "user not found" {
			utils.NotFound(c, "User not found")
		} else if err.Error() == "some roles not found" {
			utils.BadRequest(c, "Some roles not found")
		} else {
			utils.InternalServerError(c, "Failed to assign roles")
		}
		return
	}

	utils.Success(c, data)
}

// GetUserRoles godoc
// @Summary 获取用户的角色列表
// @Description 获取指定用户的所有角色
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} utils.Response{data=[]dto.RoleResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "用户不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/users/{id}/roles [get]
func (u *UserController) GetUserRoles(c *gin.Context) {
	var req dto.GetUserByIDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.BadRequest(c, "Invalid user ID")
		return
	}

	data, err := u.userService.GetUserRoles(req.ID)
	if err != nil {
		if err.Error() == "user not found" {
			utils.NotFound(c, "User not found")
		} else {
			utils.InternalServerError(c, "Failed to get user roles")
		}
		return
	}

	utils.Success(c, data)
}

// GetUserPermissions godoc
// @Summary 获取用户的权限列表
// @Description 获取指定用户的所有权限（从角色中聚合）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} utils.Response{data=[]dto.PermissionResponse} "成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "用户不存在"
// @Failure 500 {object} utils.Response "服务器内部错误"
// @Router /api/users/{id}/permissions [get]
func (u *UserController) GetUserPermissions(c *gin.Context) {
	var req dto.GetUserByIDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		utils.BadRequest(c, "Invalid user ID")
		return
	}

	data, err := u.userService.GetUserPermissions(req.ID)
	if err != nil {
		if err.Error() == "user not found" {
			utils.NotFound(c, "User not found")
		} else {
			utils.InternalServerError(c, "Failed to get user permissions")
		}
		return
	}

	utils.Success(c, data)
}
