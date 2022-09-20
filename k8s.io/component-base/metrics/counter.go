package metrics

import (
	"context"

	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type Counter struct {
	CounterMetric
	*CounterOpts
	lazyMetric
	selfCollector
}

var _ Metric = &Counter{}

func NewCounter(opts *CounterOpts) *Counter {
	opts.StabilityLevel.setDefaults()

	kc := &Counter{
		CounterOpts: opts,
		lazyMetric:  lazyMetric{},
	}
	kc.setPrometheusCounter(noop)
	kc.lazyInit(kc, BuildFQName(opts.Namespace, opts.Subsystem, opts.Name))
	return kc
}

func (c *Counter) Desc() *prometheus.Desc {
	return c.metric.Desc()
}

func (c *Counter) Write(to *dto.Metric) error {
	return c.metric.Write(to)
}

// Reset resets the underlying prometheus Counter to start counting from 0 again
func (c *Counter) Reset() {
	if !c.IsCreated() {
		return
	}
	c.setPrometheusCounter(prometheus.NewCounter(c.CounterOpts.toPromCounterOpts()))
}

// setPrometheusCounter sets the underlying CounterMetric object, i.e. the thing that does the measurement.
func (c *Counter) setPrometheusCounter(counter prometheus.Counter) {
	c.CounterMetric = counter
	c.initSelfCollection(counter)
}

// DeprecatedVersion returns a pointer to the Version or nil
func (c *Counter) DeprecatedVersion() *semver.Version {
	return parseSemver(c.CounterOpts.DeprecatedVersion)
}

// initializeMetric invocation creates the actual underlying Counter. Until this method is called
// the underlying counter is a no-op.
func (c *Counter) initializeMetric() {
	c.CounterOpts.annotateStabilityLevel()
	// this actually creates the underlying prometheus counter.
	c.setPrometheusCounter(prometheus.NewCounter(c.CounterOpts.toPromCounterOpts()))
}

// initializeDeprecatedMetric invocation creates the actual (but deprecated) Counter. Until this method
// is called the underlying counter is a no-op.
func (c *Counter) initializeDeprecatedMetric() {
	c.CounterOpts.markDeprecated()
	c.initializeMetric()
}

// WithContext allows the normal Counter metric to pass in context. The context is no-op now.
func (c *Counter) WithContext(ctx context.Context) CounterMetric {
	return c.CounterMetric
}

type CounterVec struct {
	*prometheus.CounterVec
	*CounterOpts
	lazyMetric
	originalLabels []string
}

func NewCounterVec(opts *CounterOpts, labels []string) *CounterVec {
	opts.StabilityLevel.setDefaults()

	fqName := BuildFQName(opts.Namespace, opts.Subsystem, opts.Name)
	allowListLock.RLock()
	if allowList, ok := labelValueAllowLists[fqName]; ok {
		opts.LabelValueAllowLists = allowList
	}
	allowListLock.RUnlock()

	cv := &CounterVec{
		CounterVec:     noopCounterVec,
		CounterOpts:    opts,
		originalLabels: labels,
		lazyMetric:     lazyMetric{},
	}
	cv.lazyInit(cv, fqName)
	return cv
}

func (v *CounterVec) DeprecatedVersion() *semver.Version {
	return parseSemver(v.CounterOpts.DeprecatedVersion)

}

// initializeMetric invocation creates the actual underlying CounterVec. Until this method is called
// the underlying counterVec is a no-op.
func (v *CounterVec) initializeMetric() {
	v.CounterOpts.annotateStabilityLevel()
	v.CounterVec = prometheus.NewCounterVec(v.CounterOpts.toPromCounterOpts(), v.originalLabels)
}

// initializeDeprecatedMetric invocation creates the actual (but deprecated) CounterVec. Until this method is called
// the underlying counterVec is a no-op.
func (v *CounterVec) initializeDeprecatedMetric() {
	v.CounterOpts.markDeprecated()
	v.initializeMetric()
}

func (v *CounterVec) WithLabelValues(lvs ...string) CounterMetric {
	if !v.IsCreated() {
		return noop // return no-op counter
	}
	if v.LabelValueAllowLists != nil {
		v.LabelValueAllowLists.ConstrainToAllowedList(v.originalLabels, lvs)
	}
	return v.CounterVec.WithLabelValues(lvs...)
}

func (v *CounterVec) With(labels map[string]string) CounterMetric {
	if !v.IsCreated() {
		return noop // return no-op counter
	}
	if v.LabelValueAllowLists != nil {
		v.LabelValueAllowLists.ConstrainLabelMap(labels)
	}
	return v.CounterVec.With(labels)
}

func (v *CounterVec) Delete(labels map[string]string) bool {
	if !v.IsCreated() {
		return false // since we haven't created the metric, we haven't deleted a metric with the passed in values
	}
	return v.CounterVec.Delete(labels)
}

// Reset deletes all metrics in this vector.
func (v *CounterVec) Reset() {
	if !v.IsCreated() {
		return
	}

	v.CounterVec.Reset()
}

// WithContext returns wrapped CounterVec with context
func (v *CounterVec) WithContext(ctx context.Context) *CounterVecWithContext {
	return &CounterVecWithContext{
		ctx:        ctx,
		CounterVec: v,
	}
}

// CounterVecWithContext is the wrapper of CounterVec with context.
type CounterVecWithContext struct {
	*CounterVec
	ctx context.Context
}

// WithLabelValues is the wrapper of CounterVec.WithLabelValues.
func (vc *CounterVecWithContext) WithLabelValues(lvs ...string) CounterMetric {
	return vc.CounterVec.WithLabelValues(lvs...)
}

// With is the wrapper of CounterVec.With.
func (vc *CounterVecWithContext) With(labels map[string]string) CounterMetric {
	return vc.CounterVec.With(labels)
}
