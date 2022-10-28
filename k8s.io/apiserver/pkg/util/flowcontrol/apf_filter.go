package flowcontrol

import (
	"context"
	"time"

	"k8s.io/apiserver/pkg/server/mux"
	fq "k8s.io/apiserver/pkg/util/flowcontrol/fairqueuing"
	"k8s.io/apiserver/pkg/util/flowcontrol/fairqueuing/eventclock"
	fqs "k8s.io/apiserver/pkg/util/flowcontrol/fairqueuing/queueset"
	"k8s.io/apiserver/pkg/util/flowcontrol/metrics"
	fcrequest "k8s.io/apiserver/pkg/util/flowcontrol/request"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/utils/clock"

	flowcontrol "k8s.io/api/flowcontrol/v1beta2"
	flowcontrolclient "k8s.io/client-go/kubernetes/typed/flowcontrol/v1beta2"
)

const ConfigConsumerAsFieldManager = "api-priority-and-fairness-config-consumer-v1"

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

// p87
func New(
	informerFactory kubeinformers.SharedInformerFactory,
	flowcontrolClient flowcontrolclient.FlowcontrolV1beta2Interface,
	serverConcurrencyLimit int,
	requestWaitLimit time.Duration,
) Interface {
	clk := eventclock.Real{}
	return NewTestable(TestableConfig{
		Name:                   "Controller",
		Clock:                  clk,
		AsFieldManager:         ConfigConsumerAsFieldManager,
		FoundToDangling:        func(found bool) bool { return !found },
		InformerFactory:        informerFactory,
		FlowcontrolClient:      flowcontrolClient,
		ServerConcurrencyLimit: serverConcurrencyLimit,
		RequestWaitLimit:       requestWaitLimit,
		ReqsObsPairGenerator:   metrics.PriorityLevelConcurrencyObserverPairGenerator,
		ExecSeatsObsGenerator:  metrics.PriorityLevelExecutionSeatsObserverGenerator,
		QueueSetFactory:        fqs.NewQueueSetFactory(clk),
	})
}

type TestableConfig struct {
	Name                   string
	Clock                  clock.PassiveClock
	AsFieldManager         string
	FoundToDangling        func(bool) bool
	InformerFactory        kubeinformers.SharedInformerFactory
	FlowcontrolClient      flowcontrolclient.FlowcontrolV1beta2Interface
	ServerConcurrencyLimit int
	RequestWaitLimit       time.Duration
	ReqsObsPairGenerator   metrics.RatioedChangeObserverPairGenerator
	ExecSeatsObsGenerator  metrics.RatioedChangeObserverGenerator
	QueueSetFactory        fq.QueueSetFactory
}

func NewTestable(config TestableConfig) Interface {
	return newTestableController(config)
}

func (cfgCtlr *configController) Handle(ctx context.Context, requestDigest RequestDigest,
	noteFn func(fs *flowcontrol.FlowSchema, pl *flowcontrol.PriorityLevelConfiguration, flowDistinguisher string),
	workEstimator func() fcrequest.WorkEstimate,
	queueNoteFn fq.QueueNoteFn,
	execFn func()) {
	panic("not implemented")
}
