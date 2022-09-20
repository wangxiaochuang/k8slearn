package controlplane

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
)

type Config struct {
}

// p662
var (
	// stableAPIGroupVersionsEnabledByDefault is a list of our stable versions.
	stableAPIGroupVersionsEnabledByDefault = []schema.GroupVersion{}

	// legacyBetaEnabledByDefaultResources is the list of beta resources we enable.  You may only add to this list
	// if your resource is already enabled by default in a beta level we still serve AND there is no stable API for it.
	// see https://github.com/kubernetes/enhancements/tree/master/keps/sig-architecture/3136-beta-apis-off-by-default
	// for more details.
	legacyBetaEnabledByDefaultResources = []schema.GroupVersionResource{}
	// betaAPIGroupVersionsDisabledByDefault is for all future beta groupVersions.
	betaAPIGroupVersionsDisabledByDefault = []schema.GroupVersion{}

	// alphaAPIGroupVersionsDisabledByDefault holds the alpha APIs we have.  They are always disabled by default.
	alphaAPIGroupVersionsDisabledByDefault = []schema.GroupVersion{}
)

// p728
func DefaultAPIResourceConfigSource() *serverstorage.ResourceConfig {
	ret := serverstorage.NewResourceConfig()
	ret.EnableVersions(stableAPIGroupVersionsEnabledByDefault...)

	ret.DisableVersions(betaAPIGroupVersionsDisabledByDefault...)
	ret.DisableVersions(alphaAPIGroupVersionsDisabledByDefault...)

	ret.EnableResources(legacyBetaEnabledByDefaultResources...)

	return ret
}
