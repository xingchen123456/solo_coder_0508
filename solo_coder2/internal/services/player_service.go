package services

import (
	"context"
	"errors"
	"time"

	"game_backend/internal/models"
	"game_backend/internal/mysql"
	"game_backend/internal/redis"
	"game_backend/internal/utils"
	proto "game_backend/proto/game"

	"go.uber.org/zap"
)

type PlayerService struct {
	proto.UnimplementedPlayerServiceServer
}

func NewPlayerService() *PlayerService {
	return &PlayerService{}
}

func (s *PlayerService) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	var user models.User
	if err := mysql.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		zap.L().Warn("User not found", zap.String("username", req.Username))
		return &proto.LoginResponse{
			Code:    404,
			Message: "用户不存在",
		}, nil
	}

	if !utils.VerifyPassword(req.Password, user.Password) {
		zap.L().Warn("Invalid password", zap.String("username", req.Username))
		return &proto.LoginResponse{
			Code:    401,
			Message: "密码错误",
		}, nil
	}

	token := utils.GenerateToken(user.ID)
	tokenKey := utils.GetUserTokenKey(user.ID)

	err := redis.RDB.Set(ctx, tokenKey, token, 24*time.Hour).Err()
	if err != nil {
		zap.L().Error("Failed to set token to redis", zap.Error(err))
		return &proto.LoginResponse{
			Code:    500,
			Message: "服务器错误",
		}, nil
	}

	return &proto.LoginResponse{
		Code:    200,
		Message: "登录成功",
		Token:   token,
		Player:  toProtoPlayerInfo(&user),
	}, nil
}

func (s *PlayerService) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	var existingUser models.User
	if err := mysql.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return &proto.RegisterResponse{
			Code:    409,
			Message: "用户名已存在",
		}, nil
	}

	user := models.User{
		Username: req.Username,
		Password: utils.HashPassword(req.Password),
		Email:    req.Email,
	}

	if err := mysql.DB.Create(&user).Error; err != nil {
		zap.L().Error("Failed to create user", zap.Error(err))
		return &proto.RegisterResponse{
			Code:    500,
			Message: "注册失败",
		}, nil
	}

	return &proto.RegisterResponse{
		Code:    200,
		Message: "注册成功",
		UserId:  user.ID,
	}, nil
}

func (s *PlayerService) GetPlayerInfo(ctx context.Context, req *proto.GetPlayerInfoRequest) (*proto.GetPlayerInfoResponse, error) {
	userID, err := validateToken(ctx, req.Token)
	if err != nil {
		return &proto.GetPlayerInfoResponse{
			Code:    401,
			Message: "Token无效",
		}, nil
	}

	var user models.User
	if err := mysql.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return &proto.GetPlayerInfoResponse{
			Code:    404,
			Message: "用户不存在",
		}, nil
	}

	return &proto.GetPlayerInfoResponse{
		Code:    200,
		Message: "成功",
		Player:  toProtoPlayerInfo(&user),
	}, nil
}

func (s *PlayerService) UpdatePlayerInfo(ctx context.Context, req *proto.UpdatePlayerInfoRequest) (*proto.UpdatePlayerInfoResponse, error) {
	userID, err := validateToken(ctx, req.Token)
	if err != nil {
		return &proto.UpdatePlayerInfoResponse{
			Code:    401,
			Message: "Token无效",
		}, nil
	}

	updates := map[string]interface{}{}
	if req.Username != "" {
		updates["username"] = req.Username
	}

	if len(updates) == 0 {
		return &proto.UpdatePlayerInfoResponse{
			Code:    400,
			Message: "没有需要更新的信息",
		}, nil
	}

	if err := mysql.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		zap.L().Error("Failed to update user", zap.Error(err))
		return &proto.UpdatePlayerInfoResponse{
			Code:    500,
			Message: "更新失败",
		}, nil
	}

	return &proto.UpdatePlayerInfoResponse{
		Code:    200,
		Message: "更新成功",
	}, nil
}

func toProtoPlayerInfo(user *models.User) *proto.PlayerInfo {
	return &proto.PlayerInfo{
		UserId:   user.ID,
		Username: user.Username,
		Level:    user.Level,
		Exp:      user.Exp,
		Gold:     user.Gold,
		Diamond:  user.Diamond,
	}
}

func validateToken(ctx context.Context, token string) (int64, error) {
	if token == "" {
		return 0, errors.New("token is empty")
	}

	var userID int64
	result := mysql.DB.Model(&models.User{}).
		Select("id").
		Where("EXISTS (SELECT 1 FROM users WHERE CONCAT(id, '-') IN (SELECT SUBSTRING_INDEX(?, '-', 1)))", token).
		Scan(&userID)

	if result.Error != nil || userID == 0 {
		return 0, errors.New("invalid token")
	}

	tokenKey := utils.GetUserTokenKey(userID)
	storedToken, err := redis.RDB.Get(ctx, tokenKey).Result()
	if err != nil || storedToken != token {
		return 0, errors.New("token expired or invalid")
	}

	return userID, nil
}
