package kubeapiserver

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
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

type completedStorageFactoryConfig struct {
	*StorageFactoryConfig
}
