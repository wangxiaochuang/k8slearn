package v1beta1

import (
	"context"
	time "time"

	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	internalinterfaces "k8s.io/client-go/informers/internalinterfaces"
	kubernetes "k8s.io/client-go/kubernetes"
	v1beta1 "k8s.io/client-go/listers/admissionregistration/v1beta1"
	cache "k8s.io/client-go/tools/cache"
)

// MutatingWebhookConfigurationInformer provides access to a shared informer and lister for
// MutatingWebhookConfigurations.
type MutatingWebhookConfigurationInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1beta1.MutatingWebhookConfigurationLister
}

type mutatingWebhookConfigurationInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewMutatingWebhookConfigurationInformer constructs a new informer for MutatingWebhookConfiguration type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewMutatingWebhookConfigurationInformer(client kubernetes.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredMutatingWebhookConfigurationInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredMutatingWebhookConfigurationInformer constructs a new informer for MutatingWebhookConfiguration type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredMutatingWebhookConfigurationInformer(client kubernetes.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AdmissionregistrationV1beta1().MutatingWebhookConfigurations().Watch(context.TODO(), options)
			},
		},
		&admissionregistrationv1beta1.MutatingWebhookConfiguration{},
		resyncPeriod,
		indexers,
	)
}

func (f *mutatingWebhookConfigurationInformer) defaultInformer(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredMutatingWebhookConfigurationInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *mutatingWebhookConfigurationInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&admissionregistrationv1beta1.MutatingWebhookConfiguration{}, f.defaultInformer)
}

func (f *mutatingWebhookConfigurationInformer) Lister() v1beta1.MutatingWebhookConfigurationLister {
	return v1beta1.NewMutatingWebhookConfigurationLister(f.Informer().GetIndexer())
}
