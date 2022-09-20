package resourceconfig

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
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
