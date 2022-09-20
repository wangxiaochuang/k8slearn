package generic

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/apiserver/pkg/storage/storagebackend/factory"
	"k8s.io/client-go/tools/cache"
)

type StorageDecorator func(
	config *storagebackend.ConfigForResource,
	resourcePrefix string,
	keyFunc func(obj runtime.Object) (string, error),
	newFunc func() runtime.Object,
	newListFunc func() runtime.Object,
	getAttrsFunc storage.AttrFunc,
	trigger storage.IndexerFuncs,
	indexers *cache.Indexers) (storage.Interface, factory.DestroyFunc, error)

func UndecoratedStorage(
	config *storagebackend.ConfigForResource,
	resourcePrefix string,
	keyFunc func(obj runtime.Object) (string, error),
	newFunc func() runtime.Object,
	newListFunc func() runtime.Object,
	getAttrsFunc storage.AttrFunc,
	trigger storage.IndexerFuncs,
	indexers *cache.Indexers) (storage.Interface, factory.DestroyFunc, error) {
	return NewRawStorage(config, newFunc)
}

func NewRawStorage(config *storagebackend.ConfigForResource, newFunc func() runtime.Object) (storage.Interface, factory.DestroyFunc, error) {
	panic("UndecoratedStorage")
}
