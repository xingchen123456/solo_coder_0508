package models

import "time"

type RolePermission struct {
	RoleID       uint      `gorm:"primaryKey;not null" json:"role_id"`
	PermissionID uint      `gorm:"primaryKey;not null" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
