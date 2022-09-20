package storage

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

type Versioner interface {
	UpdateObject(obj runtime.Object, resourceVersion uint64) error
	UpdateList(obj runtime.Object, resourceVersion uint64, continueValue string, remainingItemCount *int64) error
	PrepareObjectForStorage(obj runtime.Object) error
	ObjectResourceVersion(obj runtime.Object) (uint64, error)
	ParseResourceVersion(resourceVersion string) (uint64, error)
}

type ResponseMeta struct {
	TTL             int64
	ResourceVersion uint64
}

type IndexerFunc func(obj runtime.Object) string

type IndexerFuncs map[string]IndexerFunc

var Everything = SelectionPredicate{
	Label: labels.Everything(),
	Field: fields.Everything(),
}

type MatchValue struct {
	IndexName string
	Value     string
}

type UpdateFunc func(input runtime.Object, res ResponseMeta) (output runtime.Object, ttl *uint64, err error)

type ValidateObjectFunc func(ctx context.Context, obj runtime.Object) error

func ValidateAllObjectFunc(ctx context.Context, obj runtime.Object) error {
	return nil
}

type Preconditions struct {
	UID             *types.UID `json:"uid,omitempty"`
	ResourceVersion *string    `json:"resourceVersion,omitempty"`
}

func NewUIDPreconditions(uid string) *Preconditions {
	u := types.UID(uid)
	return &Preconditions{UID: &u}
}

func (p *Preconditions) Check(key string, obj runtime.Object) error {
	if p == nil {
		return nil
	}
	objMeta, err := meta.Accessor(obj)
	if err != nil {
		return NewInternalErrorf(
			"can't enforce preconditions %v on un-introspectable object %v, got error: %v",
			*p,
			obj,
			err)
	}
	if p.UID != nil && *p.UID != objMeta.GetUID() {
		err := fmt.Sprintf(
			"Precondition failed: UID in precondition: %v, UID in object meta: %v",
			*p.UID,
			objMeta.GetUID())
		return NewInvalidObjError(key, err)
	}
	if p.ResourceVersion != nil && *p.ResourceVersion != objMeta.GetResourceVersion() {
		err := fmt.Sprintf(
			"Precondition failed: ResourceVersion in precondition: %v, ResourceVersion in object meta: %v",
			*p.ResourceVersion,
			objMeta.GetResourceVersion())
		return NewInvalidObjError(key, err)
	}
	return nil
}

type Interface interface {
	Versioner() Versioner
	Create(ctx context.Context, key string, obj, out runtime.Object, ttl uint64) error
	Delete(
		ctx context.Context, key string, out runtime.Object, preconditions *Preconditions,
		validateDeletion ValidateObjectFunc, cachedExistingObject runtime.Object) error
	Watch(ctx context.Context, key string, opts ListOptions) (watch.Interface, error)
	Get(ctx context.Context, key string, opts GetOptions, objPtr runtime.Object) error
	GetList(ctx context.Context, key string, opts ListOptions, listObj runtime.Object) error
	GuaranteedUpdate(
		ctx context.Context, key string, ptrToType runtime.Object, ignoreNotFound bool,
		preconditions *Preconditions, tryUpdate UpdateFunc, cachedExistingObject runtime.Object) error
	Count(key string) (int64, error)
}

type GetOptions struct {
	IgnoreNotFound  bool
	ResourceVersion string
}

type ListOptions struct {
	ResourceVersion      string
	ResourceVersionMatch metav1.ResourceVersionMatch
	Predicate            SelectionPredicate
	Recursive            bool
	ProgressNotify       bool
}
