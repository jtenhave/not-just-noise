package dispatcher

import (
	"context"

	"github.com/jtenhave/not-just-noise/contracts/queue"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type QueueClient interface {
	SendMessage(ctx context.Context, queueUrl string, message queue.Message) error
}

type queueDispatcher struct {
	queueClient QueueClient
}

func NewQueueDispatcher(queueClient QueueClient) *queueDispatcher {
	return &queueDispatcher{
		queueClient: queueClient,
	}
}

func (dispatcher *queueDispatcher) Dispatch(ctx context.Context, destination string, payload string) error {
	err := dispatcher.queueClient.SendMessage(ctx, destination, queue.Message{
		Body: &payload,
	})

	if err != nil {
		return njnerror.Wrapf("queuedispatcher.Dispatch: failed to send queue message: %w", err)
	}

	return nil
}
