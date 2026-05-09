package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitLogger(mode string) error {
	var cfg zap.Config
	if mode == "release" {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.LevelKey = "level"
		cfg.EncoderConfig.TimeKey = "time"
		cfg.EncoderConfig.MessageKey = "msg"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var err error
	Logger, err = cfg.Build()
	if err != nil {
		return err
	}

	defer Logger.Sync()
	return nil
}
