package options

import (
	"github.com/spf13/pflag"
	// @todo
	_ "k8s.io/apiserver/pkg/admission"
)

func AddCustomGlobalFlags(fs *pflag.FlagSet) {
	// @todo cloud
	// registerLegacyGlobalFlags(fs)

	// @todo cloud
	// globalflag.Register(fs, "default-not-ready-toleration-seconds")
	// globalflag.Register(fs, "default-unreachable-toleration-seconds")
}
