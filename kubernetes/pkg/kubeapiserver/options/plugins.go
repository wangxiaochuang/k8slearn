package options

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/admission/plugin/namespace/lifecycle"
	mutatingwebhook "k8s.io/apiserver/pkg/admission/plugin/webhook/mutating"
	validatingwebhook "k8s.io/apiserver/pkg/admission/plugin/webhook/validating"
	"k8s.io/kubernetes/plugin/pkg/admission/alwayspullimages"
)

var AllOrderedPlugins = []string{
	lifecycle.PluginName,         // NamespaceLifecycle
	alwayspullimages.PluginName,  // AlwaysPullImages
	mutatingwebhook.PluginName,   // MutatingAdmissionWebhook
	validatingwebhook.PluginName, // ValidatingAdmissionWebhook
}

func RegisterAllAdmissionPlugins(plugins *admission.Plugins) {
	alwayspullimages.Register(plugins)
}

func DefaultOffAdmissionPlugins() sets.String {
	defaultOnPlugins := sets.NewString()
	return sets.NewString(AllOrderedPlugins...).Difference(defaultOnPlugins)
}
