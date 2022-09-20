package value

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/errors"
)

func init() {
	RegisterMetrics()
}

type Context interface {
	AuthenticatedData() []byte
}

type Transformer interface {
	TransformFromStorage(ctx context.Context, data []byte, dataCtx Context) (out []byte, stale bool, err error)
	TransformToStorage(ctx context.Context, data []byte, dataCtx Context) (out []byte, err error)
}

type identityTransformer struct{}

var IdentityTransformer Transformer = identityTransformer{}

func (identityTransformer) TransformFromStorage(ctx context.Context, data []byte, dataCtx Context) ([]byte, bool, error) {
	return data, false, nil
}
func (identityTransformer) TransformToStorage(ctx context.Context, data []byte, dataCtx Context) ([]byte, error) {
	return data, nil
}

type DefaultContext []byte

func (c DefaultContext) AuthenticatedData() []byte { return []byte(c) }

type MutableTransformer struct {
	lock        sync.RWMutex
	transformer Transformer
}

func NewMutableTransformer(transformer Transformer) *MutableTransformer {
	return &MutableTransformer{transformer: transformer}
}

func (t *MutableTransformer) Set(transformer Transformer) {
	t.lock.Lock()
	t.transformer = transformer
	t.lock.Unlock()
}

func (t *MutableTransformer) TransformFromStorage(ctx context.Context, data []byte, dataCtx Context) (out []byte, stale bool, err error) {
	t.lock.RLock()
	transformer := t.transformer
	t.lock.RUnlock()
	return transformer.TransformFromStorage(ctx, data, dataCtx)
}
func (t *MutableTransformer) TransformToStorage(ctx context.Context, data []byte, dataCtx Context) (out []byte, err error) {
	t.lock.RLock()
	transformer := t.transformer
	t.lock.RUnlock()
	return transformer.TransformToStorage(ctx, data, dataCtx)
}

type PrefixTransformer struct {
	Prefix      []byte
	Transformer Transformer
}

type prefixTransformers struct {
	transformers []PrefixTransformer
	err          error
}

var _ Transformer = &prefixTransformers{}

func NewPrefixTransformers(err error, transformers ...PrefixTransformer) Transformer {
	if err == nil {
		err = fmt.Errorf("the provided value does not match any of the supported transformers")
	}
	return &prefixTransformers{
		transformers: transformers,
		err:          err,
	}
}

func (t *prefixTransformers) TransformFromStorage(ctx context.Context, data []byte, dataCtx Context) ([]byte, bool, error) {
	start := time.Now()
	var errs []error
	for i, transformer := range t.transformers {
		if bytes.HasPrefix(data, transformer.Prefix) {
			result, stale, err := transformer.Transformer.TransformFromStorage(ctx, data[len(transformer.Prefix):], dataCtx)
			// To migrate away from encryption, user can specify an identity transformer higher up
			// (in the config file) than the encryption transformer. In that scenario, the identity transformer needs to
			// identify (during reads from disk) whether the data being read is encrypted or not. If the data is encrypted,
			// it shall throw an error, but that error should not prevent the next subsequent transformer from being tried.
			if len(transformer.Prefix) == 0 && err != nil {
				continue
			}
			if len(transformer.Prefix) == 0 {
				RecordTransformation("from_storage", "identity", start, err)
			} else {
				RecordTransformation("from_storage", string(transformer.Prefix), start, err)
			}

			if err != nil {
				errs = append(errs, err)
				continue
			}

			return result, stale || i != 0, err
		}
	}
	if err := errors.Reduce(errors.NewAggregate(errs)); err != nil {
		return nil, false, err
	}
	RecordTransformation("from_storage", "unknown", start, t.err)
	return nil, false, t.err
}

// TransformToStorage uses the first transformer and adds its prefix to the data.
func (t *prefixTransformers) TransformToStorage(ctx context.Context, data []byte, dataCtx Context) ([]byte, error) {
	start := time.Now()
	transformer := t.transformers[0]
	prefixedData := make([]byte, len(transformer.Prefix), len(data)+len(transformer.Prefix))
	copy(prefixedData, transformer.Prefix)
	result, err := transformer.Transformer.TransformToStorage(ctx, data, dataCtx)
	RecordTransformation("to_storage", string(transformer.Prefix), start, err)
	if err != nil {
		return nil, err
	}
	prefixedData = append(prefixedData, result...)
	return prefixedData, nil
}
