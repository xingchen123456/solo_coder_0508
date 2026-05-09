package services

import (
	"context"

	"game_backend/internal/websocket"
	proto "game_backend/proto/game"

	"go.uber.org/zap"
)

type MessageService struct {
	proto.UnimplementedMessageServiceServer
	wsManager *websocket.Manager
}

func NewMessageService(wsManager *websocket.Manager) *MessageService {
	return &MessageService{
		wsManager: wsManager,
	}
}

func (s *MessageService) SendChat(ctx context.Context, req *proto.SendChatRequest) (*proto.SendChatResponse, error) {
	zap.L().Info("SendChat called", zap.String("room_id", req.RoomId))

	if s.wsManager == nil {
		return &proto.SendChatResponse{
			Code:    500,
			Message: "WebSocket管理器未初始化",
		}, nil
	}

	msg := &proto.ChatMessage{
		RoomId:    req.RoomId,
		Content:   req.Content,
		Timestamp: 0,
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return &proto.SendChatResponse{
			Code:    500,
			Message: "消息序列化失败",
		}, nil
	}

	wsMsg := &proto.WSMessage{
		Type: proto.MessageType_MESSAGE_TYPE_CHAT,
		Data: data,
	}

	s.wsManager.BroadcastToRoom(req.RoomId, wsMsg)

	return &proto.SendChatResponse{
		Code:    200,
		Message: "发送成功",
	}, nil
}

func (s *MessageService) Broadcast(ctx context.Context, req *proto.BroadcastRequest) (*proto.BroadcastResponse, error) {
	zap.L().Info("Broadcast called", zap.String("room_id", req.RoomId))

	if s.wsManager == nil {
		return &proto.BroadcastResponse{
			Code:    500,
			Message: "WebSocket管理器未初始化",
		}, nil
	}

	wsMsg := &proto.WSMessage{
		Type: req.Type,
		Data: req.Data,
	}

	s.wsManager.BroadcastToRoom(req.RoomId, wsMsg)

	return &proto.BroadcastResponse{
		Code:    200,
		Message: "广播成功",
	}, nil
}
