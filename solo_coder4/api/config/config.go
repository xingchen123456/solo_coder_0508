package config

import (
	"solo-coder4/common/db"
	"solo-coder4/common/utils"

	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	DB     db.Config
	Jwt    utils.JwtConfig
	Wechat utils.WechatConfig
}
