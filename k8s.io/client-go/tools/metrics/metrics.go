package metrics

import (
	"context"
	"net/url"
	"sync"
	"time"
)

var registerMetrics sync.Once

type DurationMetric interface {
	Observe(duration time.Duration)
}

type ExpiryMetric interface {
	Set(expiry *time.Time)
}

type LatencyMetric interface {
	Observe(ctx context.Context, verb string, u url.URL, latency time.Duration)
}

type SizeMetric interface {
	Observe(ctx context.Context, verb string, host string, size float64)
}

type ResultMetric interface {
	Increment(ctx context.Context, code string, method string, host string)
}

type CallsMetric interface {
	// Increment increments a counter per exitCode and callStatus.
	Increment(exitCode int, callStatus string)
}

var (
	// ClientCertExpiry is the expiry time of a client certificate
	ClientCertExpiry ExpiryMetric = noopExpiry{}
	// ClientCertRotationAge is the age of a certificate that has just been rotated.
	ClientCertRotationAge DurationMetric = noopDuration{}
	// RequestLatency is the latency metric that rest clients will update.
	RequestLatency LatencyMetric = noopLatency{}
	// RequestSize is the request size metric that rest clients will update.
	RequestSize SizeMetric = noopSize{}
	// ResponseSize is the response size metric that rest clients will update.
	ResponseSize SizeMetric = noopSize{}
	// RateLimiterLatency is the client side rate limiter latency metric.
	RateLimiterLatency LatencyMetric = noopLatency{}
	// RequestResult is the result metric that rest clients will update.
	RequestResult ResultMetric = noopResult{}
	// ExecPluginCalls is the number of calls made to an exec plugin, partitioned by
	// exit code and call status.
	ExecPluginCalls CallsMetric = noopCalls{}
)

type RegisterOpts struct {
	ClientCertExpiry      ExpiryMetric
	ClientCertRotationAge DurationMetric
	RequestLatency        LatencyMetric
	RequestSize           SizeMetric
	ResponseSize          SizeMetric
	RateLimiterLatency    LatencyMetric
	RequestResult         ResultMetric
	ExecPluginCalls       CallsMetric
}

func Register(opts RegisterOpts) {
	registerMetrics.Do(func() {
		if opts.ClientCertExpiry != nil {
			ClientCertExpiry = opts.ClientCertExpiry
		}
		if opts.ClientCertRotationAge != nil {
			ClientCertRotationAge = opts.ClientCertRotationAge
		}
		if opts.RequestLatency != nil {
			RequestLatency = opts.RequestLatency
		}
		if opts.RequestSize != nil {
			RequestSize = opts.RequestSize
		}
		if opts.ResponseSize != nil {
			ResponseSize = opts.ResponseSize
		}
		if opts.RateLimiterLatency != nil {
			RateLimiterLatency = opts.RateLimiterLatency
		}
		if opts.RequestResult != nil {
			RequestResult = opts.RequestResult
		}
		if opts.ExecPluginCalls != nil {
			ExecPluginCalls = opts.ExecPluginCalls
		}
	})
}

type noopDuration struct{}

func (noopDuration) Observe(time.Duration) {}

type noopExpiry struct{}

func (noopExpiry) Set(*time.Time) {}

type noopLatency struct{}

func (noopLatency) Observe(context.Context, string, url.URL, time.Duration) {}

type noopSize struct{}

func (noopSize) Observe(context.Context, string, string, float64) {}

type noopResult struct{}

func (noopResult) Increment(context.Context, string, string, string) {}

type noopCalls struct{}

func (noopCalls) Increment(int, string) {}
