package dispatcher

import (
	"context"

	"github.com/jtenhave/not-just-noise/lib/log"
)

type logDispatcher struct {}

func NewLogDispatcher() *logDispatcher {
	return &logDispatcher{}
}

func (dispatcher *logDispatcher) Dispatch(ctx context.Context, destination string, payload string) error {
	log.Logger(ctx).Info("dispatching log message", "payload", payload)
	return nil
}