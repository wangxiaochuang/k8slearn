package logs

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
)

const logFlushFreqFlagName = "log-flush-frequency"
const deprecated = "will be removed in a future release, see https://github.com/kubernetes/enhancements/tree/master/keps/sig-instrumentation/2845-deprecate-klog-specific-flags-in-k8s-components"

var (
	packageFlags = flag.NewFlagSet("logging", flag.ContinueOnError)

	logFlushFreq      time.Duration
	logFlushFreqAdded bool
)

func init() {
	// 这里会把默认的log参数都添加上
	klog.InitFlags(packageFlags)
	packageFlags.DurationVar(&logFlushFreq, logFlushFreqFlagName, 5*time.Second, "Maximum number of seconds between log flushes")
}

type addFlagsOptions struct {
	skipLoggingConfigurationFlags bool
}

type Option func(*addFlagsOptions)

func SkipLoggingConfigurationFlags() Option {
	return func(o *addFlagsOptions) {
		o.skipLoggingConfigurationFlags = true
	}
}

func AddFlags(fs *pflag.FlagSet, opts ...Option) {
	if fs.Lookup("logtostderr") != nil {
		return
	}
	o := addFlagsOptions{}
	for _, opt := range opts {
		opt(&o)
	}

	// 把除了指定的flag外的其他都设置为废弃状态
	packageFlags.VisitAll(func(f *flag.Flag) {
		pf := pflag.PFlagFromGoFlag(f)
		switch f.Name {
		case "v":
			if o.skipLoggingConfigurationFlags {
				return
			}
		case logFlushFreqFlagName:
			if o.skipLoggingConfigurationFlags {
				return
			}
			logFlushFreqAdded = true
		case "vmodule":
			if o.skipLoggingConfigurationFlags {
				return
			}
		default:
			pf.Deprecated = deprecated
		}
		fs.AddFlag(pf)
	})
}

func AddGoFlags(fs *flag.FlagSet, opts ...Option) {
	o := addFlagsOptions{}
	for _, opt := range opts {
		opt(&o)
	}

	packageFlags.VisitAll(func(f *flag.Flag) {
		usage := f.Usage
		switch f.Name {
		case "v":
			// unchanged
			if o.skipLoggingConfigurationFlags {
				return
			}
		case logFlushFreqFlagName:
			// unchanged
			if o.skipLoggingConfigurationFlags {
				return
			}
			logFlushFreqAdded = true
		case "vmodule":
			// TODO: see above
			// usage += vmoduleUsage
			if o.skipLoggingConfigurationFlags {
				return
			}
		default:
			usage += " (DEPRECATED: " + deprecated + ")"
		}
		fs.Var(f.Value, f.Name, usage)
	})
}

type KlogWriter struct{}

func (writer KlogWriter) Write(data []byte) (n int, err error) {
	klog.InfoDepth(1, string(data))
	return len(data), nil
}

func InitLogs() {
	log.SetOutput(KlogWriter{})
	log.SetFlags(0)
	if logFlushFreqAdded {
		// The flag from this file was activated, so use it now.
		// Otherwise LoggingConfiguration.Apply will do this.
		klog.StartFlushDaemon(logFlushFreq)
	}

	// This is the default in Kubernetes. Options.ValidateAndApply
	// will override this with the result of a feature gate check.
	klog.EnableContextualLogging(false)
}

func FlushLogs() {
	klog.Flush()
}

func NewLogger(prefix string) *log.Logger {
	return log.New(KlogWriter{}, prefix, 0)
}

func GlogSetter(val string) (string, error) {
	var level klog.Level
	if err := level.Set(val); err != nil {
		return "", fmt.Errorf("failed set klog.logging.verbosity %s: %v", val, err)
	}
	return fmt.Sprintf("successfully set klog.logging.verbosity to %s", val), nil
}
