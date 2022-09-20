package logs

import (
	"fmt"

	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/component-base/config"
	"k8s.io/component-base/config/v1alpha1"
	"k8s.io/component-base/featuregate"
	"k8s.io/component-base/logs/registry"
	"k8s.io/klog/v2"
)

type Options struct {
	Config config.LoggingConfiguration
}

// 使用的是v1alpha版本
func NewOptions() *Options {
	c := v1alpha1.LoggingConfiguration{}
	v1alpha1.RecommendedLoggingConfiguration(&c)
	o := &Options{}
	v1alpha1.Convert_v1alpha1_LoggingConfiguration_To_config_LoggingConfiguration(&c, &o.Config, nil)
	return o
}

func (o *Options) ValidateAndApply(featureGate featuregate.FeatureGate) error {
	errs := o.validate()
	if len(errs) > 0 {
		return utilerrors.NewAggregate(errs)
	}
	o.apply(featureGate)
	return nil
}

func (o *Options) validate() []error {
	errs := ValidateLoggingConfiguration(&o.Config, nil)
	if len(errs) != 0 {
		return errs.ToAggregate().Errors()
	}
	return nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	BindLoggingFlags(&o.Config, fs)
}

func (o *Options) apply(featureGate featuregate.FeatureGate) {
	contextualLoggingEnabled := contextualLoggingDefault
	if featureGate != nil {
		contextualLoggingEnabled = featureGate.Enabled(ContextualLogging)
	}

	factory, _ := registry.LogRegistry.Get(o.Config.Format)
	if factory == nil {
		klog.ClearLogger()
	} else {
		log, flush := factory.Create(o.Config)
		klog.SetLoggerWithOptions(log, klog.ContextualLogger(contextualLoggingEnabled), klog.FlushLogger(flush))
	}
	if err := loggingFlags.Lookup("v").Value.Set(o.Config.Verbosity.String()); err != nil {
		panic(fmt.Errorf("internal error while setting klog verbosity: %v", err))
	}
	if err := loggingFlags.Lookup("vmodule").Value.Set(o.Config.VModule.String()); err != nil {
		panic(fmt.Errorf("internal error while setting klog vmodule: %v", err))
	}
	klog.StartFlushDaemon(o.Config.FlushFrequency)
	klog.EnableContextualLogging(contextualLoggingEnabled)
}
