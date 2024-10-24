package etcd_api

import (
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// NewEtcdClient 创建新的 etcd client
func NewEtcdClient(listenAddr string, listenPort int) (*clientv3.Client, error) {
	endPoint := fmt.Sprintf("%s:%d", listenAddr, listenPort)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{endPoint},
	})
	if err != nil {
		return nil, fmt.Errorf("create etcd client failed: %w", err)
	}
	return cli, nil
}
