package v1

import (
	v1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

type ValidatingWebhookConfigurationLister interface {
	List(selector labels.Selector) (ret []*v1.ValidatingWebhookConfiguration, err error)
	Get(name string) (*v1.ValidatingWebhookConfiguration, error)
	ValidatingWebhookConfigurationListerExpansion
}

type validatingWebhookConfigurationLister struct {
	indexer cache.Indexer
}

func NewValidatingWebhookConfigurationLister(indexer cache.Indexer) ValidatingWebhookConfigurationLister {
	return &validatingWebhookConfigurationLister{indexer: indexer}
}

func (s *validatingWebhookConfigurationLister) List(selector labels.Selector) (ret []*v1.ValidatingWebhookConfiguration, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.ValidatingWebhookConfiguration))
	})
	return ret, err
}

func (s *validatingWebhookConfigurationLister) Get(name string) (*v1.ValidatingWebhookConfiguration, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("validatingwebhookconfiguration"), name)
	}
	return obj.(*v1.ValidatingWebhookConfiguration), nil
}
