package utils

import (
	"context"
	"fmt"
	"log"

	"management-system/config"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func InitRedis() error {
	cfg := config.AppConfig.Redis
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx := context.Background()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
		return err
	}

	log.Println("Redis connected successfully")
	return nil
}

func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}
