package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"inventory-service/config"
)

var RDB *redis.Client

var unlockLua = `
if redis.call('get', KEYS[1]) == ARGV[1] then
    return redis.call('del', KEYS[1])
end
return 0
`

var extendLua = `
if redis.call('get', KEYS[1]) == ARGV[1] then
    return redis.call('pexpire', KEYS[1], ARGV[2])
end
return 0
`

type Lock struct {
	key        string
	requestID  string
	expiration time.Duration
}

type Type string

const (
	TypeProductLock  Type = "product"
	TypeInventoryLock Type = "inventory"
	TypeCartLock     Type = "cart"
)

func Init(cfg *config.RedisConfig) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		return err
	}

	return nil
}

func LockKey(lockType Type, resourceID string) string {
	return fmt.Sprintf("lock:%s:%s", lockType, resourceID)
}

func TryLock(ctx context.Context, lockType Type, resourceID string, requestID string, expiration time.Duration) (*Lock, bool, error) {
	key := LockKey(lockType, resourceID)
	ok, err := RDB.SetNX(ctx, key, requestID, expiration).Result()
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	return &Lock{
		key:        key,
		requestID:  requestID,
		expiration: expiration,
	}, true, nil
}

func (l *Lock) Unlock(ctx context.Context) error {
	_, err := RDB.Eval(ctx, unlockLua, []string{l.key}, l.requestID).Result()
	return err
}

func (l *Lock) Extend(ctx context.Context, extension time.Duration) (bool, error) {
	result, err := RDB.Eval(ctx, extendLua, []string{l.key}, l.requestID, int64(extension.Milliseconds())).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func ExtendLock(ctx context.Context, l *Lock, lockExpiration time.Duration) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = l.Extend(ctx, lockExpiration)
		}
	}
}
