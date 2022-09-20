package storage

import (
	"crypto/tls"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/storage/storagebackend"
)

type Backend struct {
	// the url of storage backend like: https://etcd.domain:2379
	Server    string
	TLSConfig *tls.Config
}

type StorageFactory interface {
	NewConfig(groupResource schema.GroupResource) (*storagebackend.ConfigForResource, error)
	ResourcePrefix(groupResource schema.GroupResource) string
	Backends() []Backend
}

type DefaultStorageFactory struct {
}

var _ StorageFactory = &DefaultStorageFactory{}

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
