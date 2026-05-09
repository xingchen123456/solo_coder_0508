package main

import (
	"fmt"
	"log"
	"net/http"

	"inventory-service/config"
	"inventory-service/internal/cache"
	"inventory-service/internal/db"
	"inventory-service/internal/handler"
	"inventory-service/internal/lock"
)

// @title Inventory Service API
// @version 2.0
// @description 库存服务 API，支持商品管理、库存管理和购物车功能
// @host localhost:8080
// @BasePath /
// @schemes http

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	if err := db.Init(&cfg.MySQL); err != nil {
		log.Fatalf("init mysql failed: %v", err)
	}
	log.Println("mysql connected")

	if err := cache.Init(&cfg.Redis); err != nil {
		log.Fatalf("init redis failed: %v", err)
	}
	log.Println("redis connected")

	if err := lock.Init(&cfg.Redis); err != nil {
		log.Fatalf("init lock failed: %v", err)
	}
	log.Println("distributed lock initialized")

	mux := http.NewServeMux()
	mux.HandleFunc("/product/create", handler.CreateProduct)
	mux.HandleFunc("/product/get", handler.GetProduct)
	mux.HandleFunc("/product/list", handler.ListProducts)
	mux.HandleFunc("/product/update", handler.UpdateProduct)
	mux.HandleFunc("/product/delete", handler.DeleteProduct)
	mux.HandleFunc("/inventory/deduct", handler.Deduct)
	mux.HandleFunc("/inventory/stock", handler.GetStock)
	mux.HandleFunc("/inventory/set", handler.SetStock)
	mux.HandleFunc("/cart/add", handler.AddToCart)
	mux.HandleFunc("/cart/get", handler.GetCart)
	mux.HandleFunc("/cart/update", handler.UpdateCartItem)
	mux.HandleFunc("/cart/remove", handler.RemoveFromCart)
	mux.HandleFunc("/cart/clear", handler.ClearCart)

	mux.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"swagger": "2.0",
			"info": {
				"title": "Inventory Service API",
				"version": "1.0",
				"description": "库存服务 API，支持库存管理和购物车功能"
			},
			"host": "localhost:8080",
			"basePath": "/",
			"schemes": ["http"],
			"tags": [
				{"name": "商品管理", "description": "商品相关接口"},
				{"name": "库存管理", "description": "库存相关接口"},
				{"name": "购物车", "description": "购物车相关接口"}
			],
			"paths": {
				"/product/create": {
					"post": {
						"tags": ["商品管理"],
						"summary": "创建商品",
						"description": "创建新商品，可同时设置初始库存，商品ID由服务端自动生成",
						"parameters": [{
							"name": "body",
							"in": "body",
							"required": true,
							"schema": {
								"type": "object",
								"properties": {
									"name": {"type": "string"},
									"description": {"type": "string"},
									"price": {"type": "number"},
									"initial_stock": {"type": "integer"}
								},
								"required": ["name"]
							}
						}],
						"responses": {
							"200": {"description": "创建成功，返回生成的商品ID"},
							"400": {"description": "参数错误"},
							"500": {"description": "服务器内部错误"}
						}
					}
				},
				"/product/get": {
					"get": {
						"tags": ["商品管理"],
						"summary": "获取商品详情",
						"description": "根据商品ID获取商品详情，包含库存信息",
						"parameters": [{
							"name": "product_id",
							"in": "query",
							"type": "string",
							"required": true
						}],
						"responses": {
							"200": {"description": "获取成功"},
							"400": {"description": "请求参数错误"},
							"404": {"description": "商品不存在"},
							"500": {"description": "服务器内部错误"}
						}
					}
				},
				"/product/list": {
					"get": {
						"tags": ["商品管理"],
						"summary": "获取商品列表",
						"description": "获取所有商品列表，包含库存信息",
						"responses": {
							"200": {"description": "获取成功"},
							"500": {"description": "服务器内部错误"}
						}
					}
				},
				"/product/update": {
					"post": {
						"tags": ["商品管理"],
						"summary": "更新商品",
						"description": "更新商品信息，支持部分更新",
						"parameters": [{
							"name": "body",
							"in": "body",
							"required": true,
							"schema": {
								"type": "object",
								"properties": {
									"product_id": {"type": "string"},
									"name": {"type": "string"},
									"description": {"type": "string"},
									"price": {"type": "number"},
									"status": {"type": "integer"}
								},
								"required": ["product_id"]
							}
						}],
						"responses": {
							"200": {"description": "更新成功"},
							"400": {"description": "参数错误"},
							"404": {"description": "商品不存在"},
							"500": {"description": "服务器内部错误"}
						}
					}
				},
				"/product/delete": {
					"post": {
						"tags": ["商品管理"],
						"summary": "删除商品",
						"description": "删除商品及其库存，同时从所有购物车中移除该商品",
						"parameters": [{
							"name": "body",
							"in": "body",
							"required": true,
							"schema": {
								"type": "object",
								"properties": {
									"product_id": {"type": "string"}
								},
								"required": ["product_id"]
							}
						}],
						"responses": {
							"200": {"description": "删除成功"},
							"400": {"description": "参数错误"},
							"404": {"description": "商品不存在"},
							"500": {"description": "服务器内部错误"}
						}
					}
				},
				"/inventory/deduct": {
					"post": {
						"tags": ["库存管理"],
						"summary": "扣减库存",
						"description": "根据请求ID幂等扣减商品库存",
						"parameters": [{
							"name": "body",
							"in": "body",
							"required": true,
							"schema": {
								"type": "object",
								"properties": {
									"request_id": {"type": "string"},
									"product_id": {"type": "string"},
									"quantity": {"type": "integer"}
								},
								"required": ["request_id", "product_id", "quantity"]
							}
						}],
						"responses": {
							"200": {"description": "扣减成功"},
							"400": {"description": "请求参数错误或库存不足"},
							"404": {"description": "商品不存在"},
							"429": {"description": "请求频繁，请稍后重试"},
							"500": {"description": "服务器内部错误"}
						}
					}
				},
				"/inventory/stock": {
					"get": {
						"tags": ["库存管理"],
						"summary": "获取库存",
						"description": "根据商品ID获取当前库存数量",
						"parameters": [{
							"name": "product_id",
							"in": "query",
							"type": "string",
							"required": true
						}],
						"responses": {
							"200": {"description": "获取成功"},
							"400": {"description": "请求参数错误"},
							"404": {"description": "商品不存在"},
							"500": {"description": "服务器内部错误"}
						}
					}
				},
				"/inventory/set": {
					"post": {
						"tags": ["库存管理"],
						"summary": "设置库存",
						"description": "设置指定商品的库存数量",
						"parameters": [{
							"name": "body",
							"in": "body",
							"required": true,
							"schema": {
								"type": "object",
								"properties": {
									"product_id": {"type": "string"},
									"stock": {"type": "integer"}
								},
								"required": ["product_id", "stock"]
							}
						}],
						"responses": {
							"200": {"description": "设置成功"},
							"400": {"description": "请求参数错误"},
							"500": {"description": "服务器内部错误"}
						}
					}
				},
				"/cart/add": {
					"post": {
						"tags": ["购物车"],
						"summary": "添加商品到购物车",
						"description": "将商品添加到用户购物车，不允许重复添加，同时检查库存是否充足",
						"parameters": [{
							"name": "body",
							"in": "body",
							"required": true,
							"schema": {
								"type": "object",
								"properties": {
									"user_id": {"type": "string"},
									"product_id": {"type": "string"},
									"quantity": {"type": "integer"}
								},
								"required": ["user_id", "product_id", "quantity"]
							}
						}],
						"responses": {
							"200": {"description": "添加成功"},
							"400": {"description": "商品已在购物车中或库存不足或参数错误"},
							"404": {"description": "商品不存在"},
							"500": {"description": "服务器内部错误"}
						}
					}
				},
				"/cart/get": {
					"get": {
						"tags": ["购物车"],
						"summary": "获取购物车",
						"description": "获取用户购物车中的所有商品",
						"parameters": [{
							"name": "user_id",
							"in": "query",
							"type": "string",
							"required": true
						}],
						"responses": {
							"200": {"description": "获取成功"},
							"400": {"description": "请求参数错误"},
							"500": {"description": "服务器内部错误"}
						}
					}
				},
				"/cart/update": {
					"post": {
						"tags": ["购物车"],
						"summary": "更新购物车商品数量",
						"description": "更新购物车中指定商品的数量，数量为0时移除该商品，同时检查库存是否充足",
						"parameters": [{
							"name": "body",
							"in": "body",
							"required": true,
							"schema": {
								"type": "object",
								"properties": {
									"user_id": {"type": "string"},
									"product_id": {"type": "string"},
									"quantity": {"type": "integer"}
								},
								"required": ["user_id", "product_id", "quantity"]
							}
						}],
						"responses": {
							"200": {"description": "更新成功"},
							"400": {"description": "库存不足或参数错误"},
							"404": {"description": "商品不在购物车中"},
							"500": {"description": "服务器内部错误"}
						}
					}
				},
				"/cart/remove": {
					"post": {
						"tags": ["购物车"],
						"summary": "从购物车移除商品",
						"description": "从用户购物车中移除指定商品",
						"parameters": [{
							"name": "body",
							"in": "body",
							"required": true,
							"schema": {
								"type": "object",
								"properties": {
									"user_id": {"type": "string"},
									"product_id": {"type": "string"}
								},
								"required": ["user_id", "product_id"]
							}
						}],
						"responses": {
							"200": {"description": "移除成功"},
							"400": {"description": "请求参数错误"},
							"404": {"description": "商品不在购物车中"},
							"500": {"description": "服务器内部错误"}
						}
					}
				},
				"/cart/clear": {
					"post": {
						"tags": ["购物车"],
						"summary": "清空购物车",
						"description": "清空用户购物车中的所有商品",
						"parameters": [{
							"name": "body",
							"in": "body",
							"required": true,
							"schema": {
								"type": "object",
								"properties": {
									"user_id": {"type": "string"}
								},
								"required": ["user_id"]
							}
						}],
						"responses": {
							"200": {"description": "清空成功"},
							"400": {"description": "请求参数错误"},
							"500": {"description": "服务器内部错误"}
						}
					}
				}
			}
		}`)
	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("server starting on %s", addr)
	log.Printf("swagger json available at: http://localhost:%d/swagger/doc.json", cfg.Server.Port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
