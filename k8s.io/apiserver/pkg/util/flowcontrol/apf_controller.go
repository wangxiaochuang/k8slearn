package flowcontrol

import (
	"sync"
	"time"

	flowcontrol "k8s.io/api/flowcontrol/v1beta2"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/util/apihelpers"
	fq "k8s.io/apiserver/pkg/util/flowcontrol/fairqueuing"
	"k8s.io/apiserver/pkg/util/flowcontrol/metrics"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"

	flowcontrolclient "k8s.io/client-go/kubernetes/typed/flowcontrol/v1beta2"
	flowcontrollister "k8s.io/client-go/listers/flowcontrol/v1beta2"
)

// p91
type RequestDigest struct {
	RequestInfo *request.RequestInfo
	User        user.Info
}

// p101
type configController struct {
	name                  string // varies in tests of fighting controllers
	clock                 clock.PassiveClock
	queueSetFactory       fq.QueueSetFactory
	reqsObsPairGenerator  metrics.RatioedChangeObserverPairGenerator
	execSeatsObsGenerator metrics.RatioedChangeObserverGenerator

	// How this controller appears in an ObjectMeta ManagedFieldsEntry.Manager
	asFieldManager string

	// Given a boolean indicating whether a FlowSchema's referenced
	// PriorityLevelConfig exists, return a boolean indicating whether
	// the reference is dangling
	foundToDangling func(bool) bool

	// configQueue holds `(interface{})(0)` when the configuration
	// objects need to be reprocessed.
	configQueue workqueue.RateLimitingInterface

	plLister         flowcontrollister.PriorityLevelConfigurationLister
	plInformerSynced cache.InformerSynced

	fsLister         flowcontrollister.FlowSchemaLister
	fsInformerSynced cache.InformerSynced

	flowcontrolClient flowcontrolclient.FlowcontrolV1beta2Interface

	// serverConcurrencyLimit is the limit on the server's total
	// number of non-exempt requests being served at once.  This comes
	// from server configuration.
	serverConcurrencyLimit int

	// requestWaitLimit comes from server configuration.
	requestWaitLimit time.Duration

	// watchTracker implements the necessary WatchTracker interface.
	WatchTracker

	mostRecentUpdates []updateAttempt

	lock sync.RWMutex

	flowSchemas apihelpers.FlowSchemaSequence

	priorityLevelStates map[string]*priorityLevelState
}

// p167
type updateAttempt struct {
	timeUpdated  time.Time
	updatedItems sets.String // FlowSchema names
}

type priorityLevelState struct {
	// the API object or prototype prescribing this level.  Nothing
	// reached through this pointer is mutable.
	pl *flowcontrol.PriorityLevelConfiguration

	// qsCompleter holds the QueueSetCompleter derived from `config`
	// and `queues` if config is not exempt, nil otherwise.
	qsCompleter fq.QueueSetCompleter

	// The QueueSet for this priority level.  This is nil if and only
	// if the priority level is exempt.
	queues fq.QueueSet

	// quiescing==true indicates that this priority level should be
	// removed when its queues have all drained.  May be true only if
	// queues is non-nil.
	quiescing bool

	// number of goroutines between Controller::Match and calling the
	// returned StartFunction
	numPending int

	// Observers tracking number of requests waiting, executing
	reqsObsPair metrics.RatioedChangeObserverPair

	// Observer of number of seats occupied throughout execution
	execSeatsObs metrics.RatioedChangeObserver
}

// p203
func newTestableController(config TestableConfig) *configController {
	cfgCtlr := &configController{
		name:                   config.Name,
		clock:                  config.Clock,
		queueSetFactory:        config.QueueSetFactory,
		reqsObsPairGenerator:   config.ReqsObsPairGenerator,
		execSeatsObsGenerator:  config.ExecSeatsObsGenerator,
		asFieldManager:         config.AsFieldManager,
		foundToDangling:        config.FoundToDangling,
		serverConcurrencyLimit: config.ServerConcurrencyLimit,
		requestWaitLimit:       config.RequestWaitLimit,
		flowcontrolClient:      config.FlowcontrolClient,
		priorityLevelStates:    make(map[string]*priorityLevelState),
		WatchTracker:           NewWatchTracker(),
	}
	klog.V(2).Infof("NewTestableController %q with serverConcurrencyLimit=%d, requestWaitLimit=%s, name=%s, asFieldManager=%q", cfgCtlr.name, cfgCtlr.serverConcurrencyLimit, cfgCtlr.requestWaitLimit, cfgCtlr.name, cfgCtlr.asFieldManager)
	// Start with longish delay because conflicts will be between
	// different processes, so take some time to go away.
	cfgCtlr.configQueue = workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(200*time.Millisecond, 8*time.Hour), "priority_and_fairness_config_queue")
	// ensure the data structure reflects the mandatory config
	cfgCtlr.lockAndDigestConfigObjects(nil, nil)
	fci := config.InformerFactory.Flowcontrol().V1beta2()
	pli := fci.PriorityLevelConfigurations()
	fsi := fci.FlowSchemas()
	cfgCtlr.plLister = pli.Lister()
	cfgCtlr.plInformerSynced = pli.Informer().HasSynced
	cfgCtlr.fsLister = fsi.Lister()
	cfgCtlr.fsInformerSynced = fsi.Informer().HasSynced
	pli.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pl := obj.(*flowcontrol.PriorityLevelConfiguration)
			klog.V(7).Infof("Triggered API priority and fairness config reloading in %s due to creation of PLC %s", cfgCtlr.name, pl.Name)
			cfgCtlr.configQueue.Add(0)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newPL := newObj.(*flowcontrol.PriorityLevelConfiguration)
			oldPL := oldObj.(*flowcontrol.PriorityLevelConfiguration)
			if !apiequality.Semantic.DeepEqual(oldPL.Spec, newPL.Spec) {
				klog.V(7).Infof("Triggered API priority and fairness config reloading in %s due to spec update of PLC %s", cfgCtlr.name, newPL.Name)
				cfgCtlr.configQueue.Add(0)
			} else {
				klog.V(7).Infof("No trigger API priority and fairness config reloading in %s due to spec non-change of PLC %s", cfgCtlr.name, newPL.Name)
			}
		},
		DeleteFunc: func(obj interface{}) {
			name, _ := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			klog.V(7).Infof("Triggered API priority and fairness config reloading in %s due to deletion of PLC %s", cfgCtlr.name, name)
			cfgCtlr.configQueue.Add(0)

		}})
	fsi.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fs := obj.(*flowcontrol.FlowSchema)
			klog.V(7).Infof("Triggered API priority and fairness config reloading in %s due to creation of FS %s", cfgCtlr.name, fs.Name)
			cfgCtlr.configQueue.Add(0)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newFS := newObj.(*flowcontrol.FlowSchema)
			oldFS := oldObj.(*flowcontrol.FlowSchema)
			// Changes to either Spec or Status are relevant.  The
			// concern is that we might, in some future release, want
			// different behavior than is implemented now. One of the
			// hardest questions is how does an operator roll out the
			// new release in a cluster with multiple kube-apiservers
			// --- in a way that works no matter what servers crash
			// and restart when. If this handler reacts only to
			// changes in Spec then we have a scenario in which the
			// rollout leaves the old Status in place. The scenario
			// ends with this subsequence: deploy the last new server
			// before deleting the last old server, and in between
			// those two operations the last old server crashes and
			// recovers. The chosen solution is making this controller
			// insist on maintaining the particular state that it
			// establishes.
			if !(apiequality.Semantic.DeepEqual(oldFS.Spec, newFS.Spec) &&
				apiequality.Semantic.DeepEqual(oldFS.Status, newFS.Status)) {
				klog.V(7).Infof("Triggered API priority and fairness config reloading in %s due to spec and/or status update of FS %s", cfgCtlr.name, newFS.Name)
				cfgCtlr.configQueue.Add(0)
			} else {
				klog.V(7).Infof("No trigger of API priority and fairness config reloading in %s due to spec and status non-change of FS %s", cfgCtlr.name, newFS.Name)
			}
		},
		DeleteFunc: func(obj interface{}) {
			name, _ := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			klog.V(7).Infof("Triggered API priority and fairness config reloading in %s due to deletion of FS %s", cfgCtlr.name, name)
			cfgCtlr.configQueue.Add(0)

		}})
	return cfgCtlr
}
