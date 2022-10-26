package webhook

import (
	"context"
)

// AuthorizerMetrics specifies a set of methods that are used to register various metrics for the webhook authorizer
type AuthorizerMetrics struct {
	// RecordRequestTotal increments the total number of requests for the webhook authorizer
	RecordRequestTotal func(ctx context.Context, code string)

	// RecordRequestLatency measures request latency in seconds for webhooks. Broken down by status code.
	RecordRequestLatency func(ctx context.Context, code string, latency float64)
}

type noopMetrics struct{}

func (noopMetrics) RecordRequestTotal(context.Context, string)            {}
func (noopMetrics) RecordRequestLatency(context.Context, string, float64) {}
