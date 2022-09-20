package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClientConnectionConfiguration struct {
	Kubeconfig         string
	AcceptContentTypes string
	ContentType        string
	QPS                float32
	Burst              int32
}

type LeaderElectionConfiguration struct {
	LeaderElect       bool
	LeaseDuration     metav1.Duration
	RenewDeadline     metav1.Duration
	RetryPeriod       metav1.Duration
	ResourceLock      string
	ResourceName      string
	ResourceNamespace string
}

type DebuggingConfiguration struct {
	EnableProfiling           bool
	EnableContentionProfiling bool
}

type LoggingConfiguration struct {
	Format         string
	FlushFrequency time.Duration
	Verbosity      VerbosityLevel
	VModule        VModuleConfiguration
	Options        FormatOptions
}

type FormatOptions struct {
	JSON JSONOptions
}

type JSONOptions struct {
	SplitStream    bool
	InfoBufferSize resource.QuantityValue
}

type VModuleConfiguration []VModuleItem

var _ pflag.Value = &VModuleConfiguration{}

type VModuleItem struct {
	FilePattern string
	Verbosity   VerbosityLevel
}

func (vmodule *VModuleConfiguration) String() string {
	var patterns []string
	for _, item := range *vmodule {
		patterns = append(patterns, fmt.Sprintf("%s=%d", item.FilePattern, item.Verbosity))
	}
	return strings.Join(patterns, ",")
}

func (vmodule *VModuleConfiguration) Set(value string) error {
	for _, pat := range strings.Split(value, ",") {
		if len(pat) == 0 {
			continue
		}
		patLev := strings.Split(pat, "=")
		if len(patLev) != 2 || len(patLev[0]) == 0 || len(patLev[1]) == 0 {
			return fmt.Errorf("%q does not have the pattern=N format", pat)
		}
		pattern := patLev[0]
		v, err := strconv.ParseUint(patLev[1], 10, 31)
		if err != nil {
			return fmt.Errorf("parsing verbosity in %q: %v", pat, err)
		}
		*vmodule = append(*vmodule, VModuleItem{FilePattern: pattern, Verbosity: VerbosityLevel(v)})
	}
	return nil
}

func (vmodule *VModuleConfiguration) Type() string {
	return "pattern=N,..."
}

type VerbosityLevel uint32

var _ pflag.Value = new(VerbosityLevel)

func (l *VerbosityLevel) String() string {
	return strconv.FormatInt(int64(*l), 10)
}

func (l *VerbosityLevel) Get() interface{} {
	return *l
}

func (l *VerbosityLevel) Set(value string) error {
	v, err := strconv.ParseUint(value, 10, 31)
	if err != nil {
		return err
	}
	*l = VerbosityLevel(v)
	return nil
}

func (l *VerbosityLevel) Type() string {
	return "Level"
}
