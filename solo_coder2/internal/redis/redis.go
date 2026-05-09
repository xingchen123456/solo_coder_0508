package redis

import (
	"context"
	"fmt"

	"game_backend/internal/config"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

var RDB *redis.Client

func InitRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	RDB = rdb
	zap.L().Info("Redis initialized successfully")
	return rdb, nil
}

func Close() {
	if RDB != nil {
		RDB.Close()
	}
}
