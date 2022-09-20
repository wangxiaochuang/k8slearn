package flowcontrol

import (
	"context"

	flowcontrol "k8s.io/api/flowcontrol/v1beta2"
	"k8s.io/apiserver/pkg/server/mux"
	fq "k8s.io/apiserver/pkg/util/flowcontrol/fairqueuing"
	fcrequest "k8s.io/apiserver/pkg/util/flowcontrol/request"
)

// p43
type Interface interface {
	Handle(ctx context.Context,
		requestDigest RequestDigest,
		noteFn func(fs *flowcontrol.FlowSchema, pl *flowcontrol.PriorityLevelConfiguration, flowDistinguisher string),
		workEstimator func() fcrequest.WorkEstimate,
		queueNoteFn fq.QueueNoteFn,
		execFn func(),
	)

	// MaintainObservations is a helper for maintaining statistics.
	MaintainObservations(stopCh <-chan struct{})

	Run(stopCh <-chan struct{}) error

	// Install installs debugging endpoints to the web-server.
	Install(c *mux.PathRecorderMux)

	// WatchTracker provides the WatchTracker interface.
	WatchTracker
}
