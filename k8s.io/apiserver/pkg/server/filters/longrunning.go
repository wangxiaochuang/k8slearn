package filters

import (
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
)

func BasicLongRunningRequestCheck(longRunningVerbs, longRunningSubresources sets.String) apirequest.LongRunningRequestCheck {
	return func(r *http.Request, requestInfo *apirequest.RequestInfo) bool {
		if longRunningVerbs.Has(requestInfo.Verb) {
			return true
		}
		if requestInfo.IsResourceRequest && longRunningSubresources.Has(requestInfo.Subresource) {
			return true
		}
		if !requestInfo.IsResourceRequest && strings.HasPrefix(requestInfo.Path, "/debug/pprof/") {
			return true
		}
		return false
	}
}
