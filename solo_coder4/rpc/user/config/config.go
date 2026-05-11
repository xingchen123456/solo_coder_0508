package config

import (
	"solo-coder4/common/db"
	"solo-coder4/common/utils"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DB     db.Config
	Jwt    utils.JwtConfig
	Wechat utils.WechatConfig
}
