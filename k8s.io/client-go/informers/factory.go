package informers

import (
	"reflect"
	"sync"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers/admissionregistration"
	"k8s.io/client-go/informers/apiserverinternal"
	"k8s.io/client-go/informers/apps"
	"k8s.io/client-go/informers/autoscaling"
	"k8s.io/client-go/informers/batch"
	"k8s.io/client-go/informers/certificates"
	"k8s.io/client-go/informers/coordination"
	"k8s.io/client-go/informers/core"
	"k8s.io/client-go/informers/discovery"
	"k8s.io/client-go/informers/events"
	"k8s.io/client-go/informers/extensions"
	flowcontrol "k8s.io/client-go/informers/flowcontrol"
	"k8s.io/client-go/informers/internalinterfaces"
	"k8s.io/client-go/informers/networking"
	"k8s.io/client-go/informers/node"
	"k8s.io/client-go/informers/policy"
	"k8s.io/client-go/informers/rbac"
	"k8s.io/client-go/informers/scheduling"
	storage "k8s.io/client-go/informers/storage"
	"k8s.io/client-go/kubernetes"
	cache "k8s.io/client-go/tools/cache"
)

type SharedInformerOption func(*sharedInformerFactory) *sharedInformerFactory

type sharedInformerFactory struct {
	client           kubernetes.Interface
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	lock             sync.Mutex
	defaultResync    time.Duration
	customResync     map[reflect.Type]time.Duration

	informers map[reflect.Type]cache.SharedIndexInformer
	// startedInformers is used for tracking which informers have been started.
	// This allows Start() to be called multiple times safely.
	startedInformers map[reflect.Type]bool
}

func WithCustomResyncConfig(resyncConfig map[v1.Object]time.Duration) SharedInformerOption {
	return func(factory *sharedInformerFactory) *sharedInformerFactory {
		for k, v := range resyncConfig {
			factory.customResync[reflect.TypeOf(k)] = v
		}
		return factory
	}
}

func WithTweakListOptions(tweakListOptions internalinterfaces.TweakListOptionsFunc) SharedInformerOption {
	return func(factory *sharedInformerFactory) *sharedInformerFactory {
		factory.tweakListOptions = tweakListOptions
		return factory
	}
}

func WithNamespace(namespace string) SharedInformerOption {
	return func(factory *sharedInformerFactory) *sharedInformerFactory {
		factory.namespace = namespace
		return factory
	}
}

func NewSharedInformerFactory(client kubernetes.Interface, defaultResync time.Duration) SharedInformerFactory {
	return NewSharedInformerFactoryWithOptions(client, defaultResync)
}

func NewFilteredSharedInformerFactory(client kubernetes.Interface, defaultResync time.Duration, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) SharedInformerFactory {
	return NewSharedInformerFactoryWithOptions(client, defaultResync, WithNamespace(namespace), WithTweakListOptions(tweakListOptions))
}

func NewSharedInformerFactoryWithOptions(client kubernetes.Interface, defaultResync time.Duration, options ...SharedInformerOption) SharedInformerFactory {
	factory := &sharedInformerFactory{
		client:           client,
		namespace:        v1.NamespaceAll,
		defaultResync:    defaultResync,
		informers:        make(map[reflect.Type]cache.SharedIndexInformer),
		startedInformers: make(map[reflect.Type]bool),
		customResync:     make(map[reflect.Type]time.Duration),
	}

	// Apply all options
	for _, opt := range options {
		factory = opt(factory)
	}

	return factory
}

func (f *sharedInformerFactory) Start(stopCh <-chan struct{}) {
	f.lock.Lock()
	defer f.lock.Unlock()

	for informerType, informer := range f.informers {
		if !f.startedInformers[informerType] {
			go informer.Run(stopCh)
			f.startedInformers[informerType] = true
		}
	}
}

func (f *sharedInformerFactory) WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool {
	informers := func() map[reflect.Type]cache.SharedIndexInformer {
		f.lock.Lock()
		defer f.lock.Unlock()

		informers := map[reflect.Type]cache.SharedIndexInformer{}
		for informerType, informer := range f.informers {
			if f.startedInformers[informerType] {
				informers[informerType] = informer
			}
		}
		return informers
	}()

	res := map[reflect.Type]bool{}
	for informType, informer := range informers {
		res[informType] = cache.WaitForCacheSync(stopCh, informer.HasSynced)
	}
	return res
}

func (f *sharedInformerFactory) InformerFor(obj runtime.Object, newFunc internalinterfaces.NewInformerFunc) cache.SharedIndexInformer {
	f.lock.Lock()
	defer f.lock.Unlock()

	informerType := reflect.TypeOf(obj)
	informer, exists := f.informers[informerType]
	if exists {
		return informer
	}

	resyncPeriod, exists := f.customResync[informerType]
	if !exists {
		resyncPeriod = f.defaultResync
	}

	informer = newFunc(f.client, resyncPeriod)
	f.informers[informerType] = informer

	return informer
}

type SharedInformerFactory interface {
	internalinterfaces.SharedInformerFactory
	ForResource(resource schema.GroupVersionResource) (GenericInformer, error)
	WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool

	Admissionregistration() admissionregistration.Interface
	Internal() apiserverinternal.Interface
	Apps() apps.Interface
	Autoscaling() autoscaling.Interface
	Batch() batch.Interface
	Certificates() certificates.Interface
	Coordination() coordination.Interface
	Core() core.Interface
	Discovery() discovery.Interface
	Events() events.Interface
	Extensions() extensions.Interface
	Flowcontrol() flowcontrol.Interface
	Networking() networking.Interface
	Node() node.Interface
	Policy() policy.Interface
	Rbac() rbac.Interface
	Scheduling() scheduling.Interface
	Storage() storage.Interface
}

func (f *sharedInformerFactory) Admissionregistration() admissionregistration.Interface {
	return admissionregistration.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Internal() apiserverinternal.Interface {
	return apiserverinternal.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Apps() apps.Interface {
	return apps.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Autoscaling() autoscaling.Interface {
	return autoscaling.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Batch() batch.Interface {
	return batch.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Certificates() certificates.Interface {
	return certificates.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Coordination() coordination.Interface {
	return coordination.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Core() core.Interface {
	return core.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Discovery() discovery.Interface {
	return discovery.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Events() events.Interface {
	return events.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Extensions() extensions.Interface {
	return extensions.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Flowcontrol() flowcontrol.Interface {
	return flowcontrol.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Networking() networking.Interface {
	return networking.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Node() node.Interface {
	return node.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Policy() policy.Interface {
	return policy.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Rbac() rbac.Interface {
	return rbac.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Scheduling() scheduling.Interface {
	return scheduling.New(f, f.namespace, f.tweakListOptions)
}

func (f *sharedInformerFactory) Storage() storage.Interface {
	return storage.New(f, f.namespace, f.tweakListOptions)
}
