package configuration

import (
	"fmt"
	"sort"
	"sync/atomic"

	v1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/admission/plugin/webhook"
	"k8s.io/apiserver/pkg/admission/plugin/webhook/generic"
	"k8s.io/client-go/informers"
	admissionregistrationlisters "k8s.io/client-go/listers/admissionregistration/v1"
	"k8s.io/client-go/tools/cache"
)

type validatingWebhookConfigurationManager struct {
	configuration              *atomic.Value
	lister                     admissionregistrationlisters.ValidatingWebhookConfigurationLister
	hasSynced                  func() bool
	initialConfigurationSynced *atomic.Value
}

var _ generic.Source = &validatingWebhookConfigurationManager{}

func NewValidatingWebhookConfigurationManager(f informers.SharedInformerFactory) generic.Source {
	informer := f.Admissionregistration().V1().ValidatingWebhookConfigurations()
	manager := &validatingWebhookConfigurationManager{
		configuration:              &atomic.Value{},
		lister:                     informer.Lister(),
		hasSynced:                  informer.Informer().HasSynced,
		initialConfigurationSynced: &atomic.Value{},
	}

	// Start with an empty list
	manager.configuration.Store([]webhook.WebhookAccessor{})
	manager.initialConfigurationSynced.Store(false)

	// On any change, rebuild the config
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(_ interface{}) { manager.updateConfiguration() },
		UpdateFunc: func(_, _ interface{}) { manager.updateConfiguration() },
		DeleteFunc: func(_ interface{}) { manager.updateConfiguration() },
	})

	return manager
}

func (v *validatingWebhookConfigurationManager) Webhooks() []webhook.WebhookAccessor {
	return v.configuration.Load().([]webhook.WebhookAccessor)
}

func (v *validatingWebhookConfigurationManager) HasSynced() bool {
	if !v.hasSynced() {
		return false
	}
	if v.initialConfigurationSynced.Load().(bool) {
		// the informer has synced and configuration has been updated
		return true
	}
	if configurations, err := v.lister.List(labels.Everything()); err == nil && len(configurations) == 0 {
		v.initialConfigurationSynced.Store(true)
		// the informer has synced and we don't have any items
		return true
	}
	return false

}

func (v *validatingWebhookConfigurationManager) updateConfiguration() {
	configurations, err := v.lister.List(labels.Everything())
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("error updating configuration: %v", err))
		return
	}
	v.configuration.Store(mergeValidatingWebhookConfigurations(configurations))
	v.initialConfigurationSynced.Store(true)
}

func mergeValidatingWebhookConfigurations(configurations []*v1.ValidatingWebhookConfiguration) []webhook.WebhookAccessor {
	sort.SliceStable(configurations, ValidatingWebhookConfigurationSorter(configurations).ByName)
	accessors := []webhook.WebhookAccessor{}
	for _, c := range configurations {
		// webhook names are not validated for uniqueness, so we check for duplicates and
		// add a int suffix to distinguish between them
		names := map[string]int{}
		for i := range c.Webhooks {
			n := c.Webhooks[i].Name
			uid := fmt.Sprintf("%s/%s/%d", c.Name, n, names[n])
			names[n]++
			accessors = append(accessors, webhook.NewValidatingWebhookAccessor(uid, c.Name, &c.Webhooks[i]))
		}
	}
	return accessors
}

type ValidatingWebhookConfigurationSorter []*v1.ValidatingWebhookConfiguration

func (a ValidatingWebhookConfigurationSorter) ByName(i, j int) bool {
	return a[i].Name < a[j].Name
}
