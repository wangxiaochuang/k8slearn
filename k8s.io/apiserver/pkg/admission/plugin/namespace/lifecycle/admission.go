package lifecycle

import (
	"context"
	"fmt"
	"io"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilcache "k8s.io/apimachinery/pkg/util/cache"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/admission/initializer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/utils/clock"
)

const (
	PluginName           = "NamespaceLifecycle"
	forceLiveLookupTTL   = 30 * time.Second
	missingNamespaceWait = 50 * time.Millisecond
)

func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return NewLifecycle(sets.NewString(metav1.NamespaceDefault, metav1.NamespaceSystem, metav1.NamespacePublic))
	})
}

type Lifecycle struct {
	*admission.Handler
	client             kubernetes.Interface
	immortalNamespaces sets.String
	namespaceLister    corelisters.NamespaceLister
	// forceLiveLookupCache holds a list of entries for namespaces that we have a strong reason to believe are stale in our local cache.
	// if a namespace is in this cache, then we will ignore our local state and always fetch latest from api server.
	forceLiveLookupCache *utilcache.LRUExpireCache
}

var _ = initializer.WantsExternalKubeInformerFactory(&Lifecycle{})
var _ = initializer.WantsExternalKubeClientSet(&Lifecycle{})

func (l *Lifecycle) Admit(ctx context.Context, a admission.Attributes, o admission.ObjectInterfaces) error {
	panic("not implemented")
}

func NewLifecycle(immortalNamespaces sets.String) (*Lifecycle, error) {
	return newLifecycleWithClock(immortalNamespaces, clock.RealClock{})
}

func newLifecycleWithClock(immortalNamespaces sets.String, clock utilcache.Clock) (*Lifecycle, error) {
	forceLiveLookupCache := utilcache.NewLRUExpireCacheWithClock(100, clock)
	return &Lifecycle{
		Handler:              admission.NewHandler(admission.Create, admission.Update, admission.Delete),
		immortalNamespaces:   immortalNamespaces,
		forceLiveLookupCache: forceLiveLookupCache,
	}, nil
}

func (l *Lifecycle) SetExternalKubeInformerFactory(f informers.SharedInformerFactory) {
	namespaceInformer := f.Core().V1().Namespaces()
	l.namespaceLister = namespaceInformer.Lister()
	l.SetReadyFunc(namespaceInformer.Informer().HasSynced)
}

func (l *Lifecycle) SetExternalKubeClientSet(client kubernetes.Interface) {
	l.client = client
}

func (l *Lifecycle) ValidateInitialization() error {
	if l.namespaceLister == nil {
		return fmt.Errorf("missing namespaceLister")
	}
	if l.client == nil {
		return fmt.Errorf("missing client")
	}
	return nil
}

var accessReviewResources = map[schema.GroupResource]bool{
	{Group: "authorization.k8s.io", Resource: "localsubjectaccessreviews"}: true,
}

func isAccessReview(a admission.Attributes) bool {
	return accessReviewResources[a.GetResource().GroupResource()]
}
