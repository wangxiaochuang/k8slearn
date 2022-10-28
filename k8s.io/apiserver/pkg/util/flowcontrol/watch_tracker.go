package flowcontrol

import (
	"net/http"
	"sync"

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

// p76
type builtinIndexes map[string]string

func getBuiltinIndexes() builtinIndexes {
	return map[string]string{
		"pods": "spec.nodeName",
	}
}

type watchTracker struct {
	// indexes represents a set of registered indexes.
	// It can't change after creation.
	indexes builtinIndexes

	lock       sync.Mutex
	watchCount map[watchIdentifier]int
}

func NewWatchTracker() WatchTracker {
	return &watchTracker{
		indexes:    getBuiltinIndexes(),
		watchCount: make(map[watchIdentifier]int),
	}
}

const (
	unsetValue = "<unset>"
)

// p132
func (w *watchTracker) RegisterWatch(r *http.Request) ForgetWatchFunc {
	panic("not implemented")
}

// p202
func (w *watchTracker) GetInterestedWatchCount(requestInfo *request.RequestInfo) int {
	panic("not implemented")
}
