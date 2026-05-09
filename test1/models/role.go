package models

type Role struct {
	BaseModel
	Name        string       `gorm:"uniqueIndex;size:50;not null" json:"name"`
	Code        string       `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Description string       `gorm:"size:200" json:"description"`
	Status      int          `gorm:"default:1" json:"status"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

func (Role) TableName() string {
	return "roles"
}
