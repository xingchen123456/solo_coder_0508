package http

import (
	"context"
	"fmt"
	"net/http"

	"game_backend/internal/config"
	"game_backend/internal/models"
	"game_backend/internal/mysql"
	"game_backend/internal/redis"
	"game_backend/internal/utils"
	"game_backend/internal/websocket"
	proto "game_backend/proto/game"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Server struct {
	engine    *gin.Engine
	config    *config.ServerConfig
	wsManager *websocket.Manager
}

func NewServer(cfg *config.ServerConfig, wsManager *websocket.Manager) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	server := &Server{
		engine:    engine,
		config:    cfg,
		wsManager: wsManager,
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	api := s.engine.Group("/api/v1")

	api.GET("/health", s.healthCheck)

	player := api.Group("/player")
	{
		player.POST("/login", s.login)
		player.POST("/register", s.register)
		player.GET("/info", s.getPlayerInfo)
		player.PUT("/info", s.updatePlayerInfo)
	}

	room := api.Group("/room")
	{
		room.POST("/create", s.createRoom)
		room.POST("/join", s.joinRoom)
		room.POST("/leave", s.leaveRoom)
		room.GET("/list", s.getRoomList)
		room.GET("/info", s.getRoomInfo)
	}

	s.engine.GET("/ws", s.handleWebSocket)
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.HTTPPort)
	zap.L().Info("HTTP server starting", zap.Int("port", s.config.HTTPPort))
	return s.engine.Run(addr)
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (s *Server) login(c *gin.Context) {
	var req proto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var user models.User
	if err := mysql.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	if !utils.VerifyPassword(req.Password, user.Password) {
		c.JSON(http.StatusOK, gin.H{"code": 401, "message": "密码错误"})
		return
	}

	token := utils.GenerateToken(user.ID)
	tokenKey := utils.GetUserTokenKey(user.ID)
	redis.RDB.Set(context.Background(), tokenKey, token, 24*3600)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"token":   token,
		"player": gin.H{
			"user_id":  user.ID,
			"username": user.Username,
			"level":    user.Level,
			"exp":      user.Exp,
			"gold":     user.Gold,
			"diamond":  user.Diamond,
		},
	})
}

func (s *Server) register(c *gin.Context) {
	var req proto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var existingUser models.User
	if err := mysql.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"code": 409, "message": "用户名已存在"})
		return
	}

	user := models.User{
		Username: req.Username,
		Password: utils.HashPassword(req.Password),
		Email:    req.Email,
	}

	if err := mysql.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "注册失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "注册成功",
		"user_id": user.ID,
	})
}

func (s *Server) getPlayerInfo(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	userID, err := s.validateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Token无效"})
		return
	}

	var user models.User
	if err := mysql.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"player": gin.H{
			"user_id":  user.ID,
			"username": user.Username,
			"level":    user.Level,
			"exp":      user.Exp,
			"gold":     user.Gold,
			"diamond":  user.Diamond,
		},
	})
}

func (s *Server) updatePlayerInfo(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	userID, err := s.validateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Token无效"})
		return
	}

	var req struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	updates := map[string]interface{}{}
	if req.Username != "" {
		updates["username"] = req.Username
	}

	if len(updates) > 0 {
		mysql.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

func (s *Server) createRoom(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	userID, err := s.validateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Token无效"})
		return
	}

	var req struct {
		RoomName   string `json:"room_name"`
		RoomType   int32  `json:"room_type"`
		MaxPlayers int32  `json:"max_players"`
		Password   string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	roomID := "ROOM_" + utils.GenerateRandomString(12)

	room := models.Room{
		RoomID:         roomID,
		RoomName:       req.RoomName,
		RoomType:       req.RoomType,
		Status:         1,
		MaxPlayers:     req.MaxPlayers,
		CurrentPlayers: 1,
		OwnerID:        userID,
		Password:       req.Password,
	}

	tx := mysql.DB.Begin()
	if err := tx.Create(&room).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	roomPlayer := models.RoomPlayer{
		RoomID: roomID,
		UserID: userID,
	}
	if err := tx.Create(&roomPlayer).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "创建失败"})
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"room": gin.H{
			"room_id":         roomID,
			"room_name":       req.RoomName,
			"room_type":       req.RoomType,
			"status":          1,
			"max_players":     req.MaxPlayers,
			"current_players": 1,
			"owner_id":        userID,
			"player_ids":      []int64{userID},
		},
	})
}

func (s *Server) joinRoom(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	userID, err := s.validateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Token无效"})
		return
	}

	var req struct {
		RoomID   string `json:"room_id"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var room models.Room
	if err := mysql.DB.Where("room_id = ?", req.RoomID).First(&room).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "房间不存在"})
		return
	}

	if room.Status != 1 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "房间不在等待状态"})
		return
	}

	if room.CurrentPlayers >= room.MaxPlayers {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "房间已满"})
		return
	}

	if room.Password != "" && room.Password != req.Password {
		c.JSON(http.StatusOK, gin.H{"code": 403, "message": "密码错误"})
		return
	}

	var existing models.RoomPlayer
	if err := mysql.DB.Where("room_id = ? AND user_id = ?", req.RoomID, userID).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "已在房间中"})
		return
	}

	tx := mysql.DB.Begin()
	if err := tx.Create(&models.RoomPlayer{RoomID: req.RoomID, UserID: userID}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "加入失败"})
		return
	}
	tx.Model(&room).Update("current_players", room.CurrentPlayers+1)
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "加入成功"})
}

func (s *Server) leaveRoom(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未授权"})
		return
	}

	userID, err := s.validateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "Token无效"})
		return
	}

	var req struct {
		RoomID string `json:"room_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var room models.Room
	if err := mysql.DB.Where("room_id = ?", req.RoomID).First(&room).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "房间不存在"})
		return
	}

	var rp models.RoomPlayer
	if err := mysql.DB.Where("room_id = ? AND user_id = ?", req.RoomID, userID).First(&rp).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "不在房间中"})
		return
	}

	tx := mysql.DB.Begin()
	tx.Delete(&rp)
	newCount := room.CurrentPlayers - 1
	if newCount <= 0 {
		tx.Delete(&room)
	} else {
		tx.Model(&room).Update("current_players", newCount)
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "离开成功"})
}

func (s *Server) getRoomList(c *gin.Context) {
	page := 1
	pageSize := 20

	var rooms []models.Room
	var total int64

	mysql.DB.Model(&models.Room{}).Where("status = ?", 1).Count(&total)
	mysql.DB.Where("status = ?", 1).Offset((page - 1) * pageSize).Limit(pageSize).Find(&rooms)

	roomList := make([]map[string]interface{}, 0, len(rooms))
	for _, room := range rooms {
		roomList = append(roomList, map[string]interface{}{
			"room_id":         room.RoomID,
			"room_name":       room.RoomName,
			"room_type":       room.RoomType,
			"status":          room.Status,
			"max_players":     room.MaxPlayers,
			"current_players": room.CurrentPlayers,
			"owner_id":        room.OwnerID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"rooms":   roomList,
		"total":   total,
	})
}

func (s *Server) getRoomInfo(c *gin.Context) {
	roomID := c.Query("room_id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var room models.Room
	if err := mysql.DB.Where("room_id = ?", roomID).First(&room).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "房间不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"room": gin.H{
			"room_id":         room.RoomID,
			"room_name":       room.RoomName,
			"room_type":       room.RoomType,
			"status":          room.Status,
			"max_players":     room.MaxPlayers,
			"current_players": room.CurrentPlayers,
			"owner_id":        room.OwnerID,
		},
	})
}

func (s *Server) handleWebSocket(c *gin.Context) {
	token := c.Query("token")
	roomID := c.Query("room_id")

	var userID int64 = 0
	if token != "" {
		uid, err := s.validateToken(token)
		if err == nil {
			userID = uid
		}
	}

	conn, err := websocket.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.L().Error("WebSocket upgrade error", zap.Error(err))
		return
	}

	client := &websocket.Client{
		UserID:  userID,
		Token:   token,
		RoomID:  roomID,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Manager: s.wsManager,
	}

	s.wsManager.Register <- client

	if roomID != "" {
		s.wsManager.JoinRoom(client, roomID)
	}

	go client.WritePump()
	go client.ReadPump()
}

func (s *Server) validateToken(token string) (int64, error) {
	ctx := context.Background()

	var user models.User
	result := mysql.DB.Model(&models.User{}).
		Select("id").
		Where("username IN (SELECT username FROM users WHERE CONCAT(id, '-') LIKE ?)", token[0:10]+"%").
		Scan(&user)

	if result.Error != nil || user.ID == 0 {
		return 0, fmt.Errorf("invalid token")
	}

	tokenKey := utils.GetUserTokenKey(user.ID)
	storedToken, err := redis.RDB.Get(ctx, tokenKey).Result()
	if err != nil || storedToken != token {
		return 0, fmt.Errorf("token expired")
	}

	return user.ID, nil
}
