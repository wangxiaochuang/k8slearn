package options

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/admission"
)

var AllOrderedPlugins = []string{}

func RegisterAllAdmissionPlugins(plugins *admission.Plugins) {
}

func DefaultOffAdmissionPlugins() sets.String {
	defaultOnPlugins := sets.NewString()
	return sets.NewString(AllOrderedPlugins...).Difference(defaultOnPlugins)
}
