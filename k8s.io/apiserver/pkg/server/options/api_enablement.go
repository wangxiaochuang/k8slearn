package options

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/server/resourceconfig"
	cliflag "k8s.io/component-base/cli/flag"
)

type APIEnablementOptions struct {
	RuntimeConfig cliflag.ConfigurationMap
}

func NewAPIEnablementOptions() *APIEnablementOptions {
	return &APIEnablementOptions{
		RuntimeConfig: make(cliflag.ConfigurationMap),
	}
}

func (s *APIEnablementOptions) AddFlags(fs *pflag.FlagSet) {
	fs.Var(&s.RuntimeConfig, "runtime-config", ""+
		"A set of key=value pairs that enable or disable built-in APIs. Supported options are:\n"+
		"v1=true|false for the core API group\n"+
		"<group>/<version>=true|false for a specific API group and version (e.g. apps/v1=true)\n"+
		"api/all=true|false controls all API versions\n"+
		"api/ga=true|false controls all API versions of the form v[0-9]+\n"+
		"api/beta=true|false controls all API versions of the form v[0-9]+beta[0-9]+\n"+
		"api/alpha=true|false controls all API versions of the form v[0-9]+alpha[0-9]+\n"+
		"api/legacy is deprecated, and will be removed in a future version")
}

func (s *APIEnablementOptions) Validate(registries ...GroupRegistry) []error {
	if s == nil {
		return nil
	}

	errors := []error{}
	if s.RuntimeConfig[resourceconfig.APIAll] == "false" && len(s.RuntimeConfig) == 1 {
		// Do not allow only set api/all=false, in such case apiserver startup has no meaning.
		return append(errors, fmt.Errorf("invalid key with only %v=false", resourceconfig.APIAll))
	}

	groups, err := resourceconfig.ParseGroups(s.RuntimeConfig)
	if err != nil {
		return append(errors, err)
	}

	for _, registry := range registries {
		// filter out known groups
		groups = unknownGroups(groups, registry)
	}
	if len(groups) != 0 {
		errors = append(errors, fmt.Errorf("unknown api groups %s", strings.Join(groups, ",")))
	}

	return errors
}

func unknownGroups(groups []string, registry GroupRegistry) []string {
	unknownGroups := []string{}
	for _, group := range groups {
		if !registry.IsGroupRegistered(group) {
			unknownGroups = append(unknownGroups, group)
		}
	}
	return unknownGroups
}

type GroupRegistry interface {
	IsGroupRegistered(group string) bool
}
