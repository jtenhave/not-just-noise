package notify

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type snspublisher struct {
	client   *sns.Client
	topicArn string
}

// NewSNSPublisher creates a new SNSPublisher using the given config and topicArn.
func NewSNSPublisher(config aws.Config, topicArn string) *snspublisher {
	return &snspublisher{
		client:   sns.NewFromConfig(config),
		topicArn: topicArn,
	}
}

// PublishMessage sends a message to the SNS topic. Returns the first error encountered.
func (p *snspublisher) PublishMessage(ctx context.Context, message string) error {
	publishInput := &sns.PublishInput{
		TopicArn: aws.String(p.topicArn),
		Message:  aws.String(message),
	}

	_, err := p.client.Publish(ctx, publishInput)
	if err != nil {
		return fmt.Errorf("libsnspublisher.PublishMessage: failed to publish message: %w", err)
	}

	return nil
}
