package service

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"inventory-service/internal/cache"
	"inventory-service/internal/db"
	"inventory-service/internal/idgen"
	"inventory-service/internal/lock"
)

var (
	ErrProductExists       = errors.New("product already exists")
	ErrProductDisabled     = errors.New("product is disabled")
	ErrInvalidPrice        = errors.New("invalid price")
	ErrProductNameRequired = errors.New("product name is required")
)

type CreateProductRequest struct {
	Name         string
	Description  string
	Price        float64
	InitialStock int
}

type CreateProductResult struct {
	ProductID string
}

type UpdateProductRequest struct {
	ProductID   string
	Name        string
	Description string
	Price       float64
	Status      *int
}

type ProductDetail struct {
	ProductID   string  `json:"product_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Status      int     `json:"status"`
	Stock       int     `json:"stock"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

const (
	productLockExpiration = 30 * time.Second
	productCacheTTL       = 10 * time.Minute
	maxProductRetryAttempts = 3
	productRetryDelay     = 100 * time.Millisecond
)

func CreateProduct(ctx context.Context, req CreateProductRequest) (*CreateProductResult, error) {
	if req.Name == "" {
		return nil, ErrProductNameRequired
	}
	if req.Price < 0 {
		return nil, ErrInvalidPrice
	}
	if req.InitialStock < 0 {
		return nil, ErrInvalidQuantity
	}

	var lastErr error
	for attempt := 0; attempt < maxProductRetryAttempts; attempt++ {
		productID := idgen.Generate()

		var existing db.Product
		err := db.DB.Where("product_id = ?", productID).First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			lastErr = err
			continue
		}

		requestID := "create_" + productID
		l, acquired, err := lock.TryLock(ctx, lock.TypeProductLock, productID, requestID, productLockExpiration)
		if err != nil {
			lastErr = err
			continue
		}
		if !acquired {
			if attempt < maxProductRetryAttempts-1 {
				time.Sleep(productRetryDelay)
				continue
			}
			return nil, ErrLockAcquireFailed
		}

		result, createErr := createProductWithLock(ctx, l, req, productID)
		if createErr != nil {
			lastErr = createErr
			continue
		}

		return result, nil
	}

	return nil, lastErr
}

func createProductWithLock(ctx context.Context, l *lock.Lock, req CreateProductRequest, productID string) (*CreateProductResult, error) {
	defer func() {
		_ = l.Unlock(context.Background())
	}()

	extendCtx, cancelExtend := context.WithCancel(context.Background())
	defer cancelExtend()
	go lock.ExtendLock(extendCtx, l, productLockExpiration)

	var existing db.Product
	err := db.DB.Where("product_id = ?", productID).First(&existing).Error
	if err == nil {
		return nil, ErrProductExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	product := db.Product{
		ProductID:   productID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Status:      1,
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&product).Error; err != nil {
			return err
		}

		if req.InitialStock > 0 {
			inv := db.Inventory{
				ProductID: productID,
				Stock:     req.InitialStock,
			}
			if err := tx.Create(&inv).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	_ = cache.SetProduct(ctx, &product, productCacheTTL)
	if req.InitialStock > 0 {
		_ = cache.SetStock(ctx, productID, req.InitialStock, 5*time.Minute)
	}
	_ = cache.DelProductList(ctx)

	return &CreateProductResult{
		ProductID: productID,
	}, nil
}

func GetProduct(ctx context.Context, productID string) (*ProductDetail, error) {
	if productID == "" {
		return nil, ErrProductNotFound
	}

	cachedProduct, err := cache.GetProduct(ctx, productID)
	if err == nil {
		stock, _ := GetStock(ctx, productID)
		return &ProductDetail{
			ProductID:   cachedProduct.ProductID,
			Name:        cachedProduct.Name,
			Description: cachedProduct.Description,
			Price:       cachedProduct.Price,
			Status:      cachedProduct.Status,
			Stock:       stock,
			CreatedAt:   cachedProduct.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   cachedProduct.UpdatedAt.Format(time.RFC3339),
		}, nil
	}

	var product db.Product
	err = db.DB.Where("product_id = ?", productID).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	_ = cache.SetProduct(ctx, &product, productCacheTTL)

	stock, err := GetStock(ctx, productID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		stock = 0
	}

	return &ProductDetail{
		ProductID:   product.ProductID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Status:      product.Status,
		Stock:       stock,
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func ListProducts(ctx context.Context) ([]ProductDetail, error) {
	var products []db.Product
	if err := db.DB.Order("created_at DESC").Find(&products).Error; err != nil {
		return nil, err
	}

	result := make([]ProductDetail, 0, len(products))
	for _, p := range products {
		stock, _ := GetStock(ctx, p.ProductID)
		result = append(result, ProductDetail{
			ProductID:   p.ProductID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Status:      p.Status,
			Stock:       stock,
			CreatedAt:   p.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
		})
	}

	return result, nil
}

func UpdateProduct(ctx context.Context, req UpdateProductRequest) error {
	if req.ProductID == "" {
		return ErrProductNotFound
	}
	if req.Price < 0 {
		return ErrInvalidPrice
	}

	requestID := "update_" + req.ProductID
	var lastErr error
	for attempt := 0; attempt < maxProductRetryAttempts; attempt++ {
		l, acquired, err := lock.TryLock(ctx, lock.TypeProductLock, req.ProductID, requestID, productLockExpiration)
		if err != nil {
			lastErr = err
			continue
		}
		if !acquired {
			if attempt < maxProductRetryAttempts-1 {
				time.Sleep(productRetryDelay)
				continue
			}
			return ErrLockAcquireFailed
		}

		updateErr := updateProductWithLock(ctx, l, req)
		if updateErr != nil {
			lastErr = updateErr
			continue
		}

		return nil
	}

	return lastErr
}

func updateProductWithLock(ctx context.Context, l *lock.Lock, req UpdateProductRequest) error {
	defer func() {
		_ = l.Unlock(context.Background())
	}()

	extendCtx, cancelExtend := context.WithCancel(context.Background())
	defer cancelExtend()
	go lock.ExtendLock(extendCtx, l, productLockExpiration)

	var product db.Product
	if err := db.DB.Where("product_id = ?", req.ProductID).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}
		return err
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Price > 0 {
		product.Price = req.Price
	}
	if req.Status != nil {
		product.Status = *req.Status
	}

	if err := db.DB.Save(&product).Error; err != nil {
		return err
	}

	_ = cache.SetProduct(ctx, &product, productCacheTTL)
	_ = cache.DelProductList(ctx)

	return nil
}

func DeleteProduct(ctx context.Context, productID string) error {
	if productID == "" {
		return ErrProductNotFound
	}

	requestID := "delete_" + productID
	var lastErr error
	for attempt := 0; attempt < maxProductRetryAttempts; attempt++ {
		l, acquired, err := lock.TryLock(ctx, lock.TypeProductLock, productID, requestID, productLockExpiration)
		if err != nil {
			lastErr = err
			continue
		}
		if !acquired {
			if attempt < maxProductRetryAttempts-1 {
				time.Sleep(productRetryDelay)
				continue
			}
			return ErrLockAcquireFailed
		}

		deleteErr := deleteProductWithLock(ctx, l, productID)
		if deleteErr != nil {
			lastErr = deleteErr
			continue
		}

		return nil
	}

	return lastErr
}

func deleteProductWithLock(ctx context.Context, l *lock.Lock, productID string) error {
	defer func() {
		_ = l.Unlock(context.Background())
	}()

	extendCtx, cancelExtend := context.WithCancel(context.Background())
	defer cancelExtend()
	go lock.ExtendLock(extendCtx, l, productLockExpiration)

	var product db.Product
	if err := db.DB.Where("product_id = ?", productID).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}
		return err
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&product).Error; err != nil {
			return err
		}

		if err := tx.Where("product_id = ?", productID).Delete(&db.Inventory{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	_ = cache.DelProduct(ctx, productID)
	_ = cache.DelStock(ctx, productID)
	_ = cache.DelProductList(ctx)

	_ = cache.CartRemoveProductFromAll(ctx, productID)

	return nil
}

func CheckProductAvailable(ctx context.Context, productID string, quantity int) error {
	product, err := cache.GetProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			var p db.Product
			if dbErr := db.DB.Where("product_id = ?", productID).First(&p).Error; dbErr != nil {
				if errors.Is(dbErr, gorm.ErrRecordNotFound) {
					return ErrProductNotFound
				}
				return dbErr
			}
			product = &p
			_ = cache.SetProduct(ctx, product, productCacheTTL)
		} else {
			return err
		}
	}

	if product.Status != 1 {
		return ErrProductDisabled
	}

	if quantity > 0 {
		stock, err := GetStock(ctx, productID)
		if err != nil {
			return err
		}
		if stock < quantity {
			return ErrInsufficientStock
		}
	}

	return nil
}
