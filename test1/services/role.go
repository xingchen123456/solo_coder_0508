package services

import (
	"errors"
	"fmt"

	"management-system/dto"
	"management-system/models"
	"management-system/utils"

	"gorm.io/gorm"
)

type RoleService struct{}

func NewRoleService() *RoleService {
	return &RoleService{}
}

func toRoleModel(req *dto.CreateRoleRequest) *models.Role {
	return &models.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      req.Status,
	}
}

func toRoleResponse(role *models.Role) dto.RoleResponse {
	return dto.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		Status:      role.Status,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

func toRoleWithPermissionsResponse(role *models.Role) dto.RoleWithPermissionsResponse {
	permissions := make([]dto.PermissionResponse, 0, len(role.Permissions))
	for _, perm := range role.Permissions {
		permissions = append(permissions, toPermissionResponse(&perm))
	}
	return dto.RoleWithPermissionsResponse{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		Status:      role.Status,
		Permissions: permissions,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

func (s *RoleService) Create(req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	role := toRoleModel(req)
	if err := utils.DB.Create(role).Error; err != nil {
		return nil, err
	}
	resp := toRoleResponse(role)
	return &resp, nil
}

func (s *RoleService) GetByID(id uint) (*dto.RoleResponse, error) {
	var role models.Role
	if err := utils.DB.First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("role not found")
		}
		return nil, err
	}
	resp := toRoleResponse(&role)
	return &resp, nil
}

func (s *RoleService) GetWithPermissions(id uint) (*dto.RoleWithPermissionsResponse, error) {
	var role models.Role
	if err := utils.DB.Preload("Permissions").First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("role not found")
		}
		return nil, err
	}
	resp := toRoleWithPermissionsResponse(&role)
	return &resp, nil
}

func (s *RoleService) List(req *dto.RoleListRequest) (*dto.RoleListResponse, error) {
	var roles []models.Role
	var total int64

	query := utils.DB.Model(&models.Role{})

	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Code != "" {
		query = query.Where("code LIKE ?", "%"+req.Code+"%")
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Find(&roles).Error; err != nil {
		return nil, err
	}

	items := make([]dto.RoleResponse, 0, len(roles))
	for _, role := range roles {
		items = append(items, toRoleResponse(&role))
	}

	return &dto.RoleListResponse{
		Items:    items,
		Total:    int(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *RoleService) Update(id uint, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	var existing models.Role
	if err := utils.DB.First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("role not found")
		}
		return nil, err
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Code != "" {
		existing.Code = req.Code
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Status != 0 {
		existing.Status = req.Status
	}

	if err := utils.DB.Save(&existing).Error; err != nil {
		return nil, err
	}

	resp := toRoleResponse(&existing)
	return &resp, nil
}

func (s *RoleService) Delete(id uint) error {
	result := utils.DB.Delete(&models.Role{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("role not found")
	}
	return nil
}

func (s *RoleService) AssignPermissions(roleID uint, permissionIDs []uint) (*dto.RoleWithPermissionsResponse, error) {
	var role models.Role
	if err := utils.DB.Preload("Permissions").First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("role not found")
		}
		return nil, err
	}

	if len(permissionIDs) > 0 {
		var permissions []models.Permission
		if err := utils.DB.Find(&permissions, permissionIDs).Error; err != nil {
			return nil, err
		}
		if len(permissions) != len(permissionIDs) {
			return nil, fmt.Errorf("some permissions not found")
		}
	}

	if err := utils.DB.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error; err != nil {
		return nil, err
	}

	for _, permissionID := range permissionIDs {
		rolePermission := &models.RolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
		}
		if err := utils.DB.Create(rolePermission).Error; err != nil {
			return nil, err
		}
	}

	if err := utils.DB.Preload("Permissions").First(&role, roleID).Error; err != nil {
		return nil, err
	}

	resp := toRoleWithPermissionsResponse(&role)
	return &resp, nil
}

func (s *RoleService) GetRolePermissions(roleID uint) ([]dto.PermissionResponse, error) {
	var role models.Role
	if err := utils.DB.Preload("Permissions").First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("role not found")
		}
		return nil, err
	}

	permissions := make([]dto.PermissionResponse, 0, len(role.Permissions))
	for _, perm := range role.Permissions {
		permissions = append(permissions, toPermissionResponse(&perm))
	}
	return permissions, nil
}
