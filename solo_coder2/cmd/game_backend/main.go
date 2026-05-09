package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"game_backend/internal/config"
	"game_backend/internal/etcd"
	"game_backend/internal/logger"
	"game_backend/internal/mysql"
	"game_backend/internal/redis"
	"game_backend/internal/services"
	"game_backend/internal/websocket"
	proto "game_backend/proto/game"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	configPath string
	rootCmd    = &cobra.Command{
		Use:   "game_backend",
		Short: "Game backend server",
		Run:   runAll,
	}

	grpcCmd = &cobra.Command{
		Use:   "grpc",
		Short: "Start gRPC server",
		Run:   runGRPC,
	}

	httpCmd = &cobra.Command{
		Use:   "http",
		Short: "Start HTTP gateway server",
		Run:   runHTTP,
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config file path")
	rootCmd.AddCommand(grpcCmd)
	rootCmd.AddCommand(httpCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runAll(cmd *cobra.Command, args []string) {
	initDependencies()

	wsManager := websocket.NewManager()
	go wsManager.Run()

	go func() {
		if err := startGRPCServer(wsManager); err != nil {
			logger.Fatal("Failed to start gRPC server", zap.Error(err))
		}
	}()

	go func() {
		if err := startHTTPServer(wsManager); err != nil {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	waitForSignal()
}

func runGRPC(cmd *cobra.Command, args []string) {
	initDependencies()

	wsManager := websocket.NewManager()
	go wsManager.Run()

	if err := startGRPCServer(wsManager); err != nil {
		logger.Fatal("Failed to start gRPC server", zap.Error(err))
	}
}

func runHTTP(cmd *cobra.Command, args []string) {
	initDependencies()

	wsManager := websocket.NewManager()
	go wsManager.Run()

	if err := startHTTPServer(wsManager); err != nil {
		logger.Fatal("Failed to start HTTP server", zap.Error(err))
	}
}

func initDependencies() {
	config.InitConfig(configPath)
	logger.InitLogger(config.AppConfig.Log.Level, config.AppConfig.Log.Format)

	if _, err := mysql.InitMySQL(&config.AppConfig.MySQL); err != nil {
		logger.Fatal("Failed to initialize MySQL", zap.Error(err))
	}

	if _, err := redis.InitRedis(&config.AppConfig.Redis); err != nil {
		logger.Fatal("Failed to initialize Redis", zap.Error(err))
	}

	if _, err := etcd.InitEtcd(&config.AppConfig.Etcd); err != nil {
		logger.Warn("Failed to initialize Etcd", zap.Error(err))
	}
}

func startGRPCServer(wsManager *websocket.Manager) error {
	port := config.AppConfig.Server.GrpcPort
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()

	playerService := services.NewPlayerService()
	roomService := services.NewRoomService()
	messageService := services.NewMessageService(wsManager)

	proto.RegisterPlayerServiceServer(grpcServer, playerService)
	proto.RegisterRoomServiceServer(grpcServer, roomService)
	proto.RegisterMessageServiceServer(grpcServer, messageService)

	reflection.Register(grpcServer)

	logger.Info("gRPC server starting", zap.Int("port", port))

	if err := registerServiceToEtcd("grpc", port); err != nil {
		logger.Warn("Failed to register to etcd", zap.Error(err))
	}

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func startHTTPServer(wsManager *websocket.Manager) error {
	port := config.AppConfig.Server.HTTPPort
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())

	api := router.Group("/api/v1")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/ws", func(c *gin.Context) {
		handleWebSocket(c, wsManager)
	})

	logger.Info("HTTP server starting", zap.Int("port", port))

	if err := registerServiceToEtcd("http", port); err != nil {
		logger.Warn("Failed to register to etcd", zap.Error(err))
	}

	return router.Run(fmt.Sprintf(":%d", port))
}

func registerServiceToEtcd(serviceType string, port int) error {
	if etcd.Client == nil {
		return nil
	}

	hostname, _ := os.Hostname()
	serviceInfo := &etcd.ServiceInfo{
		ServiceName: fmt.Sprintf("game-%s", serviceType),
		Address:     hostname,
		Port:        port,
	}

	return etcd.RegisterService(serviceInfo.ServiceName, serviceInfo)
}

func waitForSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutting down...")

	mysql.Close()
	redis.Close()
	etcd.Close()
	logger.Sync()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebSocket(c *gin.Context, wsManager *websocket.Manager) {
	token := c.Query("token")
	roomID := c.Query("room_id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket upgrade error", zap.Error(err))
		return
	}

	client := &websocket.Client{
		UserID:  0,
		Token:   token,
		RoomID:  roomID,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Manager: wsManager,
	}

	if token != "" {
		if userID, err := websocket.ValidateToken(token); err == nil && userID > 0 {
			client.UserID = userID
		}
	}

	wsManager.Register <- client

	if roomID != "" {
		wsManager.JoinRoom(client, roomID)
	}

	go client.WritePump()
	go client.ReadPump()
}
