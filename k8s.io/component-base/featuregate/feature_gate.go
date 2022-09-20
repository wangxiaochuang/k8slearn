package featuregate

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/naming"
	"k8s.io/klog/v2"
)

type Feature string

const (
	flagName             = "feature-gates"
	allAlphaGate Feature = "AllAlpha"
	allBetaGate  Feature = "AllBeta"
)

var (
	// The generic features.
	defaultFeatures = map[Feature]FeatureSpec{
		allAlphaGate: {Default: false, PreRelease: Alpha},
		allBetaGate:  {Default: false, PreRelease: Beta},
	}

	// Special handling for a few gates.
	specialFeatures = map[Feature]func(known map[Feature]FeatureSpec, enabled map[Feature]bool, val bool){
		allAlphaGate: setUnsetAlphaGates,
		allBetaGate:  setUnsetBetaGates,
	}
)

type FeatureSpec struct {
	Default       bool
	LockToDefault bool
	PreRelease    prerelease
}

type prerelease string

const (
	// Values for PreRelease.
	Alpha = prerelease("ALPHA")
	Beta  = prerelease("BETA")
	GA    = prerelease("")

	// Deprecated
	Deprecated = prerelease("DEPRECATED")
)

type FeatureGate interface {
	Enabled(key Feature) bool
	KnownFeatures() []string
	DeepCopy() MutableFeatureGate
}

type MutableFeatureGate interface {
	FeatureGate

	AddFlag(fs *pflag.FlagSet)
	Set(value string) error
	SetFromMap(m map[string]bool) error
	Add(features map[Feature]FeatureSpec) error
	GetAll() map[Feature]FeatureSpec
}

type featureGate struct {
	featureGateName string

	special map[Feature]func(map[Feature]FeatureSpec, map[Feature]bool, bool)

	// lock guards writes to known, enabled, and reads/writes of closed
	lock sync.Mutex
	// known holds a map[Feature]FeatureSpec
	known *atomic.Value
	// enabled holds a map[Feature]bool
	enabled *atomic.Value
	// closed is set to true when AddFlag is called, and prevents subsequent calls to Add
	closed bool
}

func setUnsetAlphaGates(known map[Feature]FeatureSpec, enabled map[Feature]bool, val bool) {
	for k, v := range known {
		if v.PreRelease == Alpha {
			if _, found := enabled[k]; !found {
				enabled[k] = val
			}
		}
	}
}

func setUnsetBetaGates(known map[Feature]FeatureSpec, enabled map[Feature]bool, val bool) {
	for k, v := range known {
		if v.PreRelease == Beta {
			if _, found := enabled[k]; !found {
				enabled[k] = val
			}
		}
	}
}

var _ pflag.Value = &featureGate{}

var internalPackages = []string{"k8s.io/component-base/featuregate/feature_gate.go"}

func NewFeatureGate() *featureGate {
	known := map[Feature]FeatureSpec{}
	for k, v := range defaultFeatures {
		known[k] = v
	}

	knownValue := &atomic.Value{}
	knownValue.Store(known)

	enabled := map[Feature]bool{}
	enabledValue := &atomic.Value{}
	enabledValue.Store(enabled)

	f := &featureGate{
		featureGateName: naming.GetNameFromCallsite(internalPackages...),
		known:           knownValue,
		special:         specialFeatures,
		enabled:         enabledValue,
	}
	return f
}

// Set parses a string of the form "key1=value1,key2=value2,..." into a
// map[string]bool of known keys or returns an error.
func (f *featureGate) Set(value string) error {
	m := make(map[string]bool)
	for _, s := range strings.Split(value, ",") {
		if len(s) == 0 {
			continue
		}
		arr := strings.SplitN(s, "=", 2)
		k := strings.TrimSpace(arr[0])
		if len(arr) != 2 {
			return fmt.Errorf("missing bool value for %s", k)
		}
		v := strings.TrimSpace(arr[1])
		boolValue, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("invalid value of %s=%s, err: %v", k, v, err)
		}
		m[k] = boolValue
	}
	return f.SetFromMap(m)
}

func (f *featureGate) SetFromMap(m map[string]bool) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	// Copy existing state
	known := map[Feature]FeatureSpec{}
	for k, v := range f.known.Load().(map[Feature]FeatureSpec) {
		known[k] = v
	}
	enabled := map[Feature]bool{}
	for k, v := range f.enabled.Load().(map[Feature]bool) {
		enabled[k] = v
	}

	for k, v := range m {
		k := Feature(k)
		featureSpec, ok := known[k]
		if !ok {
			return fmt.Errorf("unrecognized feature gate: %s", k)
		}
		if featureSpec.LockToDefault && featureSpec.Default != v {
			return fmt.Errorf("cannot set feature gate %v to %v, feature is locked to %v", k, v, featureSpec.Default)
		}
		enabled[k] = v

		// special的要执行函数
		if fn, found := f.special[k]; found {
			fn(known, enabled, v)
		}

		if featureSpec.PreRelease == Deprecated {
			klog.Warningf("Setting deprecated feature gate %s=%t. It will be removed in a future release.", k, v)
		} else if featureSpec.PreRelease == GA {
			klog.Warningf("Setting GA feature gate %s=%t. It will be removed in a future release.", k, v)
		}
	}

	// Persist changes
	f.known.Store(known)
	f.enabled.Store(enabled)

	klog.V(1).Infof("feature gates: %v", f.enabled)
	return nil
}

func (f *featureGate) String() string {
	pairs := []string{}
	for k, v := range f.enabled.Load().(map[Feature]bool) {
		pairs = append(pairs, fmt.Sprintf("%s=%t", k, v))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ",")
}

func (f *featureGate) Type() string {
	return "mapStringBool"
}

func (f *featureGate) Add(features map[Feature]FeatureSpec) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	if f.closed {
		return fmt.Errorf("cannot add a feature gate after adding it to the flag set")
	}

	// Copy existing state
	known := map[Feature]FeatureSpec{}
	for k, v := range f.known.Load().(map[Feature]FeatureSpec) {
		known[k] = v
	}

	for name, spec := range features {
		if existingSpec, found := known[name]; found {
			if existingSpec == spec {
				continue
			}
			return fmt.Errorf("feature gate %q with different spec already exists: %v", name, existingSpec)
		}

		known[name] = spec
	}

	// Persist updated state
	f.known.Store(known)

	return nil
}

func (f *featureGate) GetAll() map[Feature]FeatureSpec {
	retval := map[Feature]FeatureSpec{}
	for k, v := range f.known.Load().(map[Feature]FeatureSpec) {
		retval[k] = v
	}
	return retval
}

func (f *featureGate) Enabled(key Feature) bool {
	if v, ok := f.enabled.Load().(map[Feature]bool)[key]; ok {
		return v
	}
	if v, ok := f.known.Load().(map[Feature]FeatureSpec)[key]; ok {
		return v.Default
	}

	panic(fmt.Errorf("feature %q is not registered in FeatureGate %q", key, f.featureGateName))
}

func (f *featureGate) AddFlag(fs *pflag.FlagSet) {
	f.lock.Lock()
	f.closed = true
	f.lock.Unlock()

	known := f.KnownFeatures()
	fs.Var(f, flagName, ""+
		"A set of key=value pairs that describe feature gates for alpha/experimental features. "+
		"Options are:\n"+strings.Join(known, "\n"))
}

func (f *featureGate) KnownFeatures() []string {
	var known []string
	for k, v := range f.known.Load().(map[Feature]FeatureSpec) {
		if v.PreRelease == GA || v.PreRelease == Deprecated {
			continue
		}
		known = append(known, fmt.Sprintf("%s=true|false (%s - default=%t)", k, v.PreRelease, v.Default))
	}
	sort.Strings(known)
	return known
}

func (f *featureGate) DeepCopy() MutableFeatureGate {
	// Copy existing state.
	known := map[Feature]FeatureSpec{}
	for k, v := range f.known.Load().(map[Feature]FeatureSpec) {
		known[k] = v
	}
	enabled := map[Feature]bool{}
	for k, v := range f.enabled.Load().(map[Feature]bool) {
		enabled[k] = v
	}

	// Store copied state in new atomics.
	knownValue := &atomic.Value{}
	knownValue.Store(known)
	enabledValue := &atomic.Value{}
	enabledValue.Store(enabled)

	// Construct a new featureGate around the copied state.
	// Note that specialFeatures is treated as immutable by convention,
	// and we maintain the value of f.closed across the copy.
	return &featureGate{
		special: specialFeatures,
		known:   knownValue,
		enabled: enabledValue,
		closed:  f.closed,
	}
}
