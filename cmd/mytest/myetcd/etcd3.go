package myetcd

import (
	"context"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"k8s.io/apiserver/pkg/storage"
)

const (
	keepaliveTime          = 30 * time.Second
	keepaliveTimeout       = 10 * time.Second
	dialTimeout            = 20 * time.Second
	dbMetricsMonitorJitter = 0.5
)

type DestroyFunc func()

func newETCD3Storage() (storage.Interface, DestroyFunc, error) {
	client, err := newETCD3Client()
	if err != nil {
		return nil, nil, err
	}
	// transformer := value.IdentityTransformer
	client.Get(context.Background(), "aaa")
	return nil, nil, nil
}

func newETCD3Client() (*clientv3.Client, error) {
	tlsInfo := transport.TLSInfo{
		CertFile:      "",
		KeyFile:       "",
		TrustedCAFile: "",
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return nil, err
	}
	tlsConfig = nil
	dialOptions := []grpc.DialOption{
		grpc.WithBlock(),
	}
	cfg := clientv3.Config{
		DialTimeout:          dialTimeout,
		DialKeepAliveTime:    keepaliveTime,
		DialKeepAliveTimeout: keepaliveTimeout,
		DialOptions:          dialOptions,
		Endpoints:            []string{"http://localhost:2379"},
		TLS:                  tlsConfig,
	}

	return clientv3.New(cfg)
}

func Run() {
	newETCD3Storage()
}
