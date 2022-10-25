package v1

import "k8s.io/client-go/informers/internalinterfaces"

type Interface interface {
	MutatingWebhookConfigurations() MutatingWebhookConfigurationInformer
	ValidatingWebhookConfigurations() ValidatingWebhookConfigurationInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

func (v *version) MutatingWebhookConfigurations() MutatingWebhookConfigurationInformer {
	return &mutatingWebhookConfigurationInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

func (v *version) ValidatingWebhookConfigurations() ValidatingWebhookConfigurationInformer {
	return &validatingWebhookConfigurationInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}
