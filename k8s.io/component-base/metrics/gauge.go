package metrics

import (
	"context"

	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/component-base/version"
)

type Gauge struct {
	GaugeMetric
	*GaugeOpts
	lazyMetric
	selfCollector
}

func NewGauge(opts *GaugeOpts) *Gauge {
	opts.StabilityLevel.setDefaults()

	kc := &Gauge{
		GaugeOpts:  opts,
		lazyMetric: lazyMetric{},
	}
	kc.setPrometheusGauge(noop)
	kc.lazyInit(kc, BuildFQName(opts.Namespace, opts.Subsystem, opts.Name))
	return kc
}

func (g *Gauge) setPrometheusGauge(gauge prometheus.Gauge) {
	g.GaugeMetric = gauge
	g.initSelfCollection(gauge)
}

func (g *Gauge) DeprecatedVersion() *semver.Version {
	return parseSemver(g.GaugeOpts.DeprecatedVersion)
}

func (g *Gauge) initializeMetric() {
	g.GaugeOpts.annotateStabilityLevel()
	g.setPrometheusGauge(prometheus.NewGauge(g.GaugeOpts.toPromGaugeOpts()))
}

func (g *Gauge) initializeDeprecatedMetric() {
	g.GaugeOpts.markDeprecated()
	g.initializeMetric()
}

func (g *Gauge) WithContext(ctx context.Context) GaugeMetric {
	return g.GaugeMetric
}

type GaugeVec struct {
	*prometheus.GaugeVec
	*GaugeOpts
	lazyMetric
	originalLabels []string
}

func NewGaugeVec(opts *GaugeOpts, labels []string) *GaugeVec {
	// 默认是ALPHA
	opts.StabilityLevel.setDefaults()

	// 组合成一个名字
	fqName := BuildFQName(opts.Namespace, opts.Subsystem, opts.Name)
	allowListLock.RLock()
	if allowList, ok := labelValueAllowLists[fqName]; ok {
		opts.LabelValueAllowLists = allowList
	}
	allowListLock.RUnlock()

	cv := &GaugeVec{
		GaugeVec:       noopGaugeVec,
		GaugeOpts:      opts,
		originalLabels: labels,
		lazyMetric:     lazyMetric{},
	}
	cv.lazyInit(cv, fqName)
	// lazyMetric.Create
	return cv
}

func (v *GaugeVec) DeprecatedVersion() *semver.Version {
	return parseSemver(v.GaugeOpts.DeprecatedVersion)
}

func (v *GaugeVec) initializeMetric() {
	v.GaugeOpts.annotateStabilityLevel()
	v.GaugeVec = prometheus.NewGaugeVec(v.GaugeOpts.toPromGaugeOpts(), v.originalLabels)
}

func (v *GaugeVec) initializeDeprecatedMetric() {
	v.GaugeOpts.markDeprecated()
	v.initializeMetric()
}

func (v *GaugeVec) WithLabelValues(lvs ...string) GaugeMetric {
	if !v.IsCreated() {
		return noop
	}
	if v.LabelValueAllowLists != nil {
		v.LabelValueAllowLists.ConstrainToAllowedList(v.originalLabels, lvs)
	}
	return v.GaugeVec.WithLabelValues(lvs...)
}

func (v *GaugeVec) With(labels map[string]string) GaugeMetric {
	if !v.IsCreated() {
		return noop // return no-op gauge
	}
	if v.LabelValueAllowLists != nil {
		v.LabelValueAllowLists.ConstrainLabelMap(labels)
	}
	return v.GaugeVec.With(labels)
}

func (v *GaugeVec) Delete(labels map[string]string) bool {
	if !v.IsCreated() {
		return false // since we haven't created the metric, we haven't deleted a metric with the passed in values
	}
	return v.GaugeVec.Delete(labels)
}

func (v *GaugeVec) Reset() {
	if !v.IsCreated() {
		return
	}

	v.GaugeVec.Reset()
}

func newGaugeFunc(opts *GaugeOpts, function func() float64, v semver.Version) GaugeFunc {
	g := NewGauge(opts)

	if !g.Create(&v) {
		return nil
	}

	return prometheus.NewGaugeFunc(g.GaugeOpts.toPromGaugeOpts(), function)
}

func NewGaugeFunc(opts *GaugeOpts, function func() float64) GaugeFunc {
	v := parseVersion(version.Get())

	return newGaugeFunc(opts, function, v)
}

func (v *GaugeVec) WithContext(ctx context.Context) *GaugeVecWithContext {
	return &GaugeVecWithContext{
		ctx:      ctx,
		GaugeVec: v,
	}
}

type GaugeVecWithContext struct {
	*GaugeVec
	ctx context.Context
}

func (vc *GaugeVecWithContext) WithLabelValues(lvs ...string) GaugeMetric {
	return vc.GaugeVec.WithLabelValues(lvs...)
}

func (vc *GaugeVecWithContext) With(labels map[string]string) GaugeMetric {
	return vc.GaugeVec.With(labels)
}
