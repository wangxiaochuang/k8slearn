package mutating

import (
	"context"
	"io"

	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/admission/configuration"
	"k8s.io/apiserver/pkg/admission/plugin/webhook/generic"
)

const (
	// PluginName indicates the name of admission plug-in
	PluginName = "MutatingAdmissionWebhook"
)

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(configFile io.Reader) (admission.Interface, error) {
		plugin, err := NewMutatingWebhook(configFile)
		if err != nil {
			return nil, err
		}

		return plugin, nil
	})
}

// Plugin is an implementation of admission.Interface.
type Plugin struct {
	*generic.Webhook
}

var _ admission.MutationInterface = &Plugin{}

// NewMutatingWebhook returns a generic admission webhook plugin.
func NewMutatingWebhook(configFile io.Reader) (*Plugin, error) {
	handler := admission.NewHandler(admission.Connect, admission.Create, admission.Delete, admission.Update)
	p := &Plugin{}
	var err error
	p.Webhook, err = generic.NewWebhook(handler, configFile, configuration.NewMutatingWebhookConfigurationManager, newMutatingDispatcher(p))
	if err != nil {
		return nil, err
	}

	return p, nil
}

// ValidateInitialization implements the InitializationValidator interface.
func (a *Plugin) ValidateInitialization() error {
	if err := a.Webhook.ValidateInitialization(); err != nil {
		return err
	}
	return nil
}

// Admit makes an admission decision based on the request attributes.
func (a *Plugin) Admit(ctx context.Context, attr admission.Attributes, o admission.ObjectInterfaces) error {
	return a.Webhook.Dispatch(ctx, attr, o)
}
