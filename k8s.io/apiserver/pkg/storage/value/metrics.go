package value

import (
	"sync"
	"time"

	"google.golang.org/grpc/status"
	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

const (
	namespace = "apiserver"
	subsystem = "storage"
)

var (
	transformerLatencies = metrics.NewHistogramVec(
		&metrics.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "transformation_duration_seconds",
			Help:      "Latencies in seconds of value transformation operations.",
			// In-process transformations (ex. AES CBC) complete on the order of 20 microseconds. However, when
			// external KMS is involved latencies may climb into hundreds of milliseconds.
			Buckets:        metrics.ExponentialBuckets(5e-6, 2, 25),
			StabilityLevel: metrics.ALPHA,
		},
		[]string{"transformation_type"},
	)
	transformerOperationsTotal = metrics.NewCounterVec(
		&metrics.CounterOpts{
			Namespace:      namespace,
			Subsystem:      subsystem,
			Name:           "transformation_operations_total",
			Help:           "Total number of transformations.",
			StabilityLevel: metrics.ALPHA,
		},
		[]string{"transformation_type", "transformer_prefix", "status"},
	)

	envelopeTransformationCacheMissTotal = metrics.NewCounter(
		&metrics.CounterOpts{
			Namespace:      namespace,
			Subsystem:      subsystem,
			Name:           "envelope_transformation_cache_misses_total",
			Help:           "Total number of cache misses while accessing key decryption key(KEK).",
			StabilityLevel: metrics.ALPHA,
		},
	)
	dataKeyGenerationLatencies = metrics.NewHistogram(
		&metrics.HistogramOpts{
			Namespace:      namespace,
			Subsystem:      subsystem,
			Name:           "data_key_generation_duration_seconds",
			Help:           "Latencies in seconds of data encryption key(DEK) generation operations.",
			Buckets:        metrics.ExponentialBuckets(5e-6, 2, 14),
			StabilityLevel: metrics.ALPHA,
		},
	)

	dataKeyGenerationFailuresTotal = metrics.NewCounter(
		&metrics.CounterOpts{
			Namespace:      namespace,
			Subsystem:      subsystem,
			Name:           "data_key_generation_failures_total",
			Help:           "Total number of failed data encryption key(DEK) generation operations.",
			StabilityLevel: metrics.ALPHA,
		},
	)
)

var registerMetrics sync.Once

func RegisterMetrics() {
	registerMetrics.Do(func() {
		legacyregistry.MustRegister(transformerLatencies)
		legacyregistry.MustRegister(transformerOperationsTotal)
		legacyregistry.MustRegister(envelopeTransformationCacheMissTotal)
		legacyregistry.MustRegister(dataKeyGenerationLatencies)
		legacyregistry.MustRegister(dataKeyGenerationFailuresTotal)
	})
}

func RecordTransformation(transformationType, transformerPrefix string, start time.Time, err error) {
	transformerOperationsTotal.WithLabelValues(transformationType, transformerPrefix, status.Code(err).String()).Inc()

	switch {
	case err == nil:
		transformerLatencies.WithLabelValues(transformationType).Observe(sinceInSeconds(start))
	}
}

func RecordCacheMiss() {
	envelopeTransformationCacheMissTotal.Inc()
}

func RecordDataKeyGeneration(start time.Time, err error) {
	if err != nil {
		dataKeyGenerationFailuresTotal.Inc()
		return
	}

	dataKeyGenerationLatencies.Observe(sinceInSeconds(start))
}

func sinceInSeconds(start time.Time) float64 {
	return time.Since(start).Seconds()
}
