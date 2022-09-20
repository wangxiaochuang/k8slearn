package flowcontrol

import (
	"net/http"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/endpoints/request"
)

var readOnlyVerbs = sets.NewString("get", "list", "watch", "proxy")

type watchIdentifier struct {
	apiGroup  string
	resource  string
	namespace string
	name      string
}

type ForgetWatchFunc func()

type WatchTracker interface {
	RegisterWatch(r *http.Request) ForgetWatchFunc
	GetInterestedWatchCount(requestInfo *request.RequestInfo) int
}
