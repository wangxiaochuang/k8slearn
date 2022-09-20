package storagebackend

import (
	"time"

	"go.opentelemetry.io/otel/trace"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/server/egressselector"
	"k8s.io/apiserver/pkg/storage/etcd3"
	"k8s.io/apiserver/pkg/storage/value"
	flowcontrolrequest "k8s.io/apiserver/pkg/util/flowcontrol/request"
)

const (
	StorageTypeUnset = ""
	StorageTypeETCD2 = "etcd2"
	StorageTypeETCD3 = "etcd3"

	DefaultCompactInterval      = 5 * time.Minute
	DefaultDBMetricPollInterval = 30 * time.Second
	DefaultHealthcheckTimeout   = 2 * time.Second
)

type TransportConfig struct {
	ServerList     []string
	KeyFile        string
	CertFile       string
	TrustedCAFile  string
	EgressLookup   egressselector.Lookup
	TracerProvider *trace.TracerProvider
}

type Config struct {
	Type      string
	Prefix    string
	Transport TransportConfig

	Paging          bool
	Codec           runtime.Codec
	EncodeVersioner runtime.GroupVersioner
	Transformer     value.Transformer

	CompactionInterval    time.Duration
	CountMetricPollPeriod time.Duration
	DBMetricPollInterval  time.Duration
	HealthcheckTimeout    time.Duration

	LeaseManagerConfig etcd3.LeaseManagerConfig

	StorageObjectCountTracker flowcontrolrequest.StorageObjectCountTracker
}

type ConfigForResource struct {
	Config
	GroupResource schema.GroupResource
}

func (config *Config) ForResource(resource schema.GroupResource) *ConfigForResource {
	return &ConfigForResource{
		Config:        *config,
		GroupResource: resource,
	}
}

func NewDefaultConfig(prefix string, codec runtime.Codec) *Config {
	return &Config{
		Paging:               true,
		Prefix:               prefix,
		Codec:                codec,
		CompactionInterval:   DefaultCompactInterval,
		DBMetricPollInterval: DefaultDBMetricPollInterval,
		HealthcheckTimeout:   DefaultHealthcheckTimeout,
		LeaseManagerConfig:   etcd3.NewDefaultLeaseManagerConfig(),
	}
}
