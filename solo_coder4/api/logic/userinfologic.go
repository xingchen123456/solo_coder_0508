package logic

import (
	"context"
	"errors"

	"solo-coder4/api/svc"
	"solo-coder4/api/types"
	"solo-coder4/common/db"

	"gorm.io/gorm"
)

type UserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserInfoLogic) GetUserInfo(userId int64) (*types.UserInfoResponse, error) {
	var user db.User
	err := l.svcCtx.DB.Where("id = ?", userId).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &types.UserInfoResponse{
		UserId:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		Avatar:    user.Avatar,
		LoginType: user.LoginType,
	}, nil
}
