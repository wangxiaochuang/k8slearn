package alwayspullimages

import (
	"context"
	"io"

	"k8s.io/apiserver/pkg/admission"
)

const PluginName = "AlwaysPullImages"

func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return NewAlwaysPullImages(), nil
	})
}

type AlwaysPullImages struct {
	*admission.Handler
}

var _ admission.MutationInterface = &AlwaysPullImages{}
var _ admission.ValidationInterface = &AlwaysPullImages{}

func (a *AlwaysPullImages) Admit(ctx context.Context, attributes admission.Attributes, o admission.ObjectInterfaces) (err error) {
	panic("not implemented")
}

func (*AlwaysPullImages) Validate(ctx context.Context, attributes admission.Attributes, o admission.ObjectInterfaces) (err error) {
	panic("not implemented")
}

func NewAlwaysPullImages() *AlwaysPullImages {
	return &AlwaysPullImages{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}
}
