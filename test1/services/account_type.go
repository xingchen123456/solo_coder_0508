package services

import (
	"errors"
	"fmt"

	"management-system/dto"
	"management-system/models"
	"management-system/utils"

	"gorm.io/gorm"
)

type AccountTypeService struct{}

func NewAccountTypeService() *AccountTypeService {
	return &AccountTypeService{}
}

func toAccountTypeModel(req *dto.CreateAccountTypeRequest) *models.AccountType {
	return &models.AccountType{
		Name:        req.Name,
		Code:        req.Code,
		Purpose:     req.Purpose,
		Description: req.Description,
		RoleCode:    req.RoleCode,
		Status:      req.Status,
	}
}

func toAccountTypeResponse(accountType *models.AccountType) dto.AccountTypeResponse {
	return dto.AccountTypeResponse{
		ID:          accountType.ID,
		Name:        accountType.Name,
		Code:        accountType.Code,
		Purpose:     accountType.Purpose,
		Description: accountType.Description,
		RoleCode:    accountType.RoleCode,
		Status:      accountType.Status,
		CreatedAt:   accountType.CreatedAt,
		UpdatedAt:   accountType.UpdatedAt,
	}
}

func (s *AccountTypeService) Create(req *dto.CreateAccountTypeRequest) (*dto.AccountTypeResponse, error) {
	accountType := toAccountTypeModel(req)
	if err := utils.DB.Create(accountType).Error; err != nil {
		return nil, err
	}
	resp := toAccountTypeResponse(accountType)
	return &resp, nil
}

func (s *AccountTypeService) GetByID(id uint) (*dto.AccountTypeResponse, error) {
	var accountType models.AccountType
	if err := utils.DB.First(&accountType, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("account type not found")
		}
		return nil, err
	}
	resp := toAccountTypeResponse(&accountType)
	return &resp, nil
}

func (s *AccountTypeService) List(req *dto.AccountTypeListRequest) (*dto.AccountTypeListResponse, error) {
	var accountTypes []models.AccountType
	var total int64

	query := utils.DB.Model(&models.AccountType{})

	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Code != "" {
		query = query.Where("code LIKE ?", "%"+req.Code+"%")
	}
	if req.Purpose != "" {
		query = query.Where("purpose LIKE ?", "%"+req.Purpose+"%")
	}
	if req.RoleCode != "" {
		query = query.Where("role_code = ?", req.RoleCode)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Find(&accountTypes).Error; err != nil {
		return nil, err
	}

	items := make([]dto.AccountTypeResponse, 0, len(accountTypes))
	for _, accountType := range accountTypes {
		items = append(items, toAccountTypeResponse(&accountType))
	}

	return &dto.AccountTypeListResponse{
		Items:    items,
		Total:    int(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *AccountTypeService) Update(id uint, req *dto.UpdateAccountTypeRequest) (*dto.AccountTypeResponse, error) {
	var existing models.AccountType
	if err := utils.DB.First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("account type not found")
		}
		return nil, err
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Code != "" {
		existing.Code = req.Code
	}
	if req.Purpose != "" {
		existing.Purpose = req.Purpose
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.RoleCode != "" {
		existing.RoleCode = req.RoleCode
	}
	if req.Status != 0 {
		existing.Status = req.Status
	}

	if err := utils.DB.Save(&existing).Error; err != nil {
		return nil, err
	}

	resp := toAccountTypeResponse(&existing)
	return &resp, nil
}

func (s *AccountTypeService) Delete(id uint) error {
	result := utils.DB.Delete(&models.AccountType{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("account type not found")
	}
	return nil
}
