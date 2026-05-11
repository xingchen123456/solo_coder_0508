package main

import (
	"log"

	"solo-coder4/common/db"
	"solo-coder4/common/utils"
	userservice "solo-coder4/rpc/user/service"

	"github.com/zeromicro/go-zero/core/conf"
	coreservice "github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type Config struct {
	zrpc.RpcServerConf
	DB     db.Config
	Jwt    utils.JwtConfig
	Wechat utils.WechatConfig
}

func main() {
	var c Config
	conf.MustLoad("etc/user.yaml", &c)

	dbConn, err := db.InitDB(c.DB)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	userService := userservice.NewUserService(dbConn, c.Jwt, c.Wechat)

	_ = userService

	srv := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
	})

	group := coreservice.NewServiceGroup()
	group.Add(srv)

	defer group.Stop()

	log.Printf("Starting user rpc server at %s...", c.ListenOn)
	group.Start()
}
