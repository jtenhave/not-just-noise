package sqs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/jtenhave/not-just-noise/contracts/queue"
)

type sqsclient struct {
	client *sqs.Client
}

func NewSQSClient(client *sqs.Client) *sqsclient {
	return &sqsclient{
		client: client,
	}
}

// SendMessage sends a message to the given queueUrl. Returns the first error encountered.
func (r *sqsclient) SendMessage(ctx context.Context, queueUrl string, message queue.Message) error {
	_, err := r.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueUrl),
		MessageBody: message.Body,
	})

	if err != nil {
		return fmt.Errorf("integrations.sqs.sqsclient.SendMessage: failed to send message: %w", err)
	}

	return nil
}

// ReceiveMessages receives messages from the given queueUrl. Returns the messages and the first error encountered.
func (r *sqsclient) ReceiveMessages(ctx context.Context, queueUrl string, limit int32, visibilityTimeout int32) ([]queue.Message, error) {
	response, err := r.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueUrl),
		MaxNumberOfMessages: limit,
		VisibilityTimeout:   visibilityTimeout,
	})

	if err != nil {
		return nil, fmt.Errorf("integrations.sqs.sqsclient.ReceiveMessages: failed to receive messages: %w", err)
	}

	if len(response.Messages) == 0 {
		return nil, nil
	}

	messages := make([]queue.Message, len(response.Messages))
	for i, message := range response.Messages {
		messages[i] = queue.Message{
			Body: message.Body,
			Delete: func(ctx context.Context) error {
				_, err := r.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
					QueueUrl:      aws.String(queueUrl),
					ReceiptHandle: message.ReceiptHandle,
				})

				if err != nil {
					return fmt.Errorf("integrations.sqs.sqsclient.ReceiveMessages: failed to delete message: %w", err)
				}

				return nil
			},
		}
	}

	return messages, nil
}
