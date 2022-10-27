package meta

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

// AmbiguousResourceError is returned if the RESTMapper finds multiple matches for a resource
type AmbiguousResourceError struct {
	PartialResource schema.GroupVersionResource

	MatchingResources []schema.GroupVersionResource
	MatchingKinds     []schema.GroupVersionKind
}

func (e *AmbiguousResourceError) Error() string {
	switch {
	case len(e.MatchingKinds) > 0 && len(e.MatchingResources) > 0:
		return fmt.Sprintf("%v matches multiple resources %v and kinds %v", e.PartialResource, e.MatchingResources, e.MatchingKinds)
	case len(e.MatchingKinds) > 0:
		return fmt.Sprintf("%v matches multiple kinds %v", e.PartialResource, e.MatchingKinds)
	case len(e.MatchingResources) > 0:
		return fmt.Sprintf("%v matches multiple resources %v", e.PartialResource, e.MatchingResources)
	}
	return fmt.Sprintf("%v matches multiple resources or kinds", e.PartialResource)
}

// AmbiguousKindError is returned if the RESTMapper finds multiple matches for a kind
type AmbiguousKindError struct {
	PartialKind schema.GroupVersionKind

	MatchingResources []schema.GroupVersionResource
	MatchingKinds     []schema.GroupVersionKind
}

func (e *AmbiguousKindError) Error() string {
	switch {
	case len(e.MatchingKinds) > 0 && len(e.MatchingResources) > 0:
		return fmt.Sprintf("%v matches multiple resources %v and kinds %v", e.PartialKind, e.MatchingResources, e.MatchingKinds)
	case len(e.MatchingKinds) > 0:
		return fmt.Sprintf("%v matches multiple kinds %v", e.PartialKind, e.MatchingKinds)
	case len(e.MatchingResources) > 0:
		return fmt.Sprintf("%v matches multiple resources %v", e.PartialKind, e.MatchingResources)
	}
	return fmt.Sprintf("%v matches multiple resources or kinds", e.PartialKind)
}

func IsAmbiguousError(err error) bool {
	if err == nil {
		return false
	}
	switch err.(type) {
	case *AmbiguousResourceError, *AmbiguousKindError:
		return true
	default:
		return false
	}
}

// NoResourceMatchError is returned if the RESTMapper can't find any match for a resource
type NoResourceMatchError struct {
	PartialResource schema.GroupVersionResource
}

func (e *NoResourceMatchError) Error() string {
	return fmt.Sprintf("no matches for %v", e.PartialResource)
}

// NoKindMatchError is returned if the RESTMapper can't find any match for a kind
type NoKindMatchError struct {
	// GroupKind is the API group and kind that was searched
	GroupKind schema.GroupKind
	// SearchedVersions is the optional list of versions the search was restricted to
	SearchedVersions []string
}

func (e *NoKindMatchError) Error() string {
	searchedVersions := sets.NewString()
	for _, v := range e.SearchedVersions {
		searchedVersions.Insert(schema.GroupVersion{Group: e.GroupKind.Group, Version: v}.String())
	}

	switch len(searchedVersions) {
	case 0:
		return fmt.Sprintf("no matches for kind %q in group %q", e.GroupKind.Kind, e.GroupKind.Group)
	case 1:
		return fmt.Sprintf("no matches for kind %q in version %q", e.GroupKind.Kind, searchedVersions.List()[0])
	default:
		return fmt.Sprintf("no matches for kind %q in versions %q", e.GroupKind.Kind, searchedVersions.List())
	}
}

func IsNoMatchError(err error) bool {
	if err == nil {
		return false
	}
	switch err.(type) {
	case *NoResourceMatchError, *NoKindMatchError:
		return true
	default:
		return false
	}
}
