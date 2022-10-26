package node

import (
	"sync"

	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

const nodeAuthorizerSubsystem = "node_authorizer"

var (
	graphActionsDuration = metrics.NewHistogramVec(
		&metrics.HistogramOpts{
			Subsystem:      nodeAuthorizerSubsystem,
			Name:           "graph_actions_duration_seconds",
			Help:           "Histogram of duration of graph actions in node authorizer.",
			StabilityLevel: metrics.ALPHA,
			// Start with 0.1ms with the last bucket being [~200ms, Inf)
			Buckets: metrics.ExponentialBuckets(0.0001, 2, 12),
		},
		[]string{"operation"},
	)
)

var registerMetrics sync.Once

// RegisterMetrics registers metrics for node package.
func RegisterMetrics() {
	registerMetrics.Do(func() {
		legacyregistry.MustRegister(graphActionsDuration)
	})
}
