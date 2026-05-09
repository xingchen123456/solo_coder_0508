package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"inventory-service/internal/service"
)

type CreateProductReq struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	InitialStock int     `json:"initial_stock"`
}

type UpdateProductReq struct {
	ProductID   string  `json:"product_id"`
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price,omitempty"`
	Status      *int    `json:"status,omitempty"`
}

type DeleteProductReq struct {
	ProductID string `json:"product_id"`
}

// @Summary 创建商品
// @Description 创建新商品，可同时设置初始库存，商品ID由服务端自动生成
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param request body CreateProductReq true "创建商品请求"
// @Success 200 {object} Response "创建成功，返回生成的商品ID"
// @Failure 400 {object} Response "参数错误"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /product/create [post]
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Code: 405, Message: "method not allowed"})
		return
	}

	var req CreateProductReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "invalid request"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "name is required"})
		return
	}
	if req.Price < 0 {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "invalid price"})
		return
	}
	if req.InitialStock < 0 {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "invalid initial_stock"})
		return
	}

	result, err := service.CreateProduct(r.Context(), service.CreateProductRequest{
		Name:         req.Name,
		Description:  req.Description,
		Price:        req.Price,
		InitialStock: req.InitialStock,
	})
	if err != nil {
		if errors.Is(err, service.ErrProductExists) {
			writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "product already exists"})
			return
		}
		if errors.Is(err, service.ErrLockAcquireFailed) {
			writeJSON(w, http.StatusTooManyRequests, Response{Code: 429, Message: "too many requests, please try again later"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, Response{Code: 500, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, Response{Code: 0, Message: "ok", Data: map[string]string{"product_id": result.ProductID}})
}

// @Summary 获取商品详情
// @Description 根据商品ID获取商品详情，包含库存信息
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param product_id query string true "商品ID"
// @Success 200 {object} Response "获取成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 404 {object} Response "商品不存在"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /product/get [get]
func GetProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Code: 405, Message: "method not allowed"})
		return
	}

	productID := r.URL.Query().Get("product_id")
	if productID == "" {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "product_id required"})
		return
	}

	product, err := service.GetProduct(r.Context(), productID)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			writeJSON(w, http.StatusNotFound, Response{Code: 404, Message: "product not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, Response{Code: 500, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, Response{Code: 0, Message: "ok", Data: product})
}

// @Summary 获取商品列表
// @Description 获取所有商品列表，包含库存信息
// @Tags 商品管理
// @Accept json
// @Produce json
// @Success 200 {object} Response "获取成功"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /product/list [get]
func ListProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Code: 405, Message: "method not allowed"})
		return
	}

	products, err := service.ListProducts(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{Code: 500, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, Response{Code: 0, Message: "ok", Data: map[string]interface{}{"products": products}})
}

// @Summary 更新商品
// @Description 更新商品信息，支持部分更新
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param request body UpdateProductReq true "更新商品请求"
// @Success 200 {object} Response "更新成功"
// @Failure 400 {object} Response "参数错误"
// @Failure 404 {object} Response "商品不存在"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /product/update [post]
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Code: 405, Message: "method not allowed"})
		return
	}

	var req UpdateProductReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "invalid request"})
		return
	}
	if req.ProductID == "" {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "product_id is required"})
		return
	}

	var updateReq service.UpdateProductRequest
	updateReq.ProductID = req.ProductID
	if req.Name != "" {
		updateReq.Name = req.Name
	}
	if req.Description != "" {
		updateReq.Description = req.Description
	}
	if req.Price > 0 {
		updateReq.Price = req.Price
	}
	if req.Status != nil {
		updateReq.Status = req.Status
	}

	if err := service.UpdateProduct(r.Context(), updateReq); err != nil {
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

	writeJSON(w, http.StatusOK, Response{Code: 0, Message: "ok"})
}

// @Summary 删除商品
// @Description 删除商品及其库存，同时从所有购物车中移除该商品
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param request body DeleteProductReq true "删除商品请求"
// @Success 200 {object} Response "删除成功"
// @Failure 400 {object} Response "参数错误"
// @Failure 404 {object} Response "商品不存在"
// @Failure 500 {object} Response "服务器内部错误"
// @Router /product/delete [post]
func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, Response{Code: 405, Message: "method not allowed"})
		return
	}

	var req DeleteProductReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "invalid request"})
		return
	}
	if req.ProductID == "" {
		writeJSON(w, http.StatusBadRequest, Response{Code: 400, Message: "product_id is required"})
		return
	}

	if err := service.DeleteProduct(r.Context(), req.ProductID); err != nil {
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

	writeJSON(w, http.StatusOK, Response{Code: 0, Message: "ok"})
}
