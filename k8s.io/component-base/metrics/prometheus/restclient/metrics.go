package restclient

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"time"

	"k8s.io/client-go/tools/metrics"
	k8smetrics "k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

var (
	// 请求耗时
	requestLatency = k8smetrics.NewHistogramVec(
		&k8smetrics.HistogramOpts{
			Name:           "rest_client_request_duration_seconds",
			Help:           "Request latency in seconds. Broken down by verb, and host.",
			StabilityLevel: k8smetrics.ALPHA,
			Buckets:        []float64{0.005, 0.025, 0.1, 0.25, 0.5, 1.0, 2.0, 4.0, 8.0, 15.0, 30.0, 60.0},
		},
		[]string{"verb", "host"},
	)
	// 请求body大小
	requestSize = k8smetrics.NewHistogramVec(
		&k8smetrics.HistogramOpts{
			Name:           "rest_client_request_size_bytes",
			Help:           "Request size in bytes. Broken down by verb and host.",
			StabilityLevel: k8smetrics.ALPHA,
			// 64 bytes to 16MB
			Buckets: []float64{64, 256, 512, 1024, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216},
		},
		[]string{"verb", "host"},
	)
	// 响应大小
	responseSize = k8smetrics.NewHistogramVec(
		&k8smetrics.HistogramOpts{
			Name:           "rest_client_response_size_bytes",
			Help:           "Response size in bytes. Broken down by verb and host.",
			StabilityLevel: k8smetrics.ALPHA,
			// 64 bytes to 16MB
			Buckets: []float64{64, 256, 512, 1024, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216},
		},
		[]string{"verb", "host"},
	)
	// 速率
	rateLimiterLatency = k8smetrics.NewHistogramVec(
		&k8smetrics.HistogramOpts{
			Name:    "rest_client_rate_limiter_duration_seconds",
			Help:    "Client side rate limiter latency in seconds. Broken down by verb, and host.",
			Buckets: []float64{0.005, 0.025, 0.1, 0.25, 0.5, 1.0, 2.0, 4.0, 8.0, 15.0, 30.0, 60.0},
		},
		[]string{"verb", "host"},
	)
	// 请求计数
	requestResult = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Name: "rest_client_requests_total",
			Help: "Number of HTTP requests, partitioned by status code, method, and host.",
		},
		[]string{"code", "method", "host"},
	)

	execPluginCertTTLAdapter = &expiryToTTLAdapter{}

	execPluginCertTTL = k8smetrics.NewGaugeFunc(
		&k8smetrics.GaugeOpts{
			Name: "rest_client_exec_plugin_ttl_seconds",
			Help: "Gauge of the shortest TTL (time-to-live) of the client " +
				"certificate(s) managed by the auth exec plugin. The value " +
				"is in seconds until certificate expiry (negative if " +
				"already expired). If auth exec plugins are unused or manage no " +
				"TLS certificates, the value will be +INF.",
			StabilityLevel: k8smetrics.ALPHA,
		},
		func() float64 {
			if execPluginCertTTLAdapter.e == nil {
				return math.Inf(1)
			}
			return execPluginCertTTLAdapter.e.Sub(time.Now()).Seconds()
		},
	)

	execPluginCertRotation = k8smetrics.NewHistogram(
		&k8smetrics.HistogramOpts{
			Name: "rest_client_exec_plugin_certificate_rotation_age",
			Help: "Histogram of the number of seconds the last auth exec " +
				"plugin client certificate lived before being rotated. " +
				"If auth exec plugin client certificates are unused, " +
				"histogram will contain no data.",
			Buckets: []float64{
				600,       // 10 minutes
				1800,      // 30 minutes
				3600,      // 1  hour
				14400,     // 4  hours
				86400,     // 1  day
				604800,    // 1  week
				2592000,   // 1  month
				7776000,   // 3  months
				15552000,  // 6  months
				31104000,  // 1  year
				124416000, // 4  years
			},
		},
	)

	execPluginCalls = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Name: "rest_client_exec_plugin_call_total",
			Help: "Number of calls to an exec plugin, partitioned by the type of " +
				"event encountered (no_error, plugin_execution_error, plugin_not_found_error, " +
				"client_internal_error) and an optional exit code. The exit code will " +
				"be set to 0 if and only if the plugin call was successful.",
		},
		[]string{"code", "call_status"},
	)
)

func init() {
	legacyregistry.MustRegister(requestLatency)
	legacyregistry.MustRegister(requestSize)
	legacyregistry.MustRegister(responseSize)
	legacyregistry.MustRegister(rateLimiterLatency)
	legacyregistry.MustRegister(requestResult)
	// 原生的注册
	legacyregistry.RawMustRegister(execPluginCertTTL)
	legacyregistry.MustRegister(execPluginCertRotation)
	// 上面是要注册到prometheus，下面是为了设置成包级别的，方便使用
	metrics.Register(metrics.RegisterOpts{
		ClientCertExpiry:      execPluginCertTTLAdapter,
		ClientCertRotationAge: &rotationAdapter{m: execPluginCertRotation},
		RequestLatency:        &latencyAdapter{m: requestLatency},
		RequestSize:           &sizeAdapter{m: requestSize},
		ResponseSize:          &sizeAdapter{m: responseSize},
		RateLimiterLatency:    &latencyAdapter{m: rateLimiterLatency},
		RequestResult:         &resultAdapter{requestResult},
		ExecPluginCalls:       &callsAdapter{m: execPluginCalls},
	})
}

type latencyAdapter struct {
	m *k8smetrics.HistogramVec
}

func (l *latencyAdapter) Observe(ctx context.Context, verb string, u url.URL, latency time.Duration) {
	l.m.WithContext(ctx).WithLabelValues(verb, u.Host).Observe(latency.Seconds())
}

type sizeAdapter struct {
	m *k8smetrics.HistogramVec
}

func (s *sizeAdapter) Observe(ctx context.Context, verb string, host string, size float64) {
	s.m.WithContext(ctx).WithLabelValues(verb, host).Observe(size)
}

type resultAdapter struct {
	m *k8smetrics.CounterVec
}

func (r *resultAdapter) Increment(ctx context.Context, code, method, host string) {
	r.m.WithContext(ctx).WithLabelValues(code, method, host).Inc()
}

type expiryToTTLAdapter struct {
	e *time.Time
}

func (e *expiryToTTLAdapter) Set(expiry *time.Time) {
	e.e = expiry
}

type rotationAdapter struct {
	m *k8smetrics.Histogram
}

func (r *rotationAdapter) Observe(d time.Duration) {
	r.m.Observe(d.Seconds())
}

type callsAdapter struct {
	m *k8smetrics.CounterVec
}

func (r *callsAdapter) Increment(code int, callStatus string) {
	r.m.WithLabelValues(fmt.Sprintf("%d", code), callStatus).Inc()
}
