package upload

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jtenhave/not-just-noise/contracts/queue"
	"github.com/jtenhave/not-just-noise/lib/log"
	"github.com/jtenhave/not-just-noise/lib/utils"
)

const (
	visibilityTimeout = 300
)

const (
	defaultMaxAttempts  = 10
	defaultRetrySeconds = 2
	defaultRetryBackoff = 1.5
)

type WorkerConfig struct {
	MaxWorkers    int
	MaxBatchSize  int
	NoJobDelay    int
	JobBufferSize int
}

type MessageQueueClient interface {
	ReceiveMessages(ctx context.Context, queueUrl string, limit int32, visibilityTimeout int32) ([]queue.Message, error)
}

type uploadWorker struct {
	config             WorkerConfig
	messageQueueClient MessageQueueClient
	queueURL           string
	handler            func(ctx context.Context, message string) error
}

// NewUploadWorker creates a new uploadWorker using the given uploadService.
func NewUploadWorker(config WorkerConfig, messageQueueClient MessageQueueClient, queueURL string, handler func(ctx context.Context, message string) error) uploadWorker {
	return uploadWorker{
		config:             config,
		messageQueueClient: messageQueueClient,
		queueURL:           queueURL,
		handler:            handler,
	}
}

// Start starts the uploadWorker.
func (worker *uploadWorker) Start(ctx context.Context) {
	log.Logger(ctx).Info("starting upload worker")
	worker.work(ctx)
}

// work is the main loop for the dispatchWorker.
func (worker *uploadWorker) work(ctx context.Context) {
	sendChannel := make(chan queue.Message, worker.config.JobBufferSize)
	for i := 0; i < worker.config.MaxWorkers; i++ {
		go worker.runWorker(ctx, sendChannel)
	}

	defer close(sendChannel)

	for {
		messages, err := worker.messageQueueClient.ReceiveMessages(ctx, worker.queueURL, int32(worker.config.MaxBatchSize), visibilityTimeout)
		if err != nil {
			log.Logger(ctx).Error("failed to receive messages from queue", "queue_url", worker.queueURL, "error", err)

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(worker.config.NoJobDelay) * time.Second):
			}
			continue
		}

		for _, message := range messages {
			select {
			case <-ctx.Done():
				return

			case sendChannel <- message:
			}
		}

		if len(messages) < worker.config.MaxBatchSize {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(worker.config.NoJobDelay) * time.Second):
			}
		}
	}
}

// runWorker runs a new worker using the given sendChannel.
func (worker *uploadWorker) runWorker(ctx context.Context, sendChannel <-chan queue.Message) {
	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-sendChannel:
			if !ok {
				return
			}

			messageBody := struct {
				Message string `json:"message"`
			}{Message: ""}

			err := json.Unmarshal([]byte(*message.Body), &messageBody)
			if err != nil {
				log.Logger(ctx).Error("failed to unmarshal message body from queue", "queue_url", worker.queueURL, "error", err)
				continue
			}

			err = worker.handler(ctx, messageBody.Message)
			if err != nil {
				log.Logger(ctx).Error("failed to handle message from queue", "queue_url", worker.queueURL, "error", err)
				continue
			}

			// Best effort retry to delete the message. Worst case, the message is processed again.
			err = utils.Retry(ctx, func() error {
				return message.Delete(ctx)
			}, defaultMaxAttempts, defaultRetrySeconds, defaultRetryBackoff)

			if err != nil {
				log.Logger(ctx).Error("failed to delete message from queue", "queue_url", worker.queueURL, "error", err)
				continue
			}
		}
	}
}
