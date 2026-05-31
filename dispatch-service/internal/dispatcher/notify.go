package dispatcher

import (
	"context"

	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type Notifier interface {
	PublishMessage(ctx context.Context, destination string, message string) error
}

type notifyDispatcher struct {
	notifier Notifier
}

func NewNotifyDispatcher(notifier Notifier) *notifyDispatcher {
	return &notifyDispatcher{
		notifier: notifier,
	}
}

func (dispatcher *notifyDispatcher) Dispatch(ctx context.Context, destination string, payload string) error {
	err := dispatcher.notifier.PublishMessage(ctx, destination, payload)
	if err != nil {
		return njnerror.Wrapf("notifydispatcher.Dispatch: failed to send notification: %w", err)
	}
	return nil
}
