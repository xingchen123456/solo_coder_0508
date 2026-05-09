package main

import (
	"fmt"
	"log"
	"net/http"

	"inventory-service/config"
	"inventory-service/internal/cache"
	"inventory-service/internal/db"
	"inventory-service/internal/handler"
)

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

	mux := http.NewServeMux()
	mux.HandleFunc("/inventory/deduct", handler.Deduct)
	mux.HandleFunc("/inventory/stock", handler.GetStock)
	mux.HandleFunc("/inventory/set", handler.SetStock)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
