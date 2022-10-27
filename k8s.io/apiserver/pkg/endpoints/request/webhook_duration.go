package request

import (
	"context"
	"sync"
	"time"

	"k8s.io/utils/clock"
)

func sumDuration(d1 time.Duration, d2 time.Duration) time.Duration {
	return d1 + d2
}

func maxDuration(d1 time.Duration, d2 time.Duration) time.Duration {
	if d1 > d2 {
		return d1
	}
	return d2
}

type DurationTracker interface {
	Track(f func())
	TrackDuration(time.Duration)
	GetLatency() time.Duration
}

type durationTracker struct {
	clock             clock.Clock
	latency           time.Duration
	mu                sync.Mutex
	aggregateFunction func(time.Duration, time.Duration) time.Duration
}

func (t *durationTracker) Track(f func()) {
	startedAt := t.clock.Now()
	defer func() {
		duration := t.clock.Since(startedAt)
		t.mu.Lock()
		defer t.mu.Unlock()
		t.latency = t.aggregateFunction(t.latency, duration)
	}()

	f()
}

func (t *durationTracker) TrackDuration(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.latency = t.aggregateFunction(t.latency, d)
}

func (t *durationTracker) GetLatency() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.latency
}

func newSumLatencyTracker(c clock.Clock) DurationTracker {
	return &durationTracker{
		clock:             c,
		aggregateFunction: sumDuration,
	}
}

func newMaxLatencyTracker(c clock.Clock) DurationTracker {
	return &durationTracker{
		clock:             c,
		aggregateFunction: maxDuration,
	}
}

type LatencyTrackers struct {
	MutatingWebhookTracker   DurationTracker
	ValidatingWebhookTracker DurationTracker
	StorageTracker           DurationTracker
	TransformTracker         DurationTracker
	SerializationTracker     DurationTracker
	ResponseWriteTracker     DurationTracker
}

type latencyTrackersKeyType int

const latencyTrackersKey latencyTrackersKeyType = iota

func WithLatencyTrackers(parent context.Context) context.Context {
	return WithLatencyTrackersAndCustomClock(parent, clock.RealClock{})
}

func WithLatencyTrackersAndCustomClock(parent context.Context, c clock.Clock) context.Context {
	return WithValue(parent, latencyTrackersKey, &LatencyTrackers{
		MutatingWebhookTracker:   newSumLatencyTracker(c),
		ValidatingWebhookTracker: newMaxLatencyTracker(c),
		StorageTracker:           newSumLatencyTracker(c),
		TransformTracker:         newSumLatencyTracker(c),
		SerializationTracker:     newSumLatencyTracker(c),
		ResponseWriteTracker:     newSumLatencyTracker(c),
	})
}

func LatencyTrackersFrom(ctx context.Context) (*LatencyTrackers, bool) {
	wd, ok := ctx.Value(latencyTrackersKey).(*LatencyTrackers)
	return wd, ok && wd != nil
}

func TrackTransformResponseObjectLatency(ctx context.Context, transform func()) {
	if tracker, ok := LatencyTrackersFrom(ctx); ok {
		tracker.TransformTracker.Track(transform)
		return
	}

	transform()
}

func TrackStorageLatency(ctx context.Context, d time.Duration) {
	if tracker, ok := LatencyTrackersFrom(ctx); ok {
		tracker.StorageTracker.TrackDuration(d)
	}
}

func TrackSerializeResponseObjectLatency(ctx context.Context, f func()) {
	if tracker, ok := LatencyTrackersFrom(ctx); ok {
		tracker.SerializationTracker.Track(f)
		return
	}

	f()
}

func TrackResponseWriteLatency(ctx context.Context, d time.Duration) {
	if tracker, ok := LatencyTrackersFrom(ctx); ok {
		tracker.ResponseWriteTracker.TrackDuration(d)
	}
}

func AuditAnnotationsFromLatencyTrackers(ctx context.Context) map[string]string {
	const (
		transformLatencyKey         = "apiserver.latency.k8s.io/transform-response-object"
		storageLatencyKey           = "apiserver.latency.k8s.io/etcd"
		serializationLatencyKey     = "apiserver.latency.k8s.io/serialize-response-object"
		responseWriteLatencyKey     = "apiserver.latency.k8s.io/response-write"
		mutatingWebhookLatencyKey   = "apiserver.latency.k8s.io/mutating-webhook"
		validatingWebhookLatencyKey = "apiserver.latency.k8s.io/validating-webhook"
	)

	tracker, ok := LatencyTrackersFrom(ctx)
	if !ok {
		return nil
	}

	annotations := map[string]string{}
	if latency := tracker.TransformTracker.GetLatency(); latency != 0 {
		annotations[transformLatencyKey] = latency.String()
	}
	if latency := tracker.StorageTracker.GetLatency(); latency != 0 {
		annotations[storageLatencyKey] = latency.String()
	}
	if latency := tracker.SerializationTracker.GetLatency(); latency != 0 {
		annotations[serializationLatencyKey] = latency.String()
	}
	if latency := tracker.ResponseWriteTracker.GetLatency(); latency != 0 {
		annotations[responseWriteLatencyKey] = latency.String()
	}
	if latency := tracker.MutatingWebhookTracker.GetLatency(); latency != 0 {
		annotations[mutatingWebhookLatencyKey] = latency.String()
	}
	if latency := tracker.ValidatingWebhookTracker.GetLatency(); latency != 0 {
		annotations[validatingWebhookLatencyKey] = latency.String()
	}

	return annotations
}
