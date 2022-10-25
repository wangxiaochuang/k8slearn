package v1

import (
	"context"
	"time"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers/internalinterfaces"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/admissionregistration/v1"
	"k8s.io/client-go/tools/cache"
)

type MutatingWebhookConfigurationInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.MutatingWebhookConfigurationLister
}

type mutatingWebhookConfigurationInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

func NewMutatingWebhookConfigurationInformer(client kubernetes.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredMutatingWebhookConfigurationInformer(client, resyncPeriod, indexers, nil)
}

func NewFilteredMutatingWebhookConfigurationInformer(client kubernetes.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AdmissionregistrationV1().MutatingWebhookConfigurations().List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AdmissionregistrationV1().MutatingWebhookConfigurations().Watch(context.TODO(), options)
			},
		},
		&admissionregistrationv1.MutatingWebhookConfiguration{},
		resyncPeriod,
		indexers,
	)
}

func (f *mutatingWebhookConfigurationInformer) defaultInformer(client kubernetes.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredMutatingWebhookConfigurationInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *mutatingWebhookConfigurationInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&admissionregistrationv1.MutatingWebhookConfiguration{}, f.defaultInformer)
}

func (f *mutatingWebhookConfigurationInformer) Lister() v1.MutatingWebhookConfigurationLister {
	return v1.NewMutatingWebhookConfigurationLister(f.Informer().GetIndexer())
}
