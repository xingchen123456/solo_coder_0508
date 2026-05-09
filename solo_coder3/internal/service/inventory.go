package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"inventory-service/internal/cache"
	"inventory-service/internal/db"
	"inventory-service/internal/lock"
)

var (
	ErrInsufficientStock    = errors.New("insufficient stock")
	ErrProductNotFound      = errors.New("product not found")
	ErrLockAcquireFailed    = errors.New("failed to acquire lock")
	ErrRequestInProgress    = errors.New("request is being processed")
	ErrCartItemExists       = errors.New("item already in cart")
	ErrCartItemNotFound     = errors.New("item not found in cart")
	ErrInvalidQuantity     = errors.New("invalid quantity")
	ErrUserIDRequired       = errors.New("user_id required")
)

type DeductRequest struct {
	RequestID string
	ProductID string
	Quantity  int
}

type DeductResult struct {
	Success bool
	Retried bool
	Reason  string
}

type CartItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type AddToCartRequest struct {
	UserID    string
	ProductID string
	Quantity  int
}

type UpdateCartRequest struct {
	UserID    string
	ProductID string
	Quantity  int
}

const (
	lockExpiration    = 30 * time.Second
	idempotentTTL     = 24 * time.Hour
	maxRetryAttempts  = 3
	retryDelay        = 100 * time.Millisecond
)

func GetOrCreateInventory(productID string, initialStock int) (*db.Inventory, error) {
	var inv db.Inventory
	err := db.DB.Where("product_id = ?", productID).First(&inv).Error
	if err == nil {
		return &inv, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	inv = db.Inventory{
		ProductID: productID,
		Stock:     initialStock,
	}
	if err := db.DB.Create(&inv).Error; err != nil {
		return nil, err
	}
	return &inv, nil
}

func DeductStock(ctx context.Context, req DeductRequest) (*DeductResult, error) {
	if req.RequestID == "" {
		return nil, errors.New("request_id is required")
	}

	processed, err := cache.IsRequestProcessed(ctx, req.RequestID)
	if err != nil {
		return nil, err
	}
	if processed {
		result, err := cache.GetRequestResult(ctx, req.RequestID)
		if err != nil {
			return nil, err
		}
		return &DeductResult{
			Success: result == "success",
			Retried: true,
			Reason:  fmt.Sprintf("request already processed, result: %s", result),
		}, nil
	}

	if err := CheckProductAvailable(ctx, req.ProductID, req.Quantity); err != nil {
		_ = cache.MarkRequestProcessed(ctx, req.RequestID, "product_error", idempotentTTL)
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		l, acquired, err := lock.TryLock(ctx, lock.TypeInventoryLock, req.ProductID, req.RequestID, lockExpiration)
		if err != nil {
			lastErr = err
			continue
		}
		if !acquired {
			if attempt < maxRetryAttempts-1 {
				time.Sleep(retryDelay)
				continue
			}
			return nil, ErrLockAcquireFailed
		}

		result, deductErr := deductWithLock(ctx, l, req)
		if deductErr != nil {
			lastErr = deductErr
			continue
		}

		return result, nil
	}

	return nil, lastErr
}

func deductWithLock(ctx context.Context, l *lock.Lock, req DeductRequest) (*DeductResult, error) {
	defer func() {
		_ = l.Unlock(context.Background())
	}()

	extendCtx, cancelExtend := context.WithCancel(context.Background())
	defer cancelExtend()
	go lock.ExtendLock(extendCtx, l, lockExpiration)

	redisResult, err := cache.DecrementStock(ctx, req.ProductID, req.Quantity)
	if err == nil && redisResult == 0 {
		_ = cache.MarkRequestProcessed(ctx, req.RequestID, "insufficient_stock", idempotentTTL)
		return &DeductResult{
			Success: false,
			Retried: false,
			Reason:  "insufficient stock",
		}, ErrInsufficientStock
	}

	dbErr := deductFromDB(ctx, req)
	if dbErr != nil {
		var result string
		if errors.Is(dbErr, ErrInsufficientStock) {
			result = "insufficient_stock"
		} else if errors.Is(dbErr, ErrProductNotFound) {
			result = "product_not_found"
		} else {
			result = "error"
		}
		_ = cache.MarkRequestProcessed(ctx, req.RequestID, result, idempotentTTL)
		return nil, dbErr
	}

	_ = cache.MarkRequestProcessed(ctx, req.RequestID, "success", idempotentTTL)
	return &DeductResult{
		Success: true,
		Retried: false,
		Reason:  "success",
	}, nil
}

func deductFromDB(ctx context.Context, req DeductRequest) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		var inv db.Inventory
		if err := tx.Clauses(gorm.Expr("FOR UPDATE")).
			Where("product_id = ?", req.ProductID).
			First(&inv).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProductNotFound
			}
			return err
		}

		if inv.Stock < req.Quantity {
			return ErrInsufficientStock
		}

		result := tx.Model(&inv).
			Where("product_id = ? AND stock >= ?", req.ProductID, req.Quantity).
			Update("stock", gorm.Expr("stock - ?", req.Quantity))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrInsufficientStock
		}

		return nil
	})
}

func GetStock(ctx context.Context, productID string) (int, error) {
	stock, err := cache.GetStock(ctx, productID)
	if err == nil {
		return stock, nil
	}

	var inv db.Inventory
	if err := db.DB.Where("product_id = ?", productID).First(&inv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrProductNotFound
		}
		return 0, err
	}

	_ = cache.SetStock(ctx, productID, inv.Stock, 5*time.Minute)

	return inv.Stock, nil
}

func SetStock(ctx context.Context, productID string, stock int) error {
	inv, err := GetOrCreateInventory(productID, stock)
	if err != nil {
		return err
	}

	if inv.Stock != stock {
		if err := db.DB.Model(&inv).Update("stock", stock).Error; err != nil {
			return err
		}
	}

	return cache.SetStock(ctx, productID, stock, 5*time.Minute)
}

func AddToCart(ctx context.Context, req AddToCartRequest) error {
	if req.UserID == "" {
		return ErrUserIDRequired
	}
	if req.ProductID == "" {
		return ErrProductNotFound
	}
	if req.Quantity <= 0 {
		return ErrInvalidQuantity
	}

	if err := CheckProductAvailable(ctx, req.ProductID, req.Quantity); err != nil {
		return err
	}

	exists, err := cache.CartItemExists(ctx, req.UserID, req.ProductID)
	if err != nil {
		return err
	}
	if exists {
		return ErrCartItemExists
	}

	return cache.CartAddItem(ctx, req.UserID, req.ProductID, req.Quantity)
}

func GetCart(ctx context.Context, userID string) ([]CartItem, error) {
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	items, err := cache.CartGetAll(ctx, userID)
	if err != nil {
		return nil, err
	}

	cartItems := make([]CartItem, 0, len(items))
	for productID, qtyStr := range items {
		var qty int
		fmt.Sscanf(qtyStr, "%d", &qty)
		cartItems = append(cartItems, CartItem{
			ProductID: productID,
			Quantity:  qty,
		})
	}

	return cartItems, nil
}

func UpdateCartItem(ctx context.Context, req UpdateCartRequest) error {
	if req.UserID == "" {
		return ErrUserIDRequired
	}
	if req.ProductID == "" {
		return ErrProductNotFound
	}
	if req.Quantity < 0 {
		return ErrInvalidQuantity
	}

	exists, err := cache.CartItemExists(ctx, req.UserID, req.ProductID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrCartItemNotFound
	}

	if req.Quantity == 0 {
		return cache.CartDeleteItem(ctx, req.UserID, req.ProductID)
	}

	if err := CheckProductAvailable(ctx, req.ProductID, req.Quantity); err != nil {
		return err
	}

	return cache.CartAddItem(ctx, req.UserID, req.ProductID, req.Quantity)
}

func RemoveFromCart(ctx context.Context, userID string, productID string) error {
	if userID == "" {
		return ErrUserIDRequired
	}
	if productID == "" {
		return ErrProductNotFound
	}

	exists, err := cache.CartItemExists(ctx, userID, productID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrCartItemNotFound
	}

	return cache.CartDeleteItem(ctx, userID, productID)
}

func ClearCart(ctx context.Context, userID string) error {
	if userID == "" {
		return ErrUserIDRequired
	}

	return cache.CartClear(ctx, userID)
}
