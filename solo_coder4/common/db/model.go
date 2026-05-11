package db

import (
	"time"

	"gorm.io/gorm"
)

type LoginType int

const (
	LoginTypeUnknown LoginType = iota
	LoginTypeAccount
	LoginTypeWechat
)

type User struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password     string         `gorm:"size:255" json:"-"`
	Email        string         `gorm:"size:100" json:"email"`
	Phone        string         `gorm:"size:20" json:"phone"`
	Avatar       string         `gorm:"size:255" json:"avatar"`
	LoginType    int            `gorm:"not null;default:1" json:"loginType"`
	WechatOpenID *string        `gorm:"uniqueIndex;size:100" json:"wechatOpenId"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}
