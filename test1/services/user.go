package services

import (
	"context"
	"errors"
	"fmt"

	"management-system/dto"
	"management-system/models"
	"management-system/utils"

	"gorm.io/gorm"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func toUserModel(req *dto.CreateUserRequest) *models.User {
	return &models.User{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Email:    req.Email,
		Status:   req.Status,
	}
}

func toUserResponse(user *models.User) dto.UserResponse {
	roles := make([]dto.RoleResponse, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, toRoleResponse(&role))
	}
	return dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Status:    user.Status,
		Roles:     roles,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (s *UserService) Create(req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	user := toUserModel(req)

	var roles []models.Role
	if len(req.RoleIDs) > 0 {
		if err := utils.DB.Find(&roles, req.RoleIDs).Error; err != nil {
			return nil, err
		}
		if len(roles) != len(req.RoleIDs) {
			return nil, fmt.Errorf("some roles not found")
		}
	}

	if err := utils.DB.Create(user).Error; err != nil {
		return nil, err
	}

	if len(req.RoleIDs) > 0 {
		for _, roleID := range req.RoleIDs {
			userRole := &models.UserRole{
				UserID: user.ID,
				RoleID: roleID,
			}
			if err := utils.DB.Create(userRole).Error; err != nil {
				return nil, err
			}
		}
	}

	if err := utils.DB.Preload("Roles").First(user, user.ID).Error; err != nil {
		return nil, err
	}

	resp := toUserResponse(user)
	return &resp, nil
}

func (s *UserService) GetByID(id uint) (*dto.UserResponse, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d", id)

	cachedUser, err := utils.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil && cachedUser != "" {
		var user models.User
		if cachedUser != "" {
			user.ID = id
			resp := toUserResponse(&user)
			return &resp, nil
		}
	}

	var user models.User
	if err := utils.DB.Preload("Roles").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	resp := toUserResponse(&user)
	return &resp, nil
}

func (s *UserService) List(req *dto.UserListRequest) (*dto.UserListResponse, error) {
	var users []models.User
	var total int64

	query := utils.DB.Model(&models.User{})

	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+req.Nickname+"%")
	}
	if req.Email != "" {
		query = query.Where("email LIKE ?", "%"+req.Email+"%")
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (req.Page - 1) * req.PageSize
	if err := query.Preload("Roles").Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		return nil, err
	}

	items := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		items = append(items, toUserResponse(&user))
	}

	return &dto.UserListResponse{
		Items:    items,
		Total:    int(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *UserService) Update(id uint, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	var existing models.User
	if err := utils.DB.Preload("Roles").First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	if req.Username != "" {
		existing.Username = req.Username
	}
	if req.Password != "" {
		existing.Password = req.Password
	}
	if req.Nickname != "" {
		existing.Nickname = req.Nickname
	}
	if req.Email != "" {
		existing.Email = req.Email
	}
	if req.Status != 0 {
		existing.Status = req.Status
	}

	if req.RoleIDs != nil {
		if len(req.RoleIDs) > 0 {
			var roles []models.Role
			if err := utils.DB.Find(&roles, req.RoleIDs).Error; err != nil {
				return nil, err
			}
			if len(roles) != len(req.RoleIDs) {
				return nil, fmt.Errorf("some roles not found")
			}
		}

		if err := utils.DB.Where("user_id = ?", id).Delete(&models.UserRole{}).Error; err != nil {
			return nil, err
		}

		for _, roleID := range req.RoleIDs {
			userRole := &models.UserRole{
				UserID: id,
				RoleID: roleID,
			}
			if err := utils.DB.Create(userRole).Error; err != nil {
				return nil, err
			}
		}
	}

	if err := utils.DB.Save(&existing).Error; err != nil {
		return nil, err
	}

	if err := utils.DB.Preload("Roles").First(&existing, id).Error; err != nil {
		return nil, err
	}

	resp := toUserResponse(&existing)
	return &resp, nil
}

func (s *UserService) Delete(id uint) error {
	result := utils.DB.Delete(&models.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (s *UserService) AssignRoles(userID uint, roleIDs []uint) (*dto.UserResponse, error) {
	var user models.User
	if err := utils.DB.Preload("Roles").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	if len(roleIDs) > 0 {
		var roles []models.Role
		if err := utils.DB.Find(&roles, roleIDs).Error; err != nil {
			return nil, err
		}
		if len(roles) != len(roleIDs) {
			return nil, fmt.Errorf("some roles not found")
		}
	}

	if err := utils.DB.Where("user_id = ?", userID).Delete(&models.UserRole{}).Error; err != nil {
		return nil, err
	}

	for _, roleID := range roleIDs {
		userRole := &models.UserRole{
			UserID: userID,
			RoleID: roleID,
		}
		if err := utils.DB.Create(userRole).Error; err != nil {
			return nil, err
		}
	}

	if err := utils.DB.Preload("Roles").First(&user, userID).Error; err != nil {
		return nil, err
	}

	resp := toUserResponse(&user)
	return &resp, nil
}

func (s *UserService) GetUserRoles(userID uint) ([]dto.RoleResponse, error) {
	var user models.User
	if err := utils.DB.Preload("Roles").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	roles := make([]dto.RoleResponse, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, toRoleResponse(&role))
	}
	return roles, nil
}

func (s *UserService) GetUserPermissions(userID uint) ([]dto.PermissionResponse, error) {
	var user models.User
	if err := utils.DB.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	permissionMap := make(map[uint]dto.PermissionResponse)
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			if _, exists := permissionMap[permission.ID]; !exists {
				permissionMap[permission.ID] = toPermissionResponse(&permission)
			}
		}
	}

	permissions := make([]dto.PermissionResponse, 0, len(permissionMap))
	for _, perm := range permissionMap {
		permissions = append(permissions, perm)
	}
	return permissions, nil
}
