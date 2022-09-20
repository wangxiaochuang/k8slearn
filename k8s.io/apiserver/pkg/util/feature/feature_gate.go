package feature

import (
	"k8s.io/component-base/featuregate"
)

var (
	DefaultMutableFeatureGate featuregate.MutableFeatureGate = featuregate.NewFeatureGate()
	DefaultFeatureGate        featuregate.FeatureGate        = DefaultMutableFeatureGate
)
