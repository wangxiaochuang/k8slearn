package options

import (
	"github.com/spf13/pflag"
	"k8s.io/component-base/cli/globalflag"
)

func registerLegacyGlobalFlags(fs *pflag.FlagSet) {
	globalflag.Register(fs, "cloud-provider-gce-lb-src-cidrs")
	globalflag.Register(fs, "cloud-provider-gce-l7lb-src-cidrs")
	fs.MarkDeprecated("cloud-provider-gce-lb-src-cidrs", "This flag will be removed once the GCE Cloud Provider is removed from kube-apiserver")
}
