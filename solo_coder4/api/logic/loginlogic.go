package logic

import (
	"context"
	"errors"

	"solo-coder4/api/svc"
	"solo-coder4/api/types"
	"solo-coder4/common/db"
	"solo-coder4/common/utils"

	"gorm.io/gorm"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (*types.LoginResponse, error) {
	switch req.LoginType {
	case types.LoginTypeAccount:
		return l.loginByAccount(req.Username, req.Password)
	case types.LoginTypeWechat:
		return l.loginByWechat(req.Code)
	default:
		return nil, errors.New("invalid login type")
	}
}

func (l *LoginLogic) loginByAccount(username, password string) (*types.LoginResponse, error) {
	if username == "" || password == "" {
		return nil, errors.New("username and password are required")
	}

	var user db.User
	err := l.svcCtx.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if !utils.CheckPassword(password, user.Password) {
		return nil, errors.New("invalid password")
	}

	token, err := utils.GenerateToken(user.ID, l.svcCtx.Config.Jwt)
	if err != nil {
		return nil, err
	}

	return &types.LoginResponse{
		UserId:    user.ID,
		Token:     token,
		Username:  user.Username,
		Avatar:    user.Avatar,
		LoginType: user.LoginType,
	}, nil
}

func (l *LoginLogic) loginByWechat(code string) (*types.LoginResponse, error) {
	if code == "" {
		return nil, errors.New("wechat code is required")
	}

	wxResp, err := utils.GetWechatOpenID(code, l.svcCtx.Config.Wechat)
	if err != nil {
		return nil, err
	}

	openID := wxResp.OpenID
	var user db.User
	err = l.svcCtx.DB.Where("wechat_open_id = ?", openID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			username := "wx_" + openID[len(openID)-8:]
			user = db.User{
				Username:     username,
				Password:     "",
				LoginType:    int(db.LoginTypeWechat),
				WechatOpenID: &openID,
			}
			if err := l.svcCtx.DB.Create(&user).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	token, err := utils.GenerateToken(user.ID, l.svcCtx.Config.Jwt)
	if err != nil {
		return nil, err
	}

	return &types.LoginResponse{
		UserId:    user.ID,
		Token:     token,
		Username:  user.Username,
		Avatar:    user.Avatar,
		LoginType: user.LoginType,
	}, nil
}
