package main

import (
	"flag"
	"fmt"

	"solo-coder4/api/config"
	"solo-coder4/api/handler"
	"solo-coder4/api/middleware"
	"solo-coder4/api/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/api.yaml", "the config file")

// @title 用户认证系统 API
// @version 1.0
// @description 基于 go-zero 框架开发的用户认证系统，支持账号密码登录和微信登录
// @host localhost:8888
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	jwtMiddleware := middleware.NewJwtAuthMiddleware(c.Jwt.Secret)

	server.AddRoute(rest.Route{
		Method:  "POST",
		Path:    "/api/register",
		Handler: handler.RegisterHandler(ctx),
	})

	server.AddRoute(rest.Route{
		Method:  "POST",
		Path:    "/api/login",
		Handler: handler.LoginHandler(ctx),
	})

	server.AddRoute(rest.Route{
		Method:  "GET",
		Path:    "/api/user/info",
		Handler: jwtMiddleware.Handle(handler.UserInfoHandler(ctx)),
	})

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
