package models

import "time"

type UserRole struct {
	UserID    uint      `gorm:"primaryKey;not null" json:"user_id"`
	RoleID    uint      `gorm:"primaryKey;not null" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserRole) TableName() string {
	return "user_roles"
}
