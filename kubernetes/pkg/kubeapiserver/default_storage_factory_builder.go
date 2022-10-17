package kubeapiserver

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	serveroptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/apiserver/pkg/server/resourceconfig"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/apis/apps"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/events"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/apis/networking"
	"k8s.io/kubernetes/pkg/apis/policy"
	apisstorage "k8s.io/kubernetes/pkg/apis/storage"
)

var SpecialDefaultResourcePrefixes = map[schema.GroupResource]string{
	{Group: "", Resource: "replicationcontrollers"}:        "controllers",
	{Group: "", Resource: "endpoints"}:                     "services/endpoints",
	{Group: "", Resource: "nodes"}:                         "minions",
	{Group: "", Resource: "services"}:                      "services/specs",
	{Group: "extensions", Resource: "ingresses"}:           "ingress",
	{Group: "networking.k8s.io", Resource: "ingresses"}:    "ingress",
	{Group: "extensions", Resource: "podsecuritypolicies"}: "podsecuritypolicy",
	{Group: "policy", Resource: "podsecuritypolicies"}:     "podsecuritypolicy",
}

func DefaultWatchCacheSizes() map[schema.GroupResource]int {
	return map[schema.GroupResource]int{
		{Resource: "events"}:                         0,
		{Group: "events.k8s.io", Resource: "events"}: 0,
	}
}

func NewStorageFactoryConfig() *StorageFactoryConfig {
	resources := []schema.GroupVersionResource{
		apisstorage.Resource("csistoragecapacities").WithVersion("v1beta1"),
	}

	return &StorageFactoryConfig{
		Serializer:                legacyscheme.Codecs,
		DefaultResourceEncoding:   serverstorage.NewDefaultResourceEncodingConfig(legacyscheme.Scheme),
		ResourceEncodingOverrides: resources,
	}
}

type StorageFactoryConfig struct {
	StorageConfig                    storagebackend.Config
	APIResourceConfig                *serverstorage.ResourceConfig
	DefaultResourceEncoding          *serverstorage.DefaultResourceEncodingConfig
	DefaultStorageMediaType          string
	Serializer                       runtime.StorageSerializer
	ResourceEncodingOverrides        []schema.GroupVersionResource
	EtcdServersOverrides             []string
	EncryptionProviderConfigFilepath string
}

func (c *StorageFactoryConfig) Complete(etcdOptions *serveroptions.EtcdOptions) (*completedStorageFactoryConfig, error) {
	c.StorageConfig = etcdOptions.StorageConfig
	c.DefaultStorageMediaType = etcdOptions.DefaultStorageMediaType
	c.EtcdServersOverrides = etcdOptions.EtcdServersOverrides
	c.EncryptionProviderConfigFilepath = etcdOptions.EncryptionProviderConfigFilepath
	return &completedStorageFactoryConfig{c}, nil
}

type completedStorageFactoryConfig struct {
	*StorageFactoryConfig
}

func (c *completedStorageFactoryConfig) New() (*serverstorage.DefaultStorageFactory, error) {
	resourceEncodingConfig := resourceconfig.MergeResourceEncodingConfigs(c.DefaultResourceEncoding, c.ResourceEncodingOverrides)
	storageFactory := serverstorage.NewDefaultStorageFactory(
		c.StorageConfig,
		c.DefaultStorageMediaType,
		c.Serializer,
		resourceEncodingConfig,
		c.APIResourceConfig,
		SpecialDefaultResourcePrefixes)
	storageFactory.AddCohabitatingResources(networking.Resource("networkpolicies"), extensions.Resource("networkpolicies"))
	storageFactory.AddCohabitatingResources(apps.Resource("deployments"), extensions.Resource("deployments"))
	storageFactory.AddCohabitatingResources(apps.Resource("daemonsets"), extensions.Resource("daemonsets"))
	storageFactory.AddCohabitatingResources(apps.Resource("replicasets"), extensions.Resource("replicasets"))
	storageFactory.AddCohabitatingResources(api.Resource("events"), events.Resource("events"))
	storageFactory.AddCohabitatingResources(api.Resource("replicationcontrollers"), extensions.Resource("replicationcontrollers")) // to make scale subresources equivalent
	storageFactory.AddCohabitatingResources(policy.Resource("podsecuritypolicies"), extensions.Resource("podsecuritypolicies"))
	storageFactory.AddCohabitatingResources(networking.Resource("ingresses"), extensions.Resource("ingresses"))
	for _, override := range c.EtcdServersOverrides {
		panic(fmt.Sprintf("not implemented: %s", override))
	}
	if len(c.EncryptionProviderConfigFilepath) != 0 {
		panic("not implemented")
	}
	return storageFactory, nil
}
