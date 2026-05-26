package transactionaljobhandlers

import (
	"context"
	"log"
)

type logHandler struct {}

func NewLogHandler() *logHandler {
	return &logHandler{}
}

func (handler *logHandler) Handle(ctx context.Context, destination string, payload string) error {
	log.Printf("transactionaljobhandlers.logHandler.Handle: handling transactional job with destination: %s\n payload: %s\n", destination, payload)
	return nil
}