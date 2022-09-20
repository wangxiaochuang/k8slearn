package globalflag

import (
	"flag"
	"fmt"

	"github.com/spf13/pflag"
	"k8s.io/component-base/logs"
)

func AddGlobalFlags(fs *pflag.FlagSet, name string, opts ...logs.Option) {
	// klog已经设置了不少flag了，将其设置到全局里
	logs.AddFlags(fs, opts...)

	fs.BoolP("help", "h", false, fmt.Sprintf("help for %s", name))
}

func Register(local *pflag.FlagSet, globalName string) {
	if f := flag.CommandLine.Lookup(globalName); f != nil {
		pflagFlag := pflag.PFlagFromGoFlag(f)
		normalizeFunc := local.GetNormalizeFunc()
		pflagFlag.Name = string(normalizeFunc(local, pflagFlag.Name))
		local.AddFlag(pflagFlag)
	} else {
		panic(fmt.Sprintf("failed to find flag in global flagset (flag): %s", globalName))
	}
}
