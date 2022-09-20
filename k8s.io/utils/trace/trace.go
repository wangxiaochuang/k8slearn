package trace

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"time"

	"k8s.io/klog/v2"
)

var klogV = func(lvl klog.Level) bool {
	return klog.V(lvl).Enabled()
}

type Field struct {
	Key   string
	Value interface{}
}

func (f Field) format() string {
	return fmt.Sprintf("%s:%v", f.Key, f.Value)
}

func writeFields(b *bytes.Buffer, l []Field) {
	for i, f := range l {
		b.WriteString(f.format())
		if i < len(l)-1 {
			b.WriteString(",")
		}
	}
}

func writeTraceItemSummary(b *bytes.Buffer, msg string, totalTime time.Duration, startTime time.Time, fields []Field) {
	b.WriteString(fmt.Sprintf("%q ", msg))
	if len(fields) > 0 {
		writeFields(b, fields)
		b.WriteString(" ")
	}

	b.WriteString(fmt.Sprintf("%vms (%v)", durationToMilliseconds(totalTime), startTime.Format("15:04:05.000")))
}

func durationToMilliseconds(timeDuration time.Duration) int64 {
	return timeDuration.Nanoseconds() / 1e6
}

type traceItem interface {
	time() time.Time
	writeItem(b *bytes.Buffer, formatter string, startTime time.Time, stepThreshold *time.Duration)
}

type traceStep struct {
	stepTime time.Time
	msg      string
	fields   []Field
}

func (s traceStep) time() time.Time {
	return s.stepTime
}

func (s traceStep) writeItem(b *bytes.Buffer, formatter string, startTime time.Time, stepThreshold *time.Duration) {
	stepDuration := s.stepTime.Sub(startTime)
	if stepThreshold == nil || *stepThreshold == 0 || stepDuration >= *stepThreshold || klogV(4) {
		b.WriteString(fmt.Sprintf("%s---", formatter))
		writeTraceItemSummary(b, s.msg, stepDuration, s.stepTime, s.fields)
	}
}

type Trace struct {
	name        string
	fields      []Field
	threshold   *time.Duration
	startTime   time.Time
	endTime     *time.Time
	traceItems  []traceItem
	parentTrace *Trace
}

func (t *Trace) time() time.Time {
	if t.endTime != nil {
		return *t.endTime
	}
	return t.startTime // if the trace is incomplete, don't assume an end time
}

func (t *Trace) writeItem(b *bytes.Buffer, formatter string, startTime time.Time, stepThreshold *time.Duration) {
	if t.durationIsWithinThreshold() || klogV(4) {
		b.WriteString(fmt.Sprintf("%v[", formatter))
		writeTraceItemSummary(b, t.name, t.TotalTime(), t.startTime, t.fields)
		if st := t.calculateStepThreshold(); st != nil {
			stepThreshold = st
		}
		t.writeTraceSteps(b, formatter+" ", stepThreshold)
		b.WriteString("]")
		return
	}
	// If the trace should not be written, still check for nested traces that should be written
	for _, s := range t.traceItems {
		if nestedTrace, ok := s.(*Trace); ok {
			nestedTrace.writeItem(b, formatter, startTime, stepThreshold)
		}
	}
}

func New(name string, fields ...Field) *Trace {
	return &Trace{name: name, startTime: time.Now(), fields: fields}
}

func (t *Trace) Step(msg string, fields ...Field) {
	if t.traceItems == nil {
		// traces almost always have less than 6 steps, do this to avoid more than a single allocation
		t.traceItems = make([]traceItem, 0, 6)
	}
	t.traceItems = append(t.traceItems, traceStep{stepTime: time.Now(), msg: msg, fields: fields})
}

func (t *Trace) Nest(msg string, fields ...Field) *Trace {
	newTrace := New(msg, fields...)
	if t != nil {
		newTrace.parentTrace = t
		t.traceItems = append(t.traceItems, newTrace)
	}
	return newTrace
}

func (t *Trace) Log() {
	endTime := time.Now()
	t.endTime = &endTime
	// an explicit logging request should dump all the steps out at the higher level
	if t.parentTrace == nil { // We don't start logging until Log or LogIfLong is called on the root trace
		t.logTrace()
	}
}

func (t *Trace) LogIfLong(threshold time.Duration) {
	t.threshold = &threshold
	t.Log()
}

func (t *Trace) logTrace() {
	if t.durationIsWithinThreshold() {
		var buffer bytes.Buffer
		traceNum := rand.Int31()

		totalTime := t.endTime.Sub(t.startTime)
		buffer.WriteString(fmt.Sprintf("Trace[%d]: %q ", traceNum, t.name))
		if len(t.fields) > 0 {
			writeFields(&buffer, t.fields)
			buffer.WriteString(" ")
		}

		// if any step took more than it's share of the total allowed time, it deserves a higher log level
		buffer.WriteString(fmt.Sprintf("(%v) (total time: %vms):", t.startTime.Format("02-Jan-2006 15:04:05.000"), totalTime.Milliseconds()))
		stepThreshold := t.calculateStepThreshold()
		t.writeTraceSteps(&buffer, fmt.Sprintf("\nTrace[%d]: ", traceNum), stepThreshold)
		buffer.WriteString(fmt.Sprintf("\nTrace[%d]: [%v] [%v] END\n", traceNum, t.endTime.Sub(t.startTime), totalTime))

		klog.Info(buffer.String())
		return
	}

	// If the trace should not be logged, still check if nested traces should be logged
	for _, s := range t.traceItems {
		if nestedTrace, ok := s.(*Trace); ok {
			nestedTrace.logTrace()
		}
	}
}

func (t *Trace) writeTraceSteps(b *bytes.Buffer, formatter string, stepThreshold *time.Duration) {
	lastStepTime := t.startTime
	for _, stepOrTrace := range t.traceItems {
		stepOrTrace.writeItem(b, formatter, lastStepTime, stepThreshold)
		lastStepTime = stepOrTrace.time()
	}
}

func (t *Trace) durationIsWithinThreshold() bool {
	if t.endTime == nil { // we don't assume incomplete traces meet the threshold
		return false
	}
	return t.threshold == nil || *t.threshold == 0 || t.endTime.Sub(t.startTime) >= *t.threshold
}

func (t *Trace) TotalTime() time.Duration {
	return time.Since(t.startTime)
}

func (t *Trace) calculateStepThreshold() *time.Duration {
	if t.threshold == nil {
		return nil
	}
	lenTrace := len(t.traceItems) + 1
	traceThreshold := *t.threshold
	for _, s := range t.traceItems {
		nestedTrace, ok := s.(*Trace)
		if ok && nestedTrace.threshold != nil {
			traceThreshold = traceThreshold - *nestedTrace.threshold
			lenTrace--
		}
	}

	// the limit threshold is used when the threshold(
	//remaining after subtracting that of the child trace) is getting very close to zero to prevent unnecessary logging
	limitThreshold := *t.threshold / 4
	if traceThreshold < limitThreshold {
		traceThreshold = limitThreshold
		lenTrace = len(t.traceItems) + 1
	}

	stepThreshold := traceThreshold / time.Duration(lenTrace)
	return &stepThreshold
}

type ContextTraceKey struct{}

func FromContext(ctx context.Context) *Trace {
	if v, ok := ctx.Value(ContextTraceKey{}).(*Trace); ok {
		return v
	}
	return nil
}

func ContextWithTrace(ctx context.Context, trace *Trace) context.Context {
	return context.WithValue(ctx, ContextTraceKey{}, trace)
}
