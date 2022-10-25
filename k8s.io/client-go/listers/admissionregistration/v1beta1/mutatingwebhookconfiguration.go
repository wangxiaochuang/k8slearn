package v1beta1

import (
	v1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// MutatingWebhookConfigurationLister helps list MutatingWebhookConfigurations.
// All objects returned here must be treated as read-only.
type MutatingWebhookConfigurationLister interface {
	// List lists all MutatingWebhookConfigurations in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1beta1.MutatingWebhookConfiguration, err error)
	// Get retrieves the MutatingWebhookConfiguration from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1beta1.MutatingWebhookConfiguration, error)
	MutatingWebhookConfigurationListerExpansion
}

// mutatingWebhookConfigurationLister implements the MutatingWebhookConfigurationLister interface.
type mutatingWebhookConfigurationLister struct {
	indexer cache.Indexer
}

// NewMutatingWebhookConfigurationLister returns a new MutatingWebhookConfigurationLister.
func NewMutatingWebhookConfigurationLister(indexer cache.Indexer) MutatingWebhookConfigurationLister {
	return &mutatingWebhookConfigurationLister{indexer: indexer}
}

// List lists all MutatingWebhookConfigurations in the indexer.
func (s *mutatingWebhookConfigurationLister) List(selector labels.Selector) (ret []*v1beta1.MutatingWebhookConfiguration, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.MutatingWebhookConfiguration))
	})
	return ret, err
}

// Get retrieves the MutatingWebhookConfiguration from the index for a given name.
func (s *mutatingWebhookConfigurationLister) Get(name string) (*v1beta1.MutatingWebhookConfiguration, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1beta1.Resource("mutatingwebhookconfiguration"), name)
	}
	return obj.(*v1beta1.MutatingWebhookConfiguration), nil
}
