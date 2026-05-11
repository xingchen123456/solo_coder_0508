package svc

import (
	"solo-coder4/api/config"
	"solo-coder4/common/db"

	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {
	dbConn, err := db.InitDB(c.DB)
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config: c,
		DB:     dbConn,
	}
}
