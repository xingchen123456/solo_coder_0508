package models

type Permission struct {
	BaseModel
	Name        string `gorm:"size:50;not null" json:"name"`
	Code        string `gorm:"uniqueIndex;size:100;not null" json:"code"`
	Type        string `gorm:"size:20;not null" json:"type"`
	Resource    string `gorm:"size:50;not null" json:"resource"`
	Action      string `gorm:"size:20;not null" json:"action"`
	Path        string `gorm:"size:200" json:"path"`
	Method      string `gorm:"size:10" json:"method"`
	Description string `gorm:"size:200" json:"description"`
	Status      int    `gorm:"default:1" json:"status"`
}

func (Permission) TableName() string {
	return "permissions"
}
