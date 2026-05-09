package utils

import (
	"log"
	"time"

	"management-system/config"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var EtcdClient *clientv3.Client

func InitEtcd() error {
	cfg := config.AppConfig.Etcd
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: time.Duration(cfg.DialTimeout) * time.Second,
	})
	if err != nil {
		log.Printf("Failed to connect to etcd: %v", err)
		return err
	}

	EtcdClient = client
	log.Println("etcd connected successfully")
	return nil
}

func CloseEtcd() error {
	if EtcdClient != nil {
		return EtcdClient.Close()
	}
	return nil
}
