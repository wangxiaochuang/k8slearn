package v1

import (
	v1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

type MutatingWebhookConfigurationLister interface {
	List(selector labels.Selector) (ret []*v1.MutatingWebhookConfiguration, err error)
	Get(name string) (*v1.MutatingWebhookConfiguration, error)
	MutatingWebhookConfigurationListerExpansion
}

type mutatingWebhookConfigurationLister struct {
	indexer cache.Indexer
}

func NewMutatingWebhookConfigurationLister(indexer cache.Indexer) MutatingWebhookConfigurationLister {
	return &mutatingWebhookConfigurationLister{indexer: indexer}
}

func (s *mutatingWebhookConfigurationLister) List(selector labels.Selector) (ret []*v1.MutatingWebhookConfiguration, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.MutatingWebhookConfiguration))
	})
	return ret, err
}

func (s *mutatingWebhookConfigurationLister) Get(name string) (*v1.MutatingWebhookConfiguration, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("mutatingwebhookconfiguration"), name)
	}
	return obj.(*v1.MutatingWebhookConfiguration), nil
}
