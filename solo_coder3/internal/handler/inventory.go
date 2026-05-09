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

func writeJSON(w http.ResponseWriter, code int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(resp)
}

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
