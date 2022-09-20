package logs

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/pflag"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/config"
	"k8s.io/component-base/logs/registry"
	"k8s.io/klog/v2"
)

const (
	DefaultLogFormat = "text"
	JSONLogFormat    = "json"
)

var loggingFlags pflag.FlagSet

func init() {
	// Text format is default klog format
	registry.LogRegistry.Register(DefaultLogFormat, nil)

	var fs flag.FlagSet
	klog.InitFlags(&fs)
	// 将klog里的flag放这里来
	loggingFlags.AddGoFlagSet(&fs)
}

var supportedLogsFlags = map[string]struct{}{
	"v": {},
}

func BindLoggingFlags(c *config.LoggingConfiguration, fs *pflag.FlagSet) {
	unsupportedFlags := strings.Join(unsupportedLoggingFlagNames(cliflag.WordSepNormalizeFunc), ", ")
	formats := fmt.Sprintf(`"%s"`, strings.Join(registry.LogRegistry.List(), `", "`))
	fs.StringVar(&c.Format, "logging-format", c.Format, fmt.Sprintf("Sets the log format. Permitted formats: %s.\nNon-default formats don't honor these flags: %s.\nNon-default choices are currently alpha and subject to change without warning.", formats, unsupportedFlags))
	// No new log formats should be added after generation is of flag options
	registry.LogRegistry.Freeze()

	fs.DurationVar(&c.FlushFrequency, logFlushFreqFlagName, logFlushFreq, "Maximum number of seconds between log flushes")
	fs.VarP(&c.Verbosity, "v", "v", "number for the log level verbosity")
	fs.Var(&c.VModule, "vmodule", "comma-separated list of pattern=N settings for file-filtered logging (only works for text log format)")

	// JSON options. We only register them if "json" is a valid format. The
	// config file API however always has them.
	if _, err := registry.LogRegistry.Get("json"); err == nil {
		fs.BoolVar(&c.Options.JSON.SplitStream, "log-json-split-stream", false, "[Experimental] In JSON format, write error messages to stderr and info messages to stdout. The default is to write a single stream to stdout.")
		fs.Var(&c.Options.JSON.InfoBufferSize, "log-json-info-buffer-size", "[Experimental] In JSON format with split output streams, the info messages can be buffered for a while to increase performance. The default value of zero bytes disables buffering. The size can be specified as number of bytes (512), multiples of 1000 (1K), multiples of 1024 (2Ki), or powers of those (3M, 4G, 5Mi, 6Gi).")
	}
}

func UnsupportedLoggingFlags(normalizeFunc func(f *pflag.FlagSet, name string) pflag.NormalizedName) []*pflag.Flag {
	// k8s.io/component-base/logs and klog flags
	pfs := &pflag.FlagSet{}
	loggingFlags.VisitAll(func(flag *pflag.Flag) {
		if _, found := supportedLogsFlags[flag.Name]; !found {
			// Normalization changes flag.Name, so make a copy.
			clone := *flag
			pfs.AddFlag(&clone)
		}
	})

	// Apply normalization.
	pfs.SetNormalizeFunc(normalizeFunc)

	var allFlags []*pflag.Flag
	pfs.VisitAll(func(flag *pflag.Flag) {
		allFlags = append(allFlags, flag)
	})
	return allFlags
}

func unsupportedLoggingFlagNames(normalizeFunc func(f *pflag.FlagSet, name string) pflag.NormalizedName) []string {
	unsupportedFlags := UnsupportedLoggingFlags(normalizeFunc)
	names := make([]string, 0, len(unsupportedFlags))
	for _, f := range unsupportedFlags {
		names = append(names, "--"+f.Name)
	}
	sort.Strings(names)
	return names
}
