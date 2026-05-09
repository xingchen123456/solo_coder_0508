package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"inventory-service/config"
)

var RDB *redis.Client

var stockLua = `
local key = KEYS[1]
local quantity = tonumber(ARGV[1])
local stock = redis.call('get', key)
if stock == false then
    return -1
end
if tonumber(stock) < quantity then
    return 0
end
redis.call('decrby', key, quantity)
return 1
`

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

func StockKey(productID string) string {
	return fmt.Sprintf("stock:%s", productID)
}

func LockKey(productID string) string {
	return fmt.Sprintf("lock:stock:%s", productID)
}

func IdempotentKey(requestID string) string {
	return fmt.Sprintf("idempotent:%s", requestID)
}

func SetStock(ctx context.Context, productID string, stock int, expiration time.Duration) error {
	return RDB.Set(ctx, StockKey(productID), stock, expiration).Err()
}

func GetStock(ctx context.Context, productID string) (int, error) {
	return RDB.Get(ctx, StockKey(productID)).Int()
}

func DecrementStock(ctx context.Context, productID string, quantity int) (int, error) {
	result, err := RDB.Eval(ctx, stockLua, []string{StockKey(productID)}, quantity).Int()
	return result, err
}

type Lock struct {
	key        string
	requestID  string
	expiration time.Duration
}

func TryLock(ctx context.Context, productID string, requestID string, expiration time.Duration) (*Lock, bool, error) {
	key := LockKey(productID)
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

func IsRequestProcessed(ctx context.Context, requestID string) (bool, error) {
	exists, err := RDB.Exists(ctx, IdempotentKey(requestID)).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

func MarkRequestProcessed(ctx context.Context, requestID string, result string, expiration time.Duration) error {
	return RDB.Set(ctx, IdempotentKey(requestID), result, expiration).Err()
}

func GetRequestResult(ctx context.Context, requestID string) (string, error) {
	return RDB.Get(ctx, IdempotentKey(requestID)).Result()
}
