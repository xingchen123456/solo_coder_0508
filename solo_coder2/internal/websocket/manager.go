package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"game_backend/internal/redis"
	"game_backend/internal/utils"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	UserID  int64
	Token   string
	RoomID  string
	Conn    *websocket.Conn
	Send    chan []byte
	Manager *Manager
}

type Manager struct {
	clients    map[*Client]bool
	rooms      map[string]map[*Client]bool
	userRooms  map[int64]string
	broadcast  chan *gamepb.WSMessage
	Register   chan *Client
	Unregister chan *Client
	mutex      sync.RWMutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func NewManager() *Manager {
	return &Manager{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		userRooms:  make(map[int64]string),
		broadcast:  make(chan *gamepb.WSMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (m *Manager) Run() {
	for {
		select {
		case client := <-m.Register:
			m.mutex.Lock()
			m.clients[client] = true
			m.mutex.Unlock()
			zap.L().Info("Client registered", zap.Int64("user_id", client.UserID))

		case client := <-m.Unregister:
			m.mutex.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				close(client.Send)
				if client.RoomID != "" {
					if room, ok := m.rooms[client.RoomID]; ok {
						delete(room, client)
						if len(room) == 0 {
							delete(m.rooms, client.RoomID)
						}
					}
					delete(m.userRooms, client.UserID)
				}
			}
			m.mutex.Unlock()
			zap.L().Info("Client unregistered", zap.Int64("user_id", client.UserID))
		}
	}
}

func (m *Manager) BroadcastToRoom(roomID string, msg *gamepb.WSMessage) {
	data, err := proto.Marshal(msg)
	if err != nil {
		zap.L().Error("Failed to marshal message", zap.Error(err))
		return
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if clients, ok := m.rooms[roomID]; ok {
		for client := range clients {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(clients, client)
			}
		}
	}
}

func (m *Manager) Broadcast(msg *gamepb.WSMessage) {
	data, err := proto.Marshal(msg)
	if err != nil {
		zap.L().Error("Failed to marshal message", zap.Error(err))
		return
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for client := range m.clients {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			delete(m.clients, client)
		}
	}
}

func (m *Manager) JoinRoom(client *Client, roomID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if client.RoomID != "" {
		if oldRoom, ok := m.rooms[client.RoomID]; ok {
			delete(oldRoom, client)
			if len(oldRoom) == 0 {
				delete(m.rooms, client.RoomID)
			}
		}
	}

	if _, ok := m.rooms[roomID]; !ok {
		m.rooms[roomID] = make(map[*Client]bool)
	}

	client.RoomID = roomID
	m.rooms[roomID][client] = true
	m.userRooms[client.UserID] = roomID

	zap.L().Info("Client joined room",
		zap.Int64("user_id", client.UserID),
		zap.String("room_id", roomID),
	)
}

func (m *Manager) LeaveRoom(client *Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if client.RoomID != "" {
		if room, ok := m.rooms[client.RoomID]; ok {
			delete(room, client)
			if len(room) == 0 {
				delete(m.rooms, client.RoomID)
			}
		}
		delete(m.userRooms, client.UserID)
		client.RoomID = ""
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Manager.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(1024 * 1024)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				zap.L().Error("WebSocket error", zap.Error(err))
			}
			break
		}

		c.handleMessage(message)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(data []byte) {
	var wsReq gamepb.WSRequest
	if err := proto.Unmarshal(data, &wsReq); err != nil {
		zap.L().Warn("Failed to unmarshal request", zap.Error(err))
		c.sendResponse(&gamepb.WSResponse{
			Code:    400,
			Message: "无效的消息格式",
		})
		return
	}

	switch wsReq.Type {
	case gamepb.MessageType_MESSAGE_TYPE_HEARTBEAT:
		c.handleHeartbeat(&wsReq)
	case gamepb.MessageType_MESSAGE_TYPE_CHAT:
		c.handleChat(&wsReq)
	case gamepb.MessageType_MESSAGE_TYPE_ROOM_NOTICE:
	case gamepb.MessageType_MESSAGE_TYPE_GAME_ACTION:
	case gamepb.MessageType_MESSAGE_TYPE_SYSTEM_NOTICE:
	default:
		c.sendResponse(&gamepb.WSResponse{
			Code:      400,
			Message:   "未知的消息类型",
			RequestId: wsReq.RequestId,
		})
	}
}

func (c *Client) handleHeartbeat(req *gamepb.WSRequest) {
	var heartbeatReq gamepb.HeartbeatRequest
	if err := proto.Unmarshal(req.Data, &heartbeatReq); err != nil {
		zap.L().Warn("Failed to unmarshal heartbeat", zap.Error(err))
		return
	}

	heartbeatResp := &gamepb.HeartbeatResponse{
		ServerTimestamp: time.Now().Unix(),
	}

	respData, _ := proto.Marshal(heartbeatResp)

	c.sendResponse(&gamepb.WSResponse{
		Type:      gamepb.MessageType_MESSAGE_TYPE_HEARTBEAT,
		Code:      200,
		Message:   "pong",
		Data:      respData,
		RequestId: req.RequestId,
	})
}

func (c *Client) handleChat(req *gamepb.WSRequest) {
	var chatReq gamepb.SendChatRequest
	if err := proto.Unmarshal(req.Data, &chatReq); err != nil {
		c.sendResponse(&gamepb.WSResponse{
			Code:      400,
			Message:   "无效的聊天消息",
			RequestId: req.RequestId,
		})
		return
	}

	if c.RoomID == "" {
		c.sendResponse(&gamepb.WSResponse{
			Code:      400,
			Message:   "未加入房间",
			RequestId: req.RequestId,
		})
		return
	}

	chatMsg := &gamepb.ChatMessage{
		UserId:    c.UserID,
		RoomId:    c.RoomID,
		Content:   chatReq.Content,
		Timestamp: time.Now().Unix(),
	}

	chatData, _ := proto.Marshal(chatMsg)

	wsMsg := &gamepb.WSMessage{
		Type:      gamepb.MessageType_MESSAGE_TYPE_CHAT,
		Data:      chatData,
		RequestId: req.RequestId,
	}

	c.Manager.BroadcastToRoom(c.RoomID, wsMsg)

	c.sendResponse(&gamepb.WSResponse{
		Code:      200,
		Message:   "发送成功",
		RequestId: req.RequestId,
	})
}

func (c *Client) sendResponse(resp *gamepb.WSResponse) {
	resp.Timestamp = time.Now().Unix()
	data, err := proto.Marshal(resp)
	if err != nil {
		zap.L().Error("Failed to marshal response", zap.Error(err))
		return
	}

	select {
	case c.Send <- data:
	default:
	}
}

func ValidateToken(token string) (int64, error) {
	if token == "" {
		return 0, nil
	}

	ctx := context.Background()

	for i := int64(1); i <= 10000; i++ {
		tokenKey := utils.GetUserTokenKey(i)
		storedToken, err := redis.RDB.Get(ctx, tokenKey).Result()
		if err == nil && storedToken == token {
			return i, nil
		}
	}

	return 0, nil
}

func MarshalJSON(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
