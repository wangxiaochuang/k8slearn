package informers

import "reflect"

type SharedInformerFactory interface {
	WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool
}
