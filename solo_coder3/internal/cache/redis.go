package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"inventory-service/config"
	"inventory-service/internal/db"
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

func ProductKey(productID string) string {
	return fmt.Sprintf("product:%s", productID)
}

func ProductListKey() string {
	return "product:list"
}

func IdempotentKey(requestID string) string {
	return fmt.Sprintf("idempotent:%s", requestID)
}

func CartKey(userID string) string {
	return fmt.Sprintf("cart:%s", userID)
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

func DelStock(ctx context.Context, productID string) error {
	return RDB.Del(ctx, StockKey(productID)).Err()
}

func SetProduct(ctx context.Context, product *db.Product, expiration time.Duration) error {
	data, err := json.Marshal(product)
	if err != nil {
		return err
	}
	return RDB.Set(ctx, ProductKey(product.ProductID), data, expiration).Err()
}

func GetProduct(ctx context.Context, productID string) (*db.Product, error) {
	data, err := RDB.Get(ctx, ProductKey(productID)).Bytes()
	if err != nil {
		return nil, err
	}
	var product db.Product
	if err := json.Unmarshal(data, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

func DelProduct(ctx context.Context, productID string) error {
	return RDB.Del(ctx, ProductKey(productID)).Err()
}

func DelProductList(ctx context.Context) error {
	return RDB.Del(ctx, ProductListKey()).Err()
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

func CartAddItem(ctx context.Context, userID string, productID string, quantity int) error {
	return RDB.HSet(ctx, CartKey(userID), productID, quantity).Err()
}

func CartGetItem(ctx context.Context, userID string, productID string) (string, error) {
	return RDB.HGet(ctx, CartKey(userID), productID).Result()
}

func CartGetAll(ctx context.Context, userID string) (map[string]string, error) {
	return RDB.HGetAll(ctx, CartKey(userID)).Result()
}

func CartItemExists(ctx context.Context, userID string, productID string) (bool, error) {
	return RDB.HExists(ctx, CartKey(userID), productID).Result()
}

func CartDeleteItem(ctx context.Context, userID string, productID string) error {
	return RDB.HDel(ctx, CartKey(userID), productID).Err()
}

func CartClear(ctx context.Context, userID string) error {
	return RDB.Del(ctx, CartKey(userID)).Err()
}

func CartGetQuantity(ctx context.Context, userID string, productID string) (int, error) {
	val, err := RDB.HGet(ctx, CartKey(userID), productID).Int()
	if err != nil {
		return 0, err
	}
	return val, nil
}

func CartHasProduct(ctx context.Context, productID string) (bool, error) {
	keys, err := RDB.Keys(ctx, "cart:*").Result()
	if err != nil {
		return false, err
	}
	for _, key := range keys {
		exists, err := RDB.HExists(ctx, key, productID).Result()
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}
	return false, nil
}

func CartRemoveProductFromAll(ctx context.Context, productID string) error {
	keys, err := RDB.Keys(ctx, "cart:*").Result()
	if err != nil {
		return err
	}
	for _, key := range keys {
		_ = RDB.HDel(ctx, key, productID).Err()
	}
	return nil
}
