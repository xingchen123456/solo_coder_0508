package services

import (
	"context"
	"errors"
	"fmt"

	"management-system/models"
	"management-system/utils"

	"gorm.io/gorm"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) Create(user *models.User) error {
	return utils.DB.Create(user).Error
}

func (s *UserService) GetByID(id uint) (*models.User, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d", id)

	cachedUser, err := utils.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil && cachedUser != "" {
		var user models.User
		if cachedUser != "" {
			user.ID = id
			return &user, nil
		}
	}

	var user models.User
	if err := utils.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (s *UserService) List() ([]models.User, error) {
	var users []models.User
	if err := utils.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *UserService) Update(user *models.User) error {
	var existing models.User
	if err := utils.DB.First(&existing, user.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found")
		}
		return err
	}
	return utils.DB.Save(user).Error
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
