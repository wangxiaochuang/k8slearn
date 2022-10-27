package warning

import (
	"context"
)

type key int

const (
	warningRecorderKey key = iota
)

type Recorder interface {
	AddWarning(agent, text string)
}

func WithWarningRecorder(ctx context.Context, recorder Recorder) context.Context {
	return context.WithValue(ctx, warningRecorderKey, recorder)
}
func warningRecorderFrom(ctx context.Context) (Recorder, bool) {
	recorder, ok := ctx.Value(warningRecorderKey).(Recorder)
	return recorder, ok
}

func AddWarning(ctx context.Context, agent string, text string) {
	recorder, ok := warningRecorderFrom(ctx)
	if !ok {
		return
	}
	recorder.AddWarning(agent, text)
}
