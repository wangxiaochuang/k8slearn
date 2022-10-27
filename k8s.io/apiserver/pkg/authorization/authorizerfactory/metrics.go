package authorizerfactory

import (
	"context"

	compbasemetrics "k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

type registerables []compbasemetrics.Registerable

// init registers all metrics
func init() {
	for _, metric := range metrics {
		legacyregistry.MustRegister(metric)
	}
}

var (
	requestTotal = compbasemetrics.NewCounterVec(
		&compbasemetrics.CounterOpts{
			Name:           "apiserver_delegated_authz_request_total",
			Help:           "Number of HTTP requests partitioned by status code.",
			StabilityLevel: compbasemetrics.ALPHA,
		},
		[]string{"code"},
	)

	requestLatency = compbasemetrics.NewHistogramVec(
		&compbasemetrics.HistogramOpts{
			Name:           "apiserver_delegated_authz_request_duration_seconds",
			Help:           "Request latency in seconds. Broken down by status code.",
			Buckets:        []float64{0.25, 0.5, 0.7, 1, 1.5, 3, 5, 10},
			StabilityLevel: compbasemetrics.ALPHA,
		},
		[]string{"code"},
	)

	metrics = registerables{
		requestTotal,
		requestLatency,
	}
)

// RecordRequestTotal increments the total number of requests for the delegated authorization.
func RecordRequestTotal(ctx context.Context, code string) {
	requestTotal.WithContext(ctx).WithLabelValues(code).Add(1)
}

// RecordRequestLatency measures request latency in seconds for the delegated authorization. Broken down by status code.
func RecordRequestLatency(ctx context.Context, code string, latency float64) {
	requestLatency.WithContext(ctx).WithLabelValues(code).Observe(latency)
}