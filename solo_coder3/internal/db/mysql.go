package db

import (
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"inventory-service/config"
)

var DB *gorm.DB

type Inventory struct {
	ID        uint   `gorm:"primaryKey"`
	ProductID string `gorm:"uniqueIndex;size:64"`
	Stock     int    `gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func Init(cfg *config.MySQLConfig) error {
	var err error
	DB, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := DB.AutoMigrate(&Inventory{}); err != nil {
		return err
	}

	return nil
}
