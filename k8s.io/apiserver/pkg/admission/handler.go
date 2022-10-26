package admission

import (
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	timeToWaitForReady = 10 * time.Second
)

type ReadyFunc func() bool

type Handler struct {
	operations sets.String
	readyFunc  ReadyFunc
}

func (h *Handler) Handles(operation Operation) bool {
	return h.operations.Has(string(operation))
}

func NewHandler(ops ...Operation) *Handler {
	operations := sets.NewString()
	for _, op := range ops {
		operations.Insert(string(op))
	}
	return &Handler{
		operations: operations,
	}
}

func (h *Handler) SetReadyFunc(readyFunc ReadyFunc) {
	h.readyFunc = readyFunc
}

func (h *Handler) WaitForReady() bool {
	// there is no ready func configured, so we return immediately
	if h.readyFunc == nil {
		return true
	}

	timeout := time.After(timeToWaitForReady)
	for !h.readyFunc() {
		select {
		case <-time.After(100 * time.Millisecond):
		case <-timeout:
			return h.readyFunc()
		}
	}
	return true
}
