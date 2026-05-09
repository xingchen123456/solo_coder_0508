package services

import (
	"errors"
	"fmt"

	"management-system/dto"
	"management-system/models"
	"management-system/utils"

	"gorm.io/gorm"
)

type PermissionService struct{}

func NewPermissionService() *PermissionService {
	return &PermissionService{}
}

func toPermissionModel(req *dto.CreatePermissionRequest) *models.Permission {
	return &models.Permission{
		Name:        req.Name,
		Code:        req.Code,
		Type:        req.Type,
		Resource:    req.Resource,
		Action:      req.Action,
		Path:        req.Path,
		Method:      req.Method,
		Description: req.Description,
		Status:      req.Status,
	}
}

func toPermissionResponse(permission *models.Permission) dto.PermissionResponse {
	return dto.PermissionResponse{
		ID:          permission.ID,
		Name:        permission.Name,
		Code:        permission.Code,
		Type:        permission.Type,
		Resource:    permission.Resource,
		Action:      permission.Action,
		Path:        permission.Path,
		Method:      permission.Method,
		Description: permission.Description,
		Status:      permission.Status,
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
	}
}

func (s *PermissionService) Create(req *dto.CreatePermissionRequest) (*dto.PermissionResponse, error) {
	permission := toPermissionModel(req)
	if err := utils.DB.Create(permission).Error; err != nil {
		return nil, err
	}
	resp := toPermissionResponse(permission)
	return &resp, nil
}

func (s *PermissionService) GetByID(id uint) (*dto.PermissionResponse, error) {
	var permission models.Permission
	if err := utils.DB.First(&permission, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("permission not found")
		}
		return nil, err
	}
	resp := toPermissionResponse(&permission)
	return &resp, nil
}

func (s *PermissionService) List(req *dto.PermissionListRequest) (*dto.PermissionListResponse, error) {
	var permissions []models.Permission
	var total int64

	query := utils.DB.Model(&models.Permission{})

	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Code != "" {
		query = query.Where("code LIKE ?", "%"+req.Code+"%")
	}
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.Resource != "" {
		query = query.Where("resource = ?", req.Resource)
	}
	if req.Action != "" {
		query = query.Where("action = ?", req.Action)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Find(&permissions).Error; err != nil {
		return nil, err
	}

	items := make([]dto.PermissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		items = append(items, toPermissionResponse(&permission))
	}

	return &dto.PermissionListResponse{
		Items:    items,
		Total:    int(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *PermissionService) Update(id uint, req *dto.UpdatePermissionRequest) (*dto.PermissionResponse, error) {
	var existing models.Permission
	if err := utils.DB.First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("permission not found")
		}
		return nil, err
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Code != "" {
		existing.Code = req.Code
	}
	if req.Type != "" {
		existing.Type = req.Type
	}
	if req.Resource != "" {
		existing.Resource = req.Resource
	}
	if req.Action != "" {
		existing.Action = req.Action
	}
	if req.Path != "" {
		existing.Path = req.Path
	}
	if req.Method != "" {
		existing.Method = req.Method
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

	resp := toPermissionResponse(&existing)
	return &resp, nil
}

func (s *PermissionService) Delete(id uint) error {
	result := utils.DB.Delete(&models.Permission{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("permission not found")
	}
	return nil
}

func (s *PermissionService) GetByIDs(ids []uint) ([]dto.PermissionResponse, error) {
	var permissions []models.Permission
	if err := utils.DB.Find(&permissions, ids).Error; err != nil {
		return nil, err
	}

	items := make([]dto.PermissionResponse, 0, len(permissions))
	for _, permission := range permissions {
		items = append(items, toPermissionResponse(&permission))
	}
	return items, nil
}
