package metrics

import (
	"sync"

	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"k8s.io/klog/v2"
)

type kubeCollector interface {
	Collector
	lazyKubeMetric
	DeprecatedVersion() *semver.Version
	initializeMetric()
	initializeDeprecatedMetric()
}

type lazyKubeMetric interface {
	Create(*semver.Version) bool
	IsCreated() bool
	IsHidden() bool
	IsDeprecated() bool
}

type lazyMetric struct {
	fqName              string
	isDeprecated        bool
	isHidden            bool
	isCreated           bool
	createLock          sync.RWMutex
	markDeprecationOnce sync.Once
	createOnce          sync.Once
	self                kubeCollector
}

func (r *lazyMetric) IsCreated() bool {
	r.createLock.RLock()
	defer r.createLock.RUnlock()
	return r.isCreated
}

func (r *lazyMetric) lazyInit(self kubeCollector, fqName string) {
	r.fqName = fqName
	r.self = self
}

func (r *lazyMetric) preprocessMetric(version semver.Version) {
	disabledMetricsLock.RLock()
	defer disabledMetricsLock.RUnlock()
	if _, ok := disabledMetrics[r.fqName]; ok {
		r.isHidden = true
		return
	}
	selfVersion := r.self.DeprecatedVersion()
	if selfVersion == nil {
		return
	}
	r.markDeprecationOnce.Do(func() {
		if selfVersion.LTE(version) {
			r.isDeprecated = true
		}

		if ShouldShowHidden() {
			klog.Warningf("Hidden metrics (%s) have been manually overridden, showing this very deprecated metric.", r.fqName)
			return
		}
		if shouldHide(&version, selfVersion) {
			r.isHidden = true
		}
	})
}

func (r *lazyMetric) IsHidden() bool {
	return r.isHidden
}

func (r *lazyMetric) IsDeprecated() bool {
	return r.isDeprecated
}

func (r *lazyMetric) Create(version *semver.Version) bool {
	if version != nil {
		r.preprocessMetric(*version)
	}
	// let's not create if this metric is slated to be hidden
	if r.IsHidden() {
		return false
	}
	r.createOnce.Do(func() {
		r.createLock.Lock()
		defer r.createLock.Unlock()
		r.isCreated = true
		if r.IsDeprecated() {
			r.self.initializeDeprecatedMetric()
		} else {
			r.self.initializeMetric()
		}
	})
	return r.IsCreated()
}

func (r *lazyMetric) ClearState() {
	r.createLock.Lock()
	defer r.createLock.Unlock()

	r.isDeprecated = false
	r.isHidden = false
	r.isCreated = false
	r.markDeprecationOnce = sync.Once{}
	r.createOnce = sync.Once{}
}

func (r *lazyMetric) FQName() string {
	return r.fqName
}

type selfCollector struct {
	metric prometheus.Metric
}

func (c *selfCollector) initSelfCollection(m prometheus.Metric) {
	c.metric = m
}

func (c *selfCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.metric.Desc()
}

func (c *selfCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- c.metric
}

// p204
var noopCounterVec = &prometheus.CounterVec{}
var noopHistogramVec = &prometheus.HistogramVec{}
var noopGaugeVec = &prometheus.GaugeVec{}
var noopObserverVec = &noopObserverVector{}

var noop = &noopMetric{}

type noopMetric struct{}

func (noopMetric) Inc()                             {}
func (noopMetric) Add(float64)                      {}
func (noopMetric) Dec()                             {}
func (noopMetric) Set(float64)                      {}
func (noopMetric) Sub(float64)                      {}
func (noopMetric) Observe(float64)                  {}
func (noopMetric) SetToCurrentTime()                {}
func (noopMetric) Desc() *prometheus.Desc           { return nil }
func (noopMetric) Write(*dto.Metric) error          { return nil }
func (noopMetric) Describe(chan<- *prometheus.Desc) {}
func (noopMetric) Collect(chan<- prometheus.Metric) {}

type noopObserverVector struct{}

func (noopObserverVector) GetMetricWith(prometheus.Labels) (prometheus.Observer, error) {
	return noop, nil
}
func (noopObserverVector) GetMetricWithLabelValues(...string) (prometheus.Observer, error) {
	return noop, nil
}
func (noopObserverVector) With(prometheus.Labels) prometheus.Observer    { return noop }
func (noopObserverVector) WithLabelValues(...string) prometheus.Observer { return noop }
func (noopObserverVector) CurryWith(prometheus.Labels) (prometheus.ObserverVec, error) {
	return noopObserverVec, nil
}
func (noopObserverVector) MustCurryWith(prometheus.Labels) prometheus.ObserverVec {
	return noopObserverVec
}
func (noopObserverVector) Describe(chan<- *prometheus.Desc) {}
func (noopObserverVector) Collect(chan<- prometheus.Metric) {}
