package v1

import (
	"context"
	time "time"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	internalinterfaces "k8s.io/client-go/informers/internalinterfaces"
	kubernetes "k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/admissionregistration/v1"
	cache "k8s.io/client-go/tools/cache"
)

type ValidatingWebhookConfigurationInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.ValidatingWebhookConfigurationLister
}

type validatingWebhookConfigurationInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

func NewValidatingWebhookConfigurationInformer(client kubernetes.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredValidatingWebhookConfigurationInformer(client, resyncPeriod, indexers, nil)
}

func NewFilteredValidatingWebhookConfigurationInformer(client kubernetes.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AdmissionregistrationV1().ValidatingWebhookConfigurations().Watch(context.TODO(), options)
			},
		},
		&admissionregistrationv1.ValidatingWebhookConfiguration{},
		resyncPeriod,
		indexers,
	)
}

func (f *validatingWebhookConfigurationInformer) defaultInformer(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredValidatingWebhookConfigurationInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *validatingWebhookConfigurationInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&admissionregistrationv1.ValidatingWebhookConfiguration{}, f.defaultInformer)
}

func (f *validatingWebhookConfigurationInformer) Lister() v1.ValidatingWebhookConfigurationLister {
	return v1.NewValidatingWebhookConfigurationLister(f.Informer().GetIndexer())
}
