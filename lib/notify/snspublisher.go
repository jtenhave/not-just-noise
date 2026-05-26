package notify

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type snspublisher struct {
	client   *sns.Client
}

// NewSNSPublisher creates a new SNSPublisher using the given config.
func NewSNSPublisher(config aws.Config) *snspublisher {
	return &snspublisher{
		client:   sns.NewFromConfig(config),
	}
}

// PublishMessage sends a message to the given topicArn. Returns the first error encountered.
func (p *snspublisher) PublishMessage(ctx context.Context, topicArn string, message string) error {
	publishInput := &sns.PublishInput{
		TopicArn: aws.String(topicArn),
		Message:  aws.String(message),
	}

	_, err := p.client.Publish(ctx, publishInput)
	if err != nil {
		return fmt.Errorf("libsnspublisher.PublishMessage: failed to publish message: %w", err)
	}

	return nil
}
