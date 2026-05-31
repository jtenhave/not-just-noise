package sns

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type snsClient struct {
	client   *sns.Client
}

// NewSNSClient creates a new SNSClient using the given client.
func NewSNSClient(client *sns.Client) *snsClient {
	return &snsClient{
		client: client,
	}
}

// PublishMessage sends a message to the given topicArn. Returns the first error encountered.
func (client *snsClient) PublishMessage(ctx context.Context, topicArn string, message string) error {
	publishInput := &sns.PublishInput{
		TopicArn: aws.String(topicArn),
		Message:  aws.String(message),
	}

	_, err := client.client.Publish(ctx, publishInput)
	if err != nil {
		return fmt.Errorf("integrations.sns.snsClient.PublishMessage: failed to publish message: %w", err)
	}

	return nil
}
