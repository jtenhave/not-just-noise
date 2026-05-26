package transactionaljobhandlers

import (
	"context"

	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type MessagePublisher interface {
	PublishMessage(ctx context.Context, destinat string, message string) error
}

type notifyHandler struct {
	messagePublisher MessagePublisher
}

func NewNotifyHandler(messagePublisher MessagePublisher) *notifyHandler {
	return &notifyHandler{
		messagePublisher: messagePublisher,
	}
}

func (handler *notifyHandler) Handle(ctx context.Context, destination string, payload string) error {
	err := handler.messagePublisher.PublishMessage(ctx, destination, payload)
	if err != nil {
		return njnerror.Wrapf("notifyhandler.Handle: failed to send notification: %w", err)
	}
	return nil
}

