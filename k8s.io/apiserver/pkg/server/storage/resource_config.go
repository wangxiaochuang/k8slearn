package storage

import "k8s.io/apimachinery/pkg/runtime/schema"

type ResourceConfig struct {
	GroupVersionConfigs map[schema.GroupVersion]bool
	ResourceConfigs     map[schema.GroupVersionResource]bool
}

func NewResourceConfig() *ResourceConfig {
	return &ResourceConfig{GroupVersionConfigs: map[schema.GroupVersion]bool{}, ResourceConfigs: map[schema.GroupVersionResource]bool{}}
}

// p65
func resourceMatcherForVersion(gv schema.GroupVersion) func(gvr schema.GroupVersionResource) bool {
	return func(gvr schema.GroupVersionResource) bool {
		return gv == gvr.GroupVersion()
	}
}

func (o *ResourceConfig) removeMatchingResourcePreferences(matcher func(gvr schema.GroupVersionResource) bool) {
	keysToRemove := []schema.GroupVersionResource{}
	for k := range o.ResourceConfigs {
		if matcher(k) {
			keysToRemove = append(keysToRemove, k)
		}
	}
	for _, k := range keysToRemove {
		delete(o.ResourceConfigs, k)
	}
}

// p85
func (o *ResourceConfig) DisableVersions(versions ...schema.GroupVersion) {
	for _, version := range versions {
		o.GroupVersionConfigs[version] = false

		o.removeMatchingResourcePreferences(resourceMatcherForVersion(version))
	}
}

func (o *ResourceConfig) EnableVersions(versions ...schema.GroupVersion) {
	for _, version := range versions {
		o.GroupVersionConfigs[version] = true

		// a preference about a version takes priority over the previously set resources
		o.removeMatchingResourcePreferences(resourceMatcherForVersion(version))
	}

}

func (o *ResourceConfig) DisableResources(resources ...schema.GroupVersionResource) {
	for _, resource := range resources {
		o.ResourceConfigs[resource] = false
	}
}

// p118
func (o *ResourceConfig) EnableResources(resources ...schema.GroupVersionResource) {
	for _, resource := range resources {
		o.ResourceConfigs[resource] = true
	}
}
