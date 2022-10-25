package admissionregistration

import (
	v1 "k8s.io/client-go/informers/admissionregistration/v1"
	v1beta1 "k8s.io/client-go/informers/admissionregistration/v1beta1"
	internalinterfaces "k8s.io/client-go/informers/internalinterfaces"
)

type Interface interface {
	V1() v1.Interface
	V1beta1() v1beta1.Interface
}

type group struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &group{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

func (g *group) V1() v1.Interface {
	return v1.New(g.factory, g.namespace, g.tweakListOptions)
}

func (g *group) V1beta1() v1beta1.Interface {
	return v1beta1.New(g.factory, g.namespace, g.tweakListOptions)
}
