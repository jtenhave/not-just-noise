package dispatcher

import (
	"context"
	"log"
)

type logDispatcher struct {}

func NewLogDispatcher() *logDispatcher {
	return &logDispatcher{}
}

func (dispatcher *logDispatcher) Dispatch(ctx context.Context, destination string, payload string) error {
	log.Printf("dispatcher.logDispatcher.Dispatch: dispatching with destination: %s\n payload: %s\n", destination, payload)
	return nil
}