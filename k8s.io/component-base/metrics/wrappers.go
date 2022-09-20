package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type Collector interface {
	Describe(chan<- *prometheus.Desc)
	Collect(chan<- prometheus.Metric)
}

type Metric interface {
	Desc() *prometheus.Desc
	Write(*dto.Metric) error
}

type CounterMetric interface {
	Inc()
	Add(float64)
}

type CounterVecMetric interface {
	WithLabelValues(...string) CounterMetric
	With(prometheus.Labels) CounterMetric
}

type GaugeMetric interface {
	Set(float64)
	Inc()
	Dec()
	Add(float64)
	Write(out *dto.Metric) error
	SetToCurrentTime()
}

type ObserverMetric interface {
	Observe(float64)
}

type PromRegistry interface {
	Register(prometheus.Collector) error
	MustRegister(...prometheus.Collector)
	Unregister(prometheus.Collector) bool
	Gather() ([]*dto.MetricFamily, error)
}

type Gatherer interface {
	prometheus.Gatherer
}

type GaugeFunc interface {
	Metric
	Collector
}
