package metrics

import (
	"context"

	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
)

var DefBuckets = prometheus.DefBuckets

// LinearBuckets is a wrapper for prometheus.LinearBuckets.
func LinearBuckets(start, width float64, count int) []float64 {
	return prometheus.LinearBuckets(start, width, count)
}

// ExponentialBuckets is a wrapper for prometheus.ExponentialBuckets.
func ExponentialBuckets(start, factor float64, count int) []float64 {
	return prometheus.ExponentialBuckets(start, factor, count)
}

type Histogram struct {
	ObserverMetric
	*HistogramOpts
	lazyMetric
	selfCollector
}

// NewHistogram returns an object which is Histogram-like. However, nothing
// will be measured until the histogram is registered somewhere.
func NewHistogram(opts *HistogramOpts) *Histogram {
	opts.StabilityLevel.setDefaults()

	h := &Histogram{
		HistogramOpts: opts,
		lazyMetric:    lazyMetric{},
	}
	h.setPrometheusHistogram(noopMetric{})
	h.lazyInit(h, BuildFQName(opts.Namespace, opts.Subsystem, opts.Name))
	return h
}

func (h *Histogram) setPrometheusHistogram(histogram prometheus.Histogram) {
	h.ObserverMetric = histogram
	h.initSelfCollection(histogram)
}

// DeprecatedVersion returns a pointer to the Version or nil
func (h *Histogram) DeprecatedVersion() *semver.Version {
	return parseSemver(h.HistogramOpts.DeprecatedVersion)
}

// initializeMetric invokes the actual prometheus.Histogram object instantiation
// and stores a reference to it
func (h *Histogram) initializeMetric() {
	h.HistogramOpts.annotateStabilityLevel()
	// this actually creates the underlying prometheus gauge.
	h.setPrometheusHistogram(prometheus.NewHistogram(h.HistogramOpts.toPromHistogramOpts()))
}

// initializeDeprecatedMetric invokes the actual prometheus.Histogram object instantiation
// but modifies the Help description prior to object instantiation.
func (h *Histogram) initializeDeprecatedMetric() {
	h.HistogramOpts.markDeprecated()
	h.initializeMetric()
}

// WithContext allows the normal Histogram metric to pass in context. The context is no-op now.
func (h *Histogram) WithContext(ctx context.Context) ObserverMetric {
	return h.ObserverMetric
}

type HistogramVec struct {
	*prometheus.HistogramVec
	*HistogramOpts
	lazyMetric
	originalLabels []string
}

// p104
func NewHistogramVec(opts *HistogramOpts, labels []string) *HistogramVec {
	opts.StabilityLevel.setDefaults()

	fqName := BuildFQName(opts.Namespace, opts.Subsystem, opts.Name)
	allowListLock.RLock()
	if allowList, ok := labelValueAllowLists[fqName]; ok {
		opts.LabelValueAllowLists = allowList
	}
	allowListLock.RUnlock()

	v := &HistogramVec{
		HistogramVec:   noopHistogramVec,
		HistogramOpts:  opts,
		originalLabels: labels,
		lazyMetric:     lazyMetric{},
	}
	v.lazyInit(v, fqName)
	return v
}

func (v *HistogramVec) DeprecatedVersion() *semver.Version {
	return parseSemver(v.HistogramOpts.DeprecatedVersion)
}

func (v *HistogramVec) initializeMetric() {
	v.HistogramOpts.annotateStabilityLevel()
	v.HistogramVec = prometheus.NewHistogramVec(v.HistogramOpts.toPromHistogramOpts(), v.originalLabels)
}

func (v *HistogramVec) initializeDeprecatedMetric() {
	v.HistogramOpts.markDeprecated()
	v.initializeMetric()
}

func (v *HistogramVec) WithLabelValues(lvs ...string) ObserverMetric {
	if !v.IsCreated() {
		return noop
	}
	if v.LabelValueAllowLists != nil {
		v.LabelValueAllowLists.ConstrainToAllowedList(v.originalLabels, lvs)
	}
	return v.HistogramVec.WithLabelValues(lvs...)
}

func (v *HistogramVec) With(labels map[string]string) ObserverMetric {
	if !v.IsCreated() {
		return noop
	}
	if v.LabelValueAllowLists != nil {
		v.LabelValueAllowLists.ConstrainLabelMap(labels)
	}
	return v.HistogramVec.With(labels)
}

func (v *HistogramVec) Delete(labels map[string]string) bool {
	if !v.IsCreated() {
		return false // since we haven't created the metric, we haven't deleted a metric with the passed in values
	}
	return v.HistogramVec.Delete(labels)
}

// Reset deletes all metrics in this vector.
func (v *HistogramVec) Reset() {
	if !v.IsCreated() {
		return
	}

	v.HistogramVec.Reset()
}

// WithContext returns wrapped HistogramVec with context
func (v *HistogramVec) WithContext(ctx context.Context) *HistogramVecWithContext {
	return &HistogramVecWithContext{
		ctx:          ctx,
		HistogramVec: v,
	}
}

type HistogramVecWithContext struct {
	*HistogramVec
	ctx context.Context
}

// WithLabelValues is the wrapper of HistogramVec.WithLabelValues.
func (vc *HistogramVecWithContext) WithLabelValues(lvs ...string) ObserverMetric {
	return vc.HistogramVec.WithLabelValues(lvs...)
}

// With is the wrapper of HistogramVec.With.
func (vc *HistogramVecWithContext) With(labels map[string]string) ObserverMetric {
	return vc.HistogramVec.With(labels)
}
