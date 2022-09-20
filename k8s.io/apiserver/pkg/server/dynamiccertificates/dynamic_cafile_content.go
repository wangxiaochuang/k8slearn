package dynamiccertificates

import (
	"context"
	"time"
)

var FileRefreshDuration = 1 * time.Minute

type ControllerRunner interface {
	RunOnce(ctx context.Context) error
	Run(ctx context.Context, workers int)
}
