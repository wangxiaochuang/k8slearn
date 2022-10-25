package v1beta1

import (
	v1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ValidatingWebhookConfigurationLister helps list ValidatingWebhookConfigurations.
// All objects returned here must be treated as read-only.
type ValidatingWebhookConfigurationLister interface {
	// List lists all ValidatingWebhookConfigurations in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1beta1.ValidatingWebhookConfiguration, err error)
	// Get retrieves the ValidatingWebhookConfiguration from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1beta1.ValidatingWebhookConfiguration, error)
	ValidatingWebhookConfigurationListerExpansion
}

// validatingWebhookConfigurationLister implements the ValidatingWebhookConfigurationLister interface.
type validatingWebhookConfigurationLister struct {
	indexer cache.Indexer
}

// NewValidatingWebhookConfigurationLister returns a new ValidatingWebhookConfigurationLister.
func NewValidatingWebhookConfigurationLister(indexer cache.Indexer) ValidatingWebhookConfigurationLister {
	return &validatingWebhookConfigurationLister{indexer: indexer}
}

// List lists all ValidatingWebhookConfigurations in the indexer.
func (s *validatingWebhookConfigurationLister) List(selector labels.Selector) (ret []*v1beta1.ValidatingWebhookConfiguration, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.ValidatingWebhookConfiguration))
	})
	return ret, err
}

// Get retrieves the ValidatingWebhookConfiguration from the index for a given name.
func (s *validatingWebhookConfigurationLister) Get(name string) (*v1beta1.ValidatingWebhookConfiguration, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1beta1.Resource("validatingwebhookconfiguration"), name)
	}
	return obj.(*v1beta1.ValidatingWebhookConfiguration), nil
}
