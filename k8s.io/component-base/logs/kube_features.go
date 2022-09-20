package logs

import (
	"k8s.io/component-base/featuregate"
)

const (
	ContextualLogging featuregate.Feature = "ContextualLogging"

	contextualLoggingDefault = false
)

func featureGates() map[featuregate.Feature]featuregate.FeatureSpec {
	return map[featuregate.Feature]featuregate.FeatureSpec{
		ContextualLogging: {Default: contextualLoggingDefault, PreRelease: featuregate.Alpha},
	}
}

func AddFeatureGates(mutableFeatureGate featuregate.MutableFeatureGate) error {
	return mutableFeatureGate.Add(featureGates())
}
