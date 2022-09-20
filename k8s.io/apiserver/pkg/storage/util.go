package storage

import "k8s.io/apimachinery/pkg/runtime"

type SimpleUpdateFunc func(runtime.Object) (runtime.Object, error)

func SimpleUpdate(fn SimpleUpdateFunc) UpdateFunc {
	return func(input runtime.Object, _ ResponseMeta) (runtime.Object, *uint64, error) {
		out, err := fn(input)
		return out, nil, err
	}
}

func EverythingFunc(runtime.Object) bool {
	return true
}
