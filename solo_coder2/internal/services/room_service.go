package services

import (
	"context"
	"time"

	"game_backend/internal/models"
	"game_backend/internal/mysql"
	proto "game_backend/proto/game"

	"go.uber.org/zap"
)

type RoomService struct {
	proto.UnimplementedRoomServiceServer
}

func NewRoomService() *RoomService {
	return &RoomService{}
}

func (s *RoomService) CreateRoom(ctx context.Context, req *proto.CreateRoomRequest) (*proto.CreateRoomResponse, error) {
	userID, err := validateToken(ctx, req.Token)
	if err != nil {
		return &proto.CreateRoomResponse{
			Code:    401,
			Message: "Token无效",
		}, nil
	}

	roomID := "ROOM_" + time.Now().Format("20060102150405") + getRandomSuffix()

	room := models.Room{
		RoomID:         roomID,
		RoomName:       req.RoomName,
		RoomType:       int32(req.RoomType),
		Status:         int32(proto.RoomStatus_ROOM_STATUS_WAITING),
		MaxPlayers:     req.MaxPlayers,
		CurrentPlayers: 1,
		OwnerID:        userID,
		Password:       req.Password,
	}

	tx := mysql.DB.Begin()
	if err := tx.Create(&room).Error; err != nil {
		tx.Rollback()
		zap.L().Error("Failed to create room", zap.Error(err))
		return &proto.CreateRoomResponse{
			Code:    500,
			Message: "创建房间失败",
		}, nil
	}

	roomPlayer := models.RoomPlayer{
		RoomID:   roomID,
		UserID:   userID,
		JoinedAt: time.Now(),
	}

	if err := tx.Create(&roomPlayer).Error; err != nil {
		tx.Rollback()
		zap.L().Error("Failed to create room player", zap.Error(err))
		return &proto.CreateRoomResponse{
			Code:    500,
			Message: "创建房间失败",
		}, nil
	}

	tx.Commit()

	return &proto.CreateRoomResponse{
		Code:    200,
		Message: "创建成功",
		Room:    toProtoRoomInfo(&room, []int64{userID}),
	}, nil
}

func (s *RoomService) JoinRoom(ctx context.Context, req *proto.JoinRoomRequest) (*proto.JoinRoomResponse, error) {
	userID, err := validateToken(ctx, req.Token)
	if err != nil {
		return &proto.JoinRoomResponse{
			Code:    401,
			Message: "Token无效",
		}, nil
	}

	var room models.Room
	if err := mysql.DB.Where("room_id = ?", req.RoomId).First(&room).Error; err != nil {
		return &proto.JoinRoomResponse{
			Code:    404,
			Message: "房间不存在",
		}, nil
	}

	if room.Status != int32(proto.RoomStatus_ROOM_STATUS_WAITING) {
		return &proto.JoinRoomResponse{
			Code:    400,
			Message: "房间不在等待状态",
		}, nil
	}

	if room.CurrentPlayers >= room.MaxPlayers {
		return &proto.JoinRoomResponse{
			Code:    400,
			Message: "房间已满",
		}, nil
	}

	if room.Password != "" && room.Password != req.Password {
		return &proto.JoinRoomResponse{
			Code:    403,
			Message: "密码错误",
		}, nil
	}

	var existingPlayer models.RoomPlayer
	if err := mysql.DB.Where("room_id = ? AND user_id = ?", req.RoomId, userID).First(&existingPlayer).Error; err == nil {
		return &proto.JoinRoomResponse{
			Code:    400,
			Message: "已经在房间中",
		}, nil
	}

	tx := mysql.DB.Begin()

	roomPlayer := models.RoomPlayer{
		RoomID:   req.RoomId,
		UserID:   userID,
		JoinedAt: time.Now(),
	}

	if err := tx.Create(&roomPlayer).Error; err != nil {
		tx.Rollback()
		zap.L().Error("Failed to join room", zap.Error(err))
		return &proto.JoinRoomResponse{
			Code:    500,
			Message: "加入房间失败",
		}, nil
	}

	if err := tx.Model(&room).Update("current_players", room.CurrentPlayers+1).Error; err != nil {
		tx.Rollback()
		zap.L().Error("Failed to update room players", zap.Error(err))
		return &proto.JoinRoomResponse{
			Code:    500,
			Message: "加入房间失败",
		}, nil
	}

	tx.Commit()

	room.CurrentPlayers++
	playerIDs := getRoomPlayerIDs(req.RoomId)

	return &proto.JoinRoomResponse{
		Code:    200,
		Message: "加入成功",
		Room:    toProtoRoomInfo(&room, playerIDs),
	}, nil
}

func (s *RoomService) LeaveRoom(ctx context.Context, req *proto.LeaveRoomRequest) (*proto.LeaveRoomResponse, error) {
	userID, err := validateToken(ctx, req.Token)
	if err != nil {
		return &proto.LeaveRoomResponse{
			Code:    401,
			Message: "Token无效",
		}, nil
	}

	var room models.Room
	if err := mysql.DB.Where("room_id = ?", req.RoomId).First(&room).Error; err != nil {
		return &proto.LeaveRoomResponse{
			Code:    404,
			Message: "房间不存在",
		}, nil
	}

	var roomPlayer models.RoomPlayer
	if err := mysql.DB.Where("room_id = ? AND user_id = ?", req.RoomId, userID).First(&roomPlayer).Error; err != nil {
		return &proto.LeaveRoomResponse{
			Code:    400,
			Message: "不在房间中",
		}, nil
	}

	tx := mysql.DB.Begin()

	if err := tx.Delete(&roomPlayer).Error; err != nil {
		tx.Rollback()
		zap.L().Error("Failed to leave room", zap.Error(err))
		return &proto.LeaveRoomResponse{
			Code:    500,
			Message: "离开房间失败",
		}, nil
	}

	newCount := room.CurrentPlayers - 1
	if newCount <= 0 {
		if err := tx.Delete(&room).Error; err != nil {
			tx.Rollback()
			zap.L().Error("Failed to delete room", zap.Error(err))
			return &proto.LeaveRoomResponse{
				Code:    500,
				Message: "离开房间失败",
			}, nil
		}
	} else {
		if err := tx.Model(&room).Update("current_players", newCount).Error; err != nil {
			tx.Rollback()
			zap.L().Error("Failed to update room players", zap.Error(err))
			return &proto.LeaveRoomResponse{
				Code:    500,
				Message: "离开房间失败",
			}, nil
		}

		if room.OwnerID == userID {
			var newOwner models.RoomPlayer
			if err := tx.Where("room_id = ?", req.RoomId).First(&newOwner).Error; err == nil {
				if err := tx.Model(&room).Update("owner_id", newOwner.UserID).Error; err != nil {
					tx.Rollback()
					zap.L().Error("Failed to update room owner", zap.Error(err))
					return &proto.LeaveRoomResponse{
						Code:    500,
						Message: "离开房间失败",
					}, nil
				}
			}
		}
	}

	tx.Commit()

	return &proto.LeaveRoomResponse{
		Code:    200,
		Message: "离开成功",
	}, nil
}

func (s *RoomService) GetRoomList(ctx context.Context, req *proto.GetRoomListRequest) (*proto.GetRoomListResponse, error) {
	offset := (req.Page - 1) * req.PageSize
	if offset < 0 {
		offset = 0
	}

	var rooms []models.Room
	var total int64

	query := mysql.DB.Model(&models.Room{}).Where("status = ?", proto.RoomStatus_ROOM_STATUS_WAITING)
	if req.RoomType != proto.RoomType_ROOM_TYPE_UNKNOWN {
		query = query.Where("room_type = ?", req.RoomType)
	}

	if err := query.Count(&total).Error; err != nil {
		zap.L().Error("Failed to count rooms", zap.Error(err))
		return &proto.GetRoomListResponse{
			Code:    500,
			Message: "获取房间列表失败",
		}, nil
	}

	if err := query.Offset(int(offset)).Limit(int(req.PageSize)).Find(&rooms).Error; err != nil {
		zap.L().Error("Failed to get rooms", zap.Error(err))
		return &proto.GetRoomListResponse{
			Code:    500,
			Message: "获取房间列表失败",
		}, nil
	}

	protoRooms := make([]*proto.RoomInfo, 0, len(rooms))
	for _, room := range rooms {
		playerIDs := getRoomPlayerIDs(room.RoomID)
		protoRooms = append(protoRooms, toProtoRoomInfo(&room, playerIDs))
	}

	return &proto.GetRoomListResponse{
		Code:    200,
		Message: "成功",
		Rooms:   protoRooms,
		Total:   int32(total),
	}, nil
}

func (s *RoomService) GetRoomInfo(ctx context.Context, req *proto.GetRoomInfoRequest) (*proto.GetRoomInfoResponse, error) {
	var room models.Room
	if err := mysql.DB.Where("room_id = ?", req.RoomId).First(&room).Error; err != nil {
		return &proto.GetRoomInfoResponse{
			Code:    404,
			Message: "房间不存在",
		}, nil
	}

	playerIDs := getRoomPlayerIDs(req.RoomId)

	return &proto.GetRoomInfoResponse{
		Code:    200,
		Message: "成功",
		Room:    toProtoRoomInfo(&room, playerIDs),
	}, nil
}

func toProtoRoomInfo(room *models.Room, playerIDs []int64) *proto.RoomInfo {
	return &proto.RoomInfo{
		RoomId:         room.RoomID,
		RoomName:       room.RoomName,
		RoomType:       proto.RoomType(room.RoomType),
		Status:         proto.RoomStatus(room.Status),
		MaxPlayers:     room.MaxPlayers,
		CurrentPlayers: room.CurrentPlayers,
		OwnerId:        room.OwnerID,
		PlayerIds:      playerIDs,
		CreateTime:     room.CreatedAt.Unix(),
	}
}

func getRoomPlayerIDs(roomID string) []int64 {
	var players []models.RoomPlayer
	playerIDs := []int64{}
	if err := mysql.DB.Where("room_id = ?", roomID).Find(&players).Error; err == nil {
		for _, p := range players {
			playerIDs = append(playerIDs, p.UserID)
		}
	}
	return playerIDs
}

func getRandomSuffix() string {
	const letters = "0123456789"
	b := make([]byte, 4)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
