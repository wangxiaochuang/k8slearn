package features

import (
	"k8s.io/apimachinery/pkg/util/runtime"

	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/featuregate"
)

const (
	AdvancedAuditing featuregate.Feature = "AdvancedAuditing"

	APIResponseCompression featuregate.Feature = "APIResponseCompression"

	APIListChunking featuregate.Feature = "APIListChunking"

	DryRun featuregate.Feature = "DryRun"

	RemainingItemCount featuregate.Feature = "RemainingItemCount"

	ServerSideApply featuregate.Feature = "ServerSideApply"

	StorageVersionHash featuregate.Feature = "StorageVersionHash"

	StorageVersionAPI featuregate.Feature = "StorageVersionAPI"

	WatchBookmark featuregate.Feature = "WatchBookmark"

	APIPriorityAndFairness featuregate.Feature = "APIPriorityAndFairness"

	RemoveSelfLink featuregate.Feature = "RemoveSelfLink"

	SelectorIndex featuregate.Feature = "SelectorIndex"

	EfficientWatchResumption featuregate.Feature = "EfficientWatchResumption"

	APIServerIdentity featuregate.Feature = "APIServerIdentity"

	APIServerTracing featuregate.Feature = "APIServerTracing"

	OpenAPIEnums featuregate.Feature = "OpenAPIEnums"

	CustomResourceValidationExpressions featuregate.Feature = "CustomResourceValidationExpressions"

	OpenAPIV3 featuregate.Feature = "OpenAPIV3"

	ServerSideFieldValidation featuregate.Feature = "ServerSideFieldValidation"
)

func init() {
	runtime.Must(utilfeature.DefaultMutableFeatureGate.Add(defaultKubernetesFeatureGates))
}

// defaultKubernetesFeatureGates consists of all known Kubernetes-specific feature keys.
// To add a new feature, define a key for it above and add it here. The features will be
// available throughout Kubernetes binaries.
var defaultKubernetesFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{
	AdvancedAuditing:                    {Default: true, PreRelease: featuregate.GA},
	APIResponseCompression:              {Default: true, PreRelease: featuregate.Beta},
	APIListChunking:                     {Default: true, PreRelease: featuregate.Beta},
	DryRun:                              {Default: true, PreRelease: featuregate.GA},
	RemainingItemCount:                  {Default: true, PreRelease: featuregate.Beta},
	ServerSideApply:                     {Default: true, PreRelease: featuregate.GA},
	StorageVersionHash:                  {Default: true, PreRelease: featuregate.Beta},
	StorageVersionAPI:                   {Default: false, PreRelease: featuregate.Alpha},
	WatchBookmark:                       {Default: true, PreRelease: featuregate.GA, LockToDefault: true},
	APIPriorityAndFairness:              {Default: true, PreRelease: featuregate.Beta},
	RemoveSelfLink:                      {Default: true, PreRelease: featuregate.GA, LockToDefault: true},
	SelectorIndex:                       {Default: true, PreRelease: featuregate.GA, LockToDefault: true},
	EfficientWatchResumption:            {Default: true, PreRelease: featuregate.GA, LockToDefault: true},
	APIServerIdentity:                   {Default: false, PreRelease: featuregate.Alpha},
	APIServerTracing:                    {Default: false, PreRelease: featuregate.Alpha},
	OpenAPIEnums:                        {Default: true, PreRelease: featuregate.Beta},
	CustomResourceValidationExpressions: {Default: false, PreRelease: featuregate.Alpha},
	OpenAPIV3:                           {Default: true, PreRelease: featuregate.Beta},
	ServerSideFieldValidation:           {Default: false, PreRelease: featuregate.Alpha},
}
