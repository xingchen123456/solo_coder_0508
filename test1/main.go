package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"management-system/config"
	_ "management-system/docs"
	"management-system/models"
	"management-system/routes"
	"management-system/utils"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title 管理系统 API
// @version 1.0
// @description 基于 Gin 的管理系统 API 文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http
func main() {
	if err := config.LoadConfig("config/config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := utils.InitLogger(config.AppConfig.Server.Mode); err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}

	if err := utils.InitMySQL(); err != nil {
		log.Fatalf("Failed to init MySQL: %v", err)
	}
	if err := utils.DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}

	if err := utils.InitRedis(); err != nil {
		log.Fatalf("Failed to init Redis: %v", err)
	}

	if err := utils.InitEtcd(); err != nil {
		log.Fatalf("Failed to init etcd: %v", err)
	}

	gin.SetMode(config.AppConfig.Server.Mode)
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.SetupRoutes(r, utils.Logger)

	addr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	srv := &struct {
		Addr    string
		Handler *gin.Engine
	}{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on %s", addr)
		log.Printf("Swagger UI available at: http://localhost:%d/swagger/index.html", config.AppConfig.Server.Port)
		if err := r.Run(srv.Addr); err != nil {
			log.Printf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = ctx
	log.Println("Server exiting")

	if err := utils.CloseMySQL(); err != nil {
		log.Printf("Error closing MySQL: %v", err)
	}
	if err := utils.CloseRedis(); err != nil {
		log.Printf("Error closing Redis: %v", err)
	}
	if err := utils.CloseEtcd(); err != nil {
		log.Printf("Error closing etcd: %v", err)
	}

	log.Println("All connections closed")
}
