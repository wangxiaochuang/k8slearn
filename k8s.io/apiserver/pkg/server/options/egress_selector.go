package options

import (
	"fmt"

	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/egressselector"
	"k8s.io/utils/path"
)

type EgressSelectorOptions struct {
	ConfigFile string
}

func NewEgressSelectorOptions() *EgressSelectorOptions {
	return &EgressSelectorOptions{}
}

func (o *EgressSelectorOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}

	fs.StringVar(&o.ConfigFile, "egress-selector-config-file", o.ConfigFile,
		"File with apiserver egress selector configuration.")
}

func (o *EgressSelectorOptions) ApplyTo(c *server.Config) error {
	if o == nil {
		return nil
	}

	npConfig, err := egressselector.ReadEgressSelectorConfiguration(o.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read egress selector config: %v", err)
	}
	errs := egressselector.ValidateEgressSelectorConfiguration(npConfig)
	if len(errs) > 0 {
		return fmt.Errorf("failed to validate egress selector configuration: %v", errs.ToAggregate())
	}

	cs, err := egressselector.NewEgressSelector(npConfig)
	if err != nil {
		return fmt.Errorf("failed to setup egress selector with config %#v: %v", npConfig, err)
	}
	c.EgressSelector = cs
	return nil
}

func (o *EgressSelectorOptions) Validate() []error {
	if o == nil || o.ConfigFile == "" {
		return nil
	}

	errs := []error{}

	if exists, err := path.Exists(path.CheckFollowSymlink, o.ConfigFile); !exists || err != nil {
		errs = append(errs, fmt.Errorf("egress-selector-config-file %s does not exist", o.ConfigFile))
	}

	return errs
}
