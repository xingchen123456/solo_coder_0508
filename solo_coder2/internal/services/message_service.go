package services

import (
	"context"

	"game_backend/internal/websocket"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type MessageService struct {
	gamepb.UnimplementedMessageServiceServer
	wsManager *websocket.Manager
}

func NewMessageService(wsManager *websocket.Manager) *MessageService {
	return &MessageService{
		wsManager: wsManager,
	}
}

func (s *MessageService) SendChat(ctx context.Context, req *gamepb.SendChatRequest) (*gamepb.SendChatResponse, error) {
	zap.L().Info("SendChat called", zap.String("room_id", req.RoomId))

	if s.wsManager == nil {
		return &gamepb.SendChatResponse{
			Code:    500,
			Message: "WebSocket管理器未初始化",
		}, nil
	}

	msg := &gamepb.ChatMessage{
		RoomId:    req.RoomId,
		Content:   req.Content,
		Timestamp: 0,
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return &gamepb.SendChatResponse{
			Code:    500,
			Message: "消息序列化失败",
		}, nil
	}

	wsMsg := &gamepb.WSMessage{
		Type: gamepb.MessageType_MESSAGE_TYPE_CHAT,
		Data: data,
	}

	s.wsManager.BroadcastToRoom(req.RoomId, wsMsg)

	return &gamepb.SendChatResponse{
		Code:    200,
		Message: "发送成功",
	}, nil
}

func (s *MessageService) Broadcast(ctx context.Context, req *gamepb.BroadcastRequest) (*gamepb.BroadcastResponse, error) {
	zap.L().Info("Broadcast called", zap.String("room_id", req.RoomId))

	if s.wsManager == nil {
		return &gamepb.BroadcastResponse{
			Code:    500,
			Message: "WebSocket管理器未初始化",
		}, nil
	}

	wsMsg := &gamepb.WSMessage{
		Type: req.Type,
		Data: req.Data,
	}

	s.wsManager.BroadcastToRoom(req.RoomId, wsMsg)

	return &gamepb.BroadcastResponse{
		Code:    200,
		Message: "广播成功",
	}, nil
}
