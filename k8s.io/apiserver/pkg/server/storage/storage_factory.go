package storage

import (
	"crypto/tls"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/features"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/apiserver/pkg/storage/value"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
)

type Backend struct {
	Server    string
	TLSConfig *tls.Config
}

type StorageFactory interface {
	NewConfig(groupResource schema.GroupResource) (*storagebackend.ConfigForResource, error)
	ResourcePrefix(groupResource schema.GroupResource) string
	Backends() []Backend
}

type DefaultStorageFactory struct {
	StorageConfig storagebackend.Config

	Overrides map[schema.GroupResource]groupResourceOverrides

	DefaultResourcePrefixes map[schema.GroupResource]string

	DefaultMediaType string

	DefaultSerializer runtime.StorageSerializer

	ResourceEncodingConfig ResourceEncodingConfig

	APIResourceConfigSource APIResourceConfigSource

	newStorageCodecFn func(opts StorageCodecConfig) (codec runtime.Codec, encodeVersioner runtime.GroupVersioner, err error)
}

type groupResourceOverrides struct {
	etcdLocation          []string
	etcdPrefix            string
	etcdResourcePrefix    string
	mediaType             string
	serializer            runtime.StorageSerializer
	cohabitatingResources []schema.GroupResource
	encoderDecoratorFn    func(runtime.Encoder) runtime.Encoder
	decoderDecoratorFn    func([]runtime.Decoder) []runtime.Decoder
	transformer           value.Transformer
	disablePaging         bool
}

var _ StorageFactory = &DefaultStorageFactory{}

const AllResources = "*"

func NewDefaultStorageFactory(
	config storagebackend.Config,
	defaultMediaType string,
	defaultSerializer runtime.StorageSerializer,
	resourceEncodingConfig ResourceEncodingConfig,
	resourceConfig APIResourceConfigSource,
	specialDefaultResourcePrefixes map[schema.GroupResource]string,
) *DefaultStorageFactory {
	config.Paging = utilfeature.DefaultFeatureGate.Enabled(features.APIListChunking)
	if len(defaultMediaType) == 0 {
		defaultMediaType = runtime.ContentTypeJSON
	}
	return &DefaultStorageFactory{
		StorageConfig:           config,
		Overrides:               map[schema.GroupResource]groupResourceOverrides{},
		DefaultMediaType:        defaultMediaType,
		DefaultSerializer:       defaultSerializer,
		ResourceEncodingConfig:  resourceEncodingConfig,
		APIResourceConfigSource: resourceConfig,
		DefaultResourcePrefixes: specialDefaultResourcePrefixes,

		newStorageCodecFn: NewStorageCodec,
	}
}

// p220
func (s *DefaultStorageFactory) AddCohabitatingResources(groupResources ...schema.GroupResource) {
	for _, groupResource := range groupResources {
		overrides := s.Overrides[groupResource]
		overrides.cohabitatingResources = groupResources
		s.Overrides[groupResource] = overrides
	}
}

// p254
func (s *DefaultStorageFactory) NewConfig(groupResource schema.GroupResource) (*storagebackend.ConfigForResource, error) {
	panic("not implemented")
}

func (s *DefaultStorageFactory) Backends() []Backend {
	panic("not implemented")
}

func (s *DefaultStorageFactory) ResourcePrefix(groupResource schema.GroupResource) string {
	panic("not implemented")
}
