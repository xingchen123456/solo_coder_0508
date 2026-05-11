package logic

import (
	"context"
	"errors"

	"solo-coder4/api/svc"
	"solo-coder4/api/types"
	"solo-coder4/common/db"
	"solo-coder4/common/utils"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (*types.RegisterResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, errors.New("username and password are required")
	}

	var count int64
	l.svcCtx.DB.Model(&db.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, errors.New("username already exists")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &db.User{
		Username:  req.Username,
		Password:  hashedPassword,
		Email:     req.Email,
		Phone:     req.Phone,
		LoginType: int(db.LoginTypeAccount),
	}

	if err := l.svcCtx.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return &types.RegisterResponse{
		UserId:    user.ID,
		Username:  user.Username,
		LoginType: user.LoginType,
		Message:   "Registration successful",
	}, nil
}
