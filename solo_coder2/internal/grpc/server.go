package grpc

import (
	"fmt"
	"net"

	"game_backend/internal/config"
	"game_backend/internal/services"
	"game_backend/internal/websocket"
	proto "game_backend/proto/game"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer *grpc.Server
	config     *config.ServerConfig
	wsManager  *websocket.Manager
}

func NewServer(cfg *config.ServerConfig, wsManager *websocket.Manager) *Server {
	return &Server{
		config:    cfg,
		wsManager: wsManager,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GrpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.grpcServer = grpc.NewServer()

	playerService := services.NewPlayerService()
	roomService := services.NewRoomService()
	messageService := services.NewMessageService(s.wsManager)

	proto.RegisterPlayerServiceServer(s.grpcServer, playerService)
	proto.RegisterRoomServiceServer(s.grpcServer, roomService)
	proto.RegisterMessageServiceServer(s.grpcServer, messageService)

	reflection.Register(s.grpcServer)

	zap.L().Info("gRPC server starting", zap.Int("port", s.config.GrpcPort))

	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}
