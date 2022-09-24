package options

import (
	"fmt"

	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/egressselector"
	"k8s.io/apiserver/pkg/tracing"
	"k8s.io/utils/path"
)

const apiserverService = "apiserver"

type TracingOptions struct {
	ConfigFile string
}

func NewTracingOptions() *TracingOptions {
	return &TracingOptions{}
}

func (o *TracingOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}

	fs.StringVar(&o.ConfigFile, "tracing-config-file", o.ConfigFile,
		"File with apiserver tracing configuration.")
}

func (o *TracingOptions) ApplyTo(es *egressselector.EgressSelector, c *server.Config) error {
	if o == nil || o.ConfigFile == "" {
		return nil
	}
	npConfig, err := tracing.ReadTracingConfiguration(o.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read tracing config: %v", err)
	}
	errs := tracing.ValidateTracingConfiguration(npConfig)
	if len(errs) > 0 {
		return fmt.Errorf("failed to validate tracing configuration: %v", errs.ToAggregate())
	}

	panic("####### tracing.go+40\n")
}

func (o *TracingOptions) Validate() (errs []error) {
	if o == nil || o.ConfigFile == "" {
		return
	}

	if exists, err := path.Exists(path.CheckFollowSymlink, o.ConfigFile); !exists {
		errs = append(errs, fmt.Errorf("tracing-config-file %s does not exist", o.ConfigFile))
	} else if err != nil {
		errs = append(errs, fmt.Errorf("error checking if tracing-config-file %s exists: %v", o.ConfigFile, err))
	}
	return
}
