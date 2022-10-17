package factory

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"path"
	"strings"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	genericfeatures "k8s.io/apiserver/pkg/features"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/traces"

	grpcprom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	utilnet "k8s.io/apimachinery/pkg/util/net"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/server/egressselector"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/component-base/metrics/legacyregistry"
)

const (
	keepaliveTime    = 30 * time.Second
	keepaliveTimeout = 10 * time.Second

	dialTimeout = 20 * time.Second

	dbMetricsMonitorJitter = 0.5
)

// p66
func init() {
	legacyregistry.RawMustRegister(grpcprom.DefaultClientMetrics)
	dbMetricsMonitors = make(map[string]struct{})
}

func newETCD3HealthCheck(c storagebackend.Config) (func() error, error) {
	clientValue := &atomic.Value{}

	clientErrMsg := &atomic.Value{}
	clientErrMsg.Store("etcd client connection not yet established")

	// 后台运行这个函数，如果失败每秒尝试一次；如果成功，将etcdclient的实例对象存起来
	go wait.PollUntil(time.Second, func() (bool, error) {
		client, err := newETCD3Client(c.Transport)
		if err != nil {
			clientErrMsg.Store(err.Error())
			return false, nil
		}
		clientValue.Store(client)
		clientErrMsg.Store("")
		return true, nil
	}, wait.NeverStop)

	return func() error {
		if errMsg := clientErrMsg.Load().(string); len(errMsg) > 0 {
			return fmt.Errorf(errMsg)
		}
		client := clientValue.Load().(*clientv3.Client)
		healthcheckTimeout := storagebackend.DefaultHealthcheckTimeout
		if c.HealthcheckTimeout != time.Duration(0) {
			healthcheckTimeout = c.HealthcheckTimeout
		}
		ctx, cancel := context.WithTimeout(context.Background(), healthcheckTimeout)
		defer cancel()
		_, err := client.Get(ctx, path.Join("/", c.Prefix, "health"))
		if err == nil {
			return nil
		}
		return fmt.Errorf("error getting data from etcd: %v", err)
	}, nil
}

func newETCD3Client(c storagebackend.TransportConfig) (*clientv3.Client, error) {
	tlsInfo := transport.TLSInfo{
		CertFile:      c.CertFile,
		KeyFile:       c.KeyFile,
		TrustedCAFile: c.TrustedCAFile,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return nil, err
	}
	if len(c.CertFile) == 0 && len(c.KeyFile) == 0 && len(c.TrustedCAFile) == 0 {
		tlsConfig = nil
	}
	networkContext := egressselector.Etcd.AsNetworkContext()
	var egressDialer utilnet.DialFunc
	if c.EgressLookup != nil {
		egressDialer, err = c.EgressLookup(networkContext)
		if err != nil {
			return nil, err
		}
	}
	dialOptions := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithChainUnaryInterceptor(grpcprom.UnaryClientInterceptor),
		grpc.WithChainStreamInterceptor(grpcprom.StreamClientInterceptor),
	}
	if utilfeature.DefaultFeatureGate.Enabled(genericfeatures.APIServerTracing) {
		tracingOpts := []otelgrpc.Option{
			otelgrpc.WithPropagators(traces.Propagators()),
		}
		if c.TracerProvider != nil {
			tracingOpts = append(tracingOpts, otelgrpc.WithTracerProvider(*c.TracerProvider))
		}
		dialOptions = append(dialOptions,
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor(tracingOpts...)),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor(tracingOpts...)))
	}
	if egressDialer != nil {
		dialer := func(ctx context.Context, addr string) (net.Conn, error) {
			if strings.Contains(addr, "//") {
				// etcd client prior to 3.5 passed URLs to dialer, normalize to address
				u, err := url.Parse(addr)
				if err != nil {
					return nil, err
				}
				addr = u.Host
			}
			return egressDialer(ctx, "tcp", addr)
		}
		dialOptions = append(dialOptions, grpc.WithContextDialer(dialer))
	}
	cfg := clientv3.Config{
		DialTimeout:          dialTimeout,
		DialKeepAliveTime:    keepaliveTime,
		DialKeepAliveTimeout: keepaliveTimeout,
		DialOptions:          dialOptions,
		Endpoints:            c.ServerList,
		TLS:                  tlsConfig,
	}
	return clientv3.New(cfg)
}

// p194
var (
	dbMetricsMonitors map[string]struct{}
)
