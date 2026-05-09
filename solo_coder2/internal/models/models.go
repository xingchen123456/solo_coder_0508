package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Email     string         `gorm:"uniqueIndex;size:128" json:"email"`
	Level     int32          `gorm:"default:1" json:"level"`
	Exp       int64          `gorm:"default:0" json:"exp"`
	Gold      int64          `gorm:"default:0" json:"gold"`
	Diamond   int64          `gorm:"default:0" json:"diamond"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Room struct {
	ID             int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	RoomID         string         `gorm:"uniqueIndex;size:64;not null" json:"room_id"`
	RoomName       string         `gorm:"size:64;not null" json:"room_name"`
	RoomType       int32          `gorm:"default:1" json:"room_type"`
	Status         int32          `gorm:"default:1" json:"status"`
	MaxPlayers     int32          `gorm:"default:2" json:"max_players"`
	CurrentPlayers int32          `gorm:"default:0" json:"current_players"`
	OwnerID        int64          `gorm:"not null" json:"owner_id"`
	Password       string         `gorm:"size:64" json:"-"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type RoomPlayer struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	RoomID    string         `gorm:"index;size:64;not null" json:"room_id"`
	UserID    int64          `gorm:"index;not null" json:"user_id"`
	JoinedAt  time.Time      `json:"joined_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
