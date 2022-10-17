package factory

import (
	"fmt"

	"k8s.io/apiserver/pkg/storage/storagebackend"
)

type DestroyFunc func()

func CreateHealthCheck(c storagebackend.Config) (func() error, error) {
	switch c.Type {
	case storagebackend.StorageTypeETCD2:
		return nil, fmt.Errorf("%s is no longer a supported storage backend", c.Type)
	case storagebackend.StorageTypeUnset, storagebackend.StorageTypeETCD3:
		return newETCD3HealthCheck(c)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", c.Type)
	}
}
