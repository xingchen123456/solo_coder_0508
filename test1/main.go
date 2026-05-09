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

func InitSeedData() error {
	if err := initRoles(); err != nil {
		return err
	}
	if err := initUsers(); err != nil {
		return err
	}
	if err := initAccountTypes(); err != nil {
		return err
	}
	log.Println("Seed data initialized successfully")
	return nil
}

func initRoles() error {
	roles := []models.Role{
		{
			Name:        "平台",
			Code:        "platform",
			Description: "平台管理员，拥有系统最高权限",
			Status:      1,
		},
		{
			Name:        "店主",
			Code:        "shop_owner",
			Description: "店铺店主，管理自己的店铺和订单",
			Status:      1,
		},
		{
			Name:        "运营",
			Code:        "operator",
			Description: "运营人员，负责平台运营工作",
			Status:      1,
		},
	}

	for _, role := range roles {
		var existing models.Role
		result := utils.DB.Where("code = ?", role.Code).First(&existing)
		if result.Error != nil {
			if result := utils.DB.Create(&role); result.Error != nil {
				return result.Error
			}
			log.Printf("Role created: %s", role.Code)
		}
	}

	return nil
}

func initUsers() error {
	var platformRole, shopOwnerRole, operatorRole models.Role
	utils.DB.Where("code = ?", "platform").First(&platformRole)
	utils.DB.Where("code = ?", "shop_owner").First(&shopOwnerRole)
	utils.DB.Where("code = ?", "operator").First(&operatorRole)

	users := []struct {
		user   models.User
		roleID uint
	}{
		{
			user: models.User{
				Username: "platform",
				Password: "platform123",
				Nickname: "平台管理员",
				Email:    "platform@example.com",
				Status:   1,
			},
			roleID: platformRole.ID,
		},
		{
			user: models.User{
				Username: "shopowner",
				Password: "shopowner123",
				Nickname: "店主用户",
				Email:    "shopowner@example.com",
				Status:   1,
			},
			roleID: shopOwnerRole.ID,
		},
		{
			user: models.User{
				Username: "operator",
				Password: "operator123",
				Nickname: "运营人员",
				Email:    "operator@example.com",
				Status:   1,
			},
			roleID: operatorRole.ID,
		},
	}

	for _, item := range users {
		var existing models.User
		result := utils.DB.Where("username = ?", item.user.Username).First(&existing)
		if result.Error != nil {
			if result := utils.DB.Create(&item.user); result.Error != nil {
				return result.Error
			}
			userRole := &models.UserRole{
				UserID: item.user.ID,
				RoleID: item.roleID,
			}
			if result := utils.DB.Create(userRole); result.Error != nil {
				return result.Error
			}
			log.Printf("User created: %s with role ID: %d", item.user.Username, item.roleID)
		}
	}

	return nil
}

func initAccountTypes() error {
	accountTypes := []models.AccountType{
		{
			Name:        "我的佣金（平台）",
			Code:        "platform_commission",
			Purpose:     "收取佣金",
			Description: "结算后平台应得的佣金部分",
			RoleCode:    "platform",
			Status:      1,
		},
		{
			Name:        "罚补账户",
			Code:        "penalty_compensation",
			Purpose:     "罚款/补贴",
			Description: "用于违约罚款和平台补贴，针对店主和运营保证金",
			RoleCode:    "platform",
			Status:      1,
		},
		{
			Name:        "托管账户（临时）",
			Code:        "escrow_temporary",
			Purpose:     "订单托管",
			Description: "锁定店主预付款，待结算时分账（合规）",
			RoleCode:    "platform",
			Status:      1,
		},
		{
			Name:        "我的预付款",
			Code:        "shop_advance",
			Purpose:     "订单托管",
			Description: "用于锁定订单金额，保障运营发货后能收到货款",
			RoleCode:    "shop_owner",
			Status:      1,
		},
		{
			Name:        "我的佣金",
			Code:        "shop_commission",
			Purpose:     "收取佣金",
			Description: "结算后店主应得的佣金部分，可全额提现",
			RoleCode:    "shop_owner",
			Status:      1,
		},
		{
			Name:        "店主（合作）保证金",
			Code:        "shop_guarantee",
			Purpose:     "合作违约",
			Description: "仅用于店主违约时扣除，最低门槛15000台币",
			RoleCode:    "shop_owner",
			Status:      1,
		},
		{
			Name:        "我的回款",
			Code:        "operator_receivable",
			Purpose:     "收取回款",
			Description: "结算后运营应得的货款部分，可全额提现",
			RoleCode:    "operator",
			Status:      1,
		},
		{
			Name:        "运营（合作）保证金",
			Code:        "operator_guarantee",
			Purpose:     "合作违约",
			Description: "字段保留，暂不要求",
			RoleCode:    "operator",
			Status:      1,
		},
	}

	for _, accountType := range accountTypes {
		var existing models.AccountType
		result := utils.DB.Where("code = ?", accountType.Code).First(&existing)
		if result.Error != nil {
			if result := utils.DB.Create(&accountType); result.Error != nil {
				return result.Error
			}
			log.Printf("AccountType created: %s", accountType.Code)
		}
	}

	return nil
}

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
	if err := utils.DB.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.AccountType{},
	); err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}

	if err := InitSeedData(); err != nil {
		log.Fatalf("Failed to init seed data: %v", err)
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
