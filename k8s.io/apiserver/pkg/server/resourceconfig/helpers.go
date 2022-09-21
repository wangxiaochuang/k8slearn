package resourceconfig

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	serverstore "k8s.io/apiserver/pkg/server/storage"
	cliflag "k8s.io/component-base/cli/flag"
)

type GroupVersionRegistry interface {
	// IsGroupRegistered returns true if given group is registered.
	IsGroupRegistered(group string) bool
	// IsVersionRegistered returns true if given version is registered.
	IsVersionRegistered(v schema.GroupVersion) bool
	// PrioritizedVersionsAllGroups returns all registered group versions.
	PrioritizedVersionsAllGroups() []schema.GroupVersion
}

func MergeResourceEncodingConfigs(
	defaultResourceEncoding *serverstore.DefaultResourceEncodingConfig,
	resourceEncodingOverrides []schema.GroupVersionResource,
) *serverstore.DefaultResourceEncodingConfig {
	resourceEncodingConfig := defaultResourceEncoding
	for _, gvr := range resourceEncodingOverrides {
		resourceEncodingConfig.SetResourceEncoding(gvr.GroupResource(), gvr.GroupVersion(),
			schema.GroupVersion{Group: gvr.Group, Version: runtime.APIVersionInternal})
	}
	return resourceEncodingConfig
}

const (
	APIAll   = "api/all"
	APIGA    = "api/ga"
	APIBeta  = "api/beta"
	APIAlpha = "api/alpha"
)

var (
	gaPattern    = regexp.MustCompile(`^v\d+$`)
	betaPattern  = regexp.MustCompile(`^v\d+beta\d+$`)
	alphaPattern = regexp.MustCompile(`^v\d+alpha\d+$`)

	groupVersionMatchers = map[string]func(gv schema.GroupVersion) bool{
		// allows users to address all api versions
		APIAll: func(gv schema.GroupVersion) bool { return true },
		// allows users to address all api versions in the form v[0-9]+
		APIGA: func(gv schema.GroupVersion) bool { return gaPattern.MatchString(gv.Version) },
		// allows users to address all beta api versions
		APIBeta: func(gv schema.GroupVersion) bool { return betaPattern.MatchString(gv.Version) },
		// allows users to address all alpha api versions
		APIAlpha: func(gv schema.GroupVersion) bool { return alphaPattern.MatchString(gv.Version) },
	}

	groupVersionMatchersOrder = []string{APIAll, APIGA, APIBeta, APIAlpha}
)

func MergeAPIResourceConfigs(
	defaultAPIResourceConfig *serverstore.ResourceConfig,
	resourceConfigOverrides cliflag.ConfigurationMap,
	registry GroupVersionRegistry,
) (*serverstore.ResourceConfig, error) {
	resourceConfig := defaultAPIResourceConfig
	overrides := resourceConfigOverrides

	for _, flag := range groupVersionMatchersOrder {
		if value, ok := overrides[flag]; ok {
			if value == "false" {
				resourceConfig.DisableMatchingVersions(groupVersionMatchers[flag])
			} else if value == "true" {
				resourceConfig.EnableMatchingVersions(groupVersionMatchers[flag])
			} else {
				return nil, fmt.Errorf("invalid value %v=%v", flag, value)
			}
		}
	}

	type versionEnablementPreference struct {
		key          string
		enabled      bool
		groupVersion schema.GroupVersion
	}
	type resourceEnablementPreference struct {
		key                  string
		enabled              bool
		groupVersionResource schema.GroupVersionResource
	}
	versionPreferences := []versionEnablementPreference{}
	resourcePreferences := []resourceEnablementPreference{}

	for key := range overrides {
		if _, ok := groupVersionMatchers[key]; ok {
			continue
		}

		tokens := strings.Split(key, "/")
		if len(tokens) < 2 || len(tokens) > 3 {
			continue
		}
		groupVersionString := tokens[0] + "/" + tokens[1]
		groupVersion, err := schema.ParseGroupVersion(groupVersionString)
		if err != nil {
			return nil, fmt.Errorf("invalid key %s", key)
		}

		if !registry.IsGroupRegistered(groupVersion.Group) {
			continue
		}

		if !registry.IsVersionRegistered(groupVersion) {
			return nil, fmt.Errorf("group version %s that has not been registered", groupVersion.String())
		}
		enabled, err := getRuntimeConfigValue(overrides, key, false)
		if err != nil {
			return nil, err
		}

		switch len(tokens) {
		case 2:
			versionPreferences = append(versionPreferences, versionEnablementPreference{
				key:          key,
				enabled:      enabled,
				groupVersion: groupVersion,
			})
		case 3:
			if strings.ToLower(tokens[2]) != tokens[2] {
				return nil, fmt.Errorf("invalid key %v: group/version/resource and resource is always lowercase plural, not %q", key, tokens[2])
			}
			resourcePreferences = append(resourcePreferences, resourceEnablementPreference{
				key:                  key,
				enabled:              enabled,
				groupVersionResource: groupVersion.WithResource(tokens[2]),
			})
		}
	}

	for _, versionPreference := range versionPreferences {
		if versionPreference.enabled {
			// enable the groupVersion for "group/version=true"
			resourceConfig.EnableVersions(versionPreference.groupVersion)

		} else {
			// disable the groupVersion only for "group/version=false"
			resourceConfig.DisableVersions(versionPreference.groupVersion)
		}
	}

	for _, resourcePreference := range resourcePreferences {
		if resourcePreference.enabled {
			// enable the resource for "group/version/resource=true"
			resourceConfig.EnableResources(resourcePreference.groupVersionResource)
		} else {
			resourceConfig.DisableResources(resourcePreference.groupVersionResource)
		}
	}

	return resourceConfig, nil
}

func getRuntimeConfigValue(overrides cliflag.ConfigurationMap, apiKey string, defaultValue bool) (bool, error) {
	flagValue, ok := overrides[apiKey]
	if ok {
		if flagValue == "" {
			return true, nil
		}
		boolValue, err := strconv.ParseBool(flagValue)
		if err != nil {
			return false, fmt.Errorf("invalid value of %s: %s, err: %v", apiKey, flagValue, err)
		}
		return boolValue, nil
	}
	return defaultValue, nil
}

func ParseGroups(resourceConfig cliflag.ConfigurationMap) ([]string, error) {
	groups := []string{}
	for key := range resourceConfig {
		if _, ok := groupVersionMatchers[key]; ok {
			continue
		}
		tokens := strings.Split(key, "/")
		if len(tokens) != 2 && len(tokens) != 3 {
			return groups, fmt.Errorf("runtime-config invalid key %s", key)
		}
		groupVersionString := tokens[0] + "/" + tokens[1]
		groupVersion, err := schema.ParseGroupVersion(groupVersionString)
		if err != nil {
			return nil, fmt.Errorf("runtime-config invalid key %s", key)
		}
		groups = append(groups, groupVersion.Group)
	}

	return groups, nil
}
