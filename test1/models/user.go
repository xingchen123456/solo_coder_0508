package models

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type User struct {
	BaseModel
	Username string `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password string `gorm:"size:100;not null" json:"-"`
	Nickname string `gorm:"size:50" json:"nickname"`
	Email    string `gorm:"size:100" json:"email"`
	Status   int    `gorm:"default:1" json:"status"`
}

func (User) TableName() string {
	return "users"
}
