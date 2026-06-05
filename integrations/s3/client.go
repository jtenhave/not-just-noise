package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3client struct {
	client *s3.Client
	bucket string
}

func NewS3Client(client *s3.Client, bucket string) *s3client {
	return &s3client{
		client: client,
		bucket: bucket,
	}
}

func (c *s3client) UploadFile(ctx context.Context, key string, body io.Reader, metadata map[string]string) error {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:   aws.String(c.bucket),
		Key:      aws.String(key),
		Body:     body,
		Metadata: metadata,
	})

	if err != nil {
		return fmt.Errorf("integrations.s3.s3client.UploadFile: failed to upload file: %w", err)
	}

	return err
}
