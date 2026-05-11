package service

import (
	"context"
	"errors"

	"solo-coder4/common/db"
	"solo-coder4/common/types"
	"solo-coder4/common/utils"

	"gorm.io/gorm"
)

type UserService struct {
	DB         *gorm.DB
	JwtConfig  utils.JwtConfig
	WechatConf utils.WechatConfig
}

func NewUserService(dbConn *gorm.DB, jwtCfg utils.JwtConfig, wechatCfg utils.WechatConfig) *UserService {
	return &UserService{
		DB:         dbConn,
		JwtConfig:  jwtCfg,
		WechatConf: wechatCfg,
	}
}

func (s *UserService) Register(ctx context.Context, req *types.RegisterRequest) error {
	var count int64
	s.DB.Model(&db.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return errors.New("username already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}

	user := &db.User{
		Username:  req.Username,
		Password:  hashedPassword,
		Email:     req.Email,
		Phone:     req.Phone,
		LoginType: int(db.LoginTypeAccount),
	}

	return s.DB.Create(user).Error
}

func (s *UserService) LoginByAccount(ctx context.Context, username, password string) (*types.LoginResponse, error) {
	var user db.User
	err := s.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if !utils.CheckPassword(password, user.Password) {
		return nil, errors.New("invalid password")
	}

	token, err := utils.GenerateToken(user.ID, s.JwtConfig)
	if err != nil {
		return nil, err
	}

	return &types.LoginResponse{
		UserId:   user.ID,
		Token:    token,
		Username: user.Username,
		Avatar:   user.Avatar,
	}, nil
}

func (s *UserService) LoginByWechat(ctx context.Context, code string) (*types.LoginResponse, error) {
	wxResp, err := utils.GetWechatOpenID(code, s.WechatConf)
	if err != nil {
		return nil, err
	}

	openID := wxResp.OpenID
	var user db.User
	err = s.DB.Where("wechat_open_id = ?", openID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = db.User{
				Username:     "wx_" + openID[len(openID)-8:],
				Password:     "",
				LoginType:    int(db.LoginTypeWechat),
				WechatOpenID: &openID,
			}
			if err := s.DB.Create(&user).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	token, err := utils.GenerateToken(user.ID, s.JwtConfig)
	if err != nil {
		return nil, err
	}

	return &types.LoginResponse{
		UserId:   user.ID,
		Token:    token,
		Username: user.Username,
		Avatar:   user.Avatar,
	}, nil
}

func (s *UserService) GetUserInfo(ctx context.Context, userId int64) (*types.UserInfo, error) {
	var user db.User
	err := s.DB.Where("id = ?", userId).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &types.UserInfo{
		UserId:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Phone:    user.Phone,
		Avatar:   user.Avatar,
	}, nil
}
