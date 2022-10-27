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

type mutatingWebhookConfigurationManager struct {
	configuration              *atomic.Value
	lister                     admissionregistrationlisters.MutatingWebhookConfigurationLister
	hasSynced                  func() bool
	initialConfigurationSynced *atomic.Value
}

var _ generic.Source = &mutatingWebhookConfigurationManager{}

func NewMutatingWebhookConfigurationManager(f informers.SharedInformerFactory) generic.Source {
	informer := f.Admissionregistration().V1().MutatingWebhookConfigurations()
	manager := &mutatingWebhookConfigurationManager{
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

func (m *mutatingWebhookConfigurationManager) Webhooks() []webhook.WebhookAccessor {
	return m.configuration.Load().([]webhook.WebhookAccessor)
}

func (m *mutatingWebhookConfigurationManager) HasSynced() bool {
	if !m.hasSynced() {
		return false
	}
	if m.initialConfigurationSynced.Load().(bool) {
		return true
	}
	if configurations, err := m.lister.List(labels.Everything()); err == nil && len(configurations) == 0 {
		m.initialConfigurationSynced.Store(true)
		return true
	}
	return false
}

func (m *mutatingWebhookConfigurationManager) updateConfiguration() {
	configurations, err := m.lister.List(labels.Everything())
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("error updating configuration: %v", err))
		return
	}
	m.configuration.Store(mergeMutatingWebhookConfigurations(configurations))
	m.initialConfigurationSynced.Store(true)
}

func mergeMutatingWebhookConfigurations(configurations []*v1.MutatingWebhookConfiguration) []webhook.WebhookAccessor {
	// The internal order of webhooks for each configuration is provided by the user
	// but configurations themselves can be in any order. As we are going to run these
	// webhooks in serial, they are sorted here to have a deterministic order.
	sort.SliceStable(configurations, MutatingWebhookConfigurationSorter(configurations).ByName)
	accessors := []webhook.WebhookAccessor{}
	for _, c := range configurations {
		// webhook names are not validated for uniqueness, so we check for duplicates and
		// add a int suffix to distinguish between them
		names := map[string]int{}
		for i := range c.Webhooks {
			n := c.Webhooks[i].Name
			uid := fmt.Sprintf("%s/%s/%d", c.Name, n, names[n])
			names[n]++
			accessors = append(accessors, webhook.NewMutatingWebhookAccessor(uid, c.Name, &c.Webhooks[i]))
		}
	}
	return accessors
}

type MutatingWebhookConfigurationSorter []*v1.MutatingWebhookConfiguration

func (a MutatingWebhookConfigurationSorter) ByName(i, j int) bool {
	return a[i].Name < a[j].Name
}
