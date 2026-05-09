package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"game_backend/internal/config"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type ServiceInfo struct {
	ServiceName string `json:"service_name"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
	Metadata    string `json:"metadata"`
}

var (
	Client *clientv3.Client
	lease  clientv3.LeaseID
)

func InitEtcd(cfg *config.EtcdConfig) (*clientv3.Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: time.Duration(cfg.DialTimeout) * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to etcd: %w", err)
	}

	Client = cli
	zap.L().Info("Etcd initialized successfully")
	return cli, nil
}

func RegisterService(serviceName string, serviceInfo *ServiceInfo) error {
	ctx := context.Background()

	resp, err := Client.Grant(ctx, 10)
	if err != nil {
		return fmt.Errorf("failed to grant lease: %w", err)
	}
	lease = resp.ID

	key := fmt.Sprintf("/services/%s/%s", serviceName, serviceInfo.Address)
	value, err := json.Marshal(serviceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}

	_, err = Client.Put(ctx, key, string(value), clientv3.WithLease(lease))
	if err != nil {
		return fmt.Errorf("failed to put service: %w", err)
	}

	keepAliveCh, err := Client.KeepAlive(ctx, lease)
	if err != nil {
		return fmt.Errorf("failed to keep alive: %w", err)
	}

	go func() {
		for range keepAliveCh {
		}
	}()

	zap.L().Info("Service registered to etcd",
		zap.String("service", serviceName),
		zap.String("address", serviceInfo.Address),
	)
	return nil
}

func DeregisterService(serviceName string, serviceInfo *ServiceInfo) error {
	ctx := context.Background()
	key := fmt.Sprintf("/services/%s/%s", serviceName, serviceInfo.Address)

	_, err := Client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	if lease != 0 {
		_, _ = Client.Revoke(ctx, lease)
	}

	zap.L().Info("Service deregistered from etcd",
		zap.String("service", serviceName),
		zap.String("address", serviceInfo.Address),
	)
	return nil
}

func DiscoverService(serviceName string) ([]*ServiceInfo, error) {
	ctx := context.Background()
	prefix := fmt.Sprintf("/services/%s/", serviceName)

	resp, err := Client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	var services []*ServiceInfo
	for _, kv := range resp.Kvs {
		var serviceInfo ServiceInfo
		if err := json.Unmarshal(kv.Value, &serviceInfo); err != nil {
			continue
		}
		services = append(services, &serviceInfo)
	}

	return services, nil
}

func WatchService(serviceName string) clientv3.WatchChan {
	prefix := fmt.Sprintf("/services/%s/", serviceName)
	return Client.Watch(context.Background(), prefix, clientv3.WithPrefix())
}

func Close() {
	if Client != nil {
		Client.Close()
	}
}

func HandleWatchEvents(watchChan clientv3.WatchChan, onAdd, onDel func(*ServiceInfo)) {
	for resp := range watchChan {
		for _, ev := range resp.Events {
			var serviceInfo ServiceInfo
			if err := json.Unmarshal(ev.Kv.Value, &serviceInfo); err != nil {
				continue
			}

			switch ev.Type {
			case mvccpb.PUT:
				if onAdd != nil {
					onAdd(&serviceInfo)
				}
			case mvccpb.DELETE:
				if onDel != nil {
					onDel(&serviceInfo)
				}
			}
		}
	}
}
