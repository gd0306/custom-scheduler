package pkg

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func InitEtcdClient(etcdCluster []string) (*clientv3.Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Context:     context.TODO(),
		Endpoints:   etcdCluster,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return cli, nil
}
