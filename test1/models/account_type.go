package models

type AccountType struct {
	BaseModel
	Name        string `gorm:"uniqueIndex;size:100;not null" json:"name"`
	Code        string `gorm:"uniqueIndex;size:100;not null" json:"code"`
	Purpose     string `gorm:"size:100;not null" json:"purpose"`
	Description string `gorm:"size:500" json:"description"`
	RoleCode    string `gorm:"size:50;not null" json:"role_code"`
	Status      int    `gorm:"default:1" json:"status"`
}

func (AccountType) TableName() string {
	return "account_types"
}
