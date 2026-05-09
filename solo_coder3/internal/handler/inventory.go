package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"inventory-service/internal/service"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type DeductReq struct {
	RequestID string `json:"request_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type DeductRespData struct {
	Success bool   `json:"success"`
	Retried bool   `json:"retried"`
	Reason  string `json:"reason"`
}

type SetStockReq struct {
	ProductID string `json:"product_id"`
	Stock     int    `json:"stock"`
}

type AddToCartReq struct {
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type UpdateCartReq struct {
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type RemoveFromCartReq struct {
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id"`
}

type ClearCartReq struct {
	UserID string `json:"user_id"`
}

func writeJSON(w http.ResponseWriter, code int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(resp)
}

// @Summary 扣减库存
// @Description 根据请求ID幂等扣减商品库存
// @Tags 库存管理
// @Accept json
// @Produce json
// @Param request body DeductReq true "扣减请求"
// @Success 200 {object} Response "扣减成功"
// @Failure 400 {object} Response "请求参数错误或库存不足"
// @Failure 404 {object} Response "商品不存在"
// @Failure 429 {object} Response "请求频繁，请稍后重试"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /inventory/deduct [post]
func Deduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Code: 405, Message: "method not allowed"})
		return
	}

	var req DeductReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "invalid request"})
		return
	}
	if req.RequestID == "" {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "request_id is required"})
		return
	}
	if req.ProductID == "" || req.Quantity <= 0 {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "invalid product_id or quantity"})
		return
	}

	result, err := service.DeductStock(r.Context(), service.DeductRequest{
		RequestID: req.RequestID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	})

	if err != nil {
		if errors.Is(err, service.ErrInsufficientStock) {
			writeJSON(w, http.StatusBadRequest, Response{
				Code:    400,
				Message: "insufficient stock",
				Data: DeductRespData{
					Success: false,
					Retried: result != nil && result.Retried,
					Reason:  "insufficient stock",
				},
			})
			return
		}
		if errors.Is(err, service.ErrProductNotFound) {
			writeJSON(w, http.StatusNotFound, Response{Code: 404, Message: "product not found"})
			return
		}
		if errors.Is(err, service.ErrLockAcquireFailed) {
			writeJSON(w, http.StatusTooManyRequests, Response{Code: 429, Message: "too many requests, please try again later"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, Response{Code: 500, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, Response{
		Code:    0,
		Message: "ok",
		Data: DeductRespData{
			Success: result.Success,
			Retried: result.Retried,
			Reason:  result.Reason,
		},
	})
}

// @Summary 获取库存
// @Description 根据商品ID获取当前库存数量
// @Tags 库存管理
// @Accept json
// @Produce json
// @Param product_id query string true "商品ID"
// @Success 200 {object} Response "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 404 {object} Response "商品不存在"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /inventory/stock [get]
func GetStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Code: 405, Message: "method not allowed"})
		return
	}

	productID := r.URL.Query().Get("product_id")
	if productID == "" {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "product_id required"})
		return
	}

	stock, err := service.GetStock(r.Context(), productID)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			writeJSON(w, http.StatusNotFound, Response{Code: 404, Message: "product not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, Response{Code: 500, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, Response{Code: 0, Message: "ok", Data: map[string]int{"stock": stock}})
}

// @Summary 设置库存
// @Description 设置指定商品的库存数量
// @Tags 库存管理
// @Accept json
// @Produce json
// @Param request body SetStockReq true "设置库存请求"
// @Success 200 {object} Response "设置成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /inventory/set [post]
func SetStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Code: 405, Message: "method not allowed"})
		return
	}

	var req SetStockReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "invalid request"})
		return
	}
	if req.ProductID == "" || req.Stock < 0 {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "invalid product_id or stock"})
		return
	}

	if err := service.SetStock(r.Context(), req.ProductID, req.Stock); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Code: 500, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, Response{Code: 0, Message: "ok"})
}

// @Summary 添加商品到购物车
// @Description 将商品添加到用户购物车，不允许重复添加，同时检查库存是否充足
// @Tags 购物车
// @Accept json
// @Produce json
// @Param request body AddToCartReq true "添加购物车请求"
// @Success 200 {object} Response "添加成功"
// @Failure 400 {object} Response "商品已在购物车中或库存不足或参数错误"
// @Failure 404 {object} Response "商品不存在"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /cart/add [post]
func AddToCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Code: 405, Message: "method not allowed"})
		return
	}

	var req AddToCartReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "invalid request"})
		return
	}
	if req.UserID == "" {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "user_id is required"})
		return
	}
	if req.ProductID == "" || req.Quantity <= 0 {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "invalid product_id or quantity"})
		return
	}

	if err := service.AddToCart(r.Context(), service.AddToCartRequest{
		UserID:    req.UserID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	}); err != nil {
		if errors.Is(err, service.ErrCartItemExists) {
			writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "item already in cart"})
			return
		}
		if errors.Is(err, service.ErrProductNotFound) {
			writeJSON(w, http.StatusNotFound, Response{Code: 404, Message: "product not found"})
			return
		}
		if errors.Is(err, service.ErrInsufficientStock) {
			writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "insufficient stock"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, Response{Code: 500, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, Response{Code: 0, Message: "ok"})
}

// @Summary 获取购物车
// @Description 获取用户购物车中的所有商品
// @Tags 购物车
// @Accept json
// @Produce json
// @Param user_id query string true "用户ID"
// @Success 200 {object} Response "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /cart/get [get]
func GetCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Code: 405, Message: "method not allowed"})
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "user_id required"})
		return
	}

	cart, err := service.GetCart(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Code: 500, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, Response{Code: 0, Message: "ok", Data: map[string]interface{}{"items": cart}})
}

// @Summary 更新购物车商品数量
// @Description 更新购物车中指定商品的数量，数量为0时移除该商品，同时检查库存是否充足
// @Tags 购物车
// @Accept json
// @Produce json
// @Param request body UpdateCartReq true "更新购物车请求"
// @Success 200 {object} Response "更新成功"
// @Failure 400 {object} Response "库存不足或参数错误"
// @Failure 404 {object} Response "商品不在购物车中"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /cart/update [post]
func UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Code: 405, Message: "method not allowed"})
		return
	}

	var req UpdateCartReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "invalid request"})
		return
	}
	if req.UserID == "" {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "user_id is required"})
		return
	}
	if req.ProductID == "" || req.Quantity < 0 {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "invalid product_id or quantity"})
		return
	}

	if err := service.UpdateCartItem(r.Context(), service.UpdateCartRequest{
		UserID:    req.UserID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	}); err != nil {
		if errors.Is(err, service.ErrCartItemNotFound) {
			writeJSON(w, http.StatusNotFound, Response{Code: 404, Message: "item not found in cart"})
			return
		}
		if errors.Is(err, service.ErrInsufficientStock) {
			writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "insufficient stock"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, Response{Code: 500, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, Response{Code: 0, Message: "ok"})
}

// @Summary 从购物车移除商品
// @Description 从用户购物车中移除指定商品
// @Tags 购物车
// @Accept json
// @Produce json
// @Param request body RemoveFromCartReq true "移除购物车请求"
// @Success 200 {object} Response "移除成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 404 {object} Response "商品不在购物车中"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /cart/remove [post]
func RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Code: 405, Message: "method not allowed"})
		return
	}

	var req RemoveFromCartReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "invalid request"})
		return
	}
	if req.UserID == "" {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "user_id is required"})
		return
	}
	if req.ProductID == "" {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "product_id is required"})
		return
	}

	if err := service.RemoveFromCart(r.Context(), req.UserID, req.ProductID); err != nil {
		if errors.Is(err, service.ErrCartItemNotFound) {
			writeJSON(w, http.StatusNotFound, Response{Code: 404, Message: "item not found in cart"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, Response{Code: 500, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, Response{Code: 0, Message: "ok"})
}

// @Summary 清空购物车
// @Description 清空用户购物车中的所有商品
// @Tags 购物车
// @Accept json
// @Produce json
// @Param request body ClearCartReq true "清空购物车请求"
// @Success 200 {object} Response "清空成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /cart/clear [post]
func ClearCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Code: 405, Message: "method not allowed"})
		return
	}

	var req ClearCartReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "invalid request"})
		return
	}
	if req.UserID == "" {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "user_id is required"})
		return
	}

	if err := service.ClearCart(r.Context(), req.UserID); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Code: 500, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, Response{Code: 0, Message: "ok"})
}
