package adapters

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jtenhave/not-just-noise/audio-service/internal/config"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
	"github.com/jtenhave/not-just-noise/lib/transactionaljob"
)

const (
	publishClaimTimeoutSeconds = 300
	publishRetrySeconds        = 2
	publishRetryBackoff        = 1.5
	publishMaxAttempts         = 10
	publishCallbackType        = "notify"
)

type publishEventType string

const (
	PublishCreatedEvent publishEventType = "CREATED"
	PublishUpdatedEvent publishEventType = "UPDATED"
	PublishDeletedEvent publishEventType = "DELETED"
)

type AudioPublishPayload struct {
	ID      string
	Title   *string
	FileURL *string
	Version int64
	Status  string
}

type TransactionalJobClient interface {
	CreateTransactionalJob(ctx context.Context, transactionalJob transactionaljob.TransactionalJob) error
}

type publishAdapter struct {
	snsConfig              config.SNSConfig
	transactionalJobClient TransactionalJobClient
}

// NewPublishAdapter creates a new publish adapter using the given snsConfig and transactionalJobClient.
func NewPublishAdapter(snsConfig config.SNSConfig, transactionalJobClient TransactionalJobClient) *publishAdapter {
	return &publishAdapter{
		snsConfig:              snsConfig,
		transactionalJobClient: transactionalJobClient,
	}
}

// PublishCreated publishes a created event for the given audio. Returns the first error encountered.
func (adapter *publishAdapter) PublishCreated(ctx context.Context, audio AudioPublishPayload) error {
	err := adapter.createTransactionalJob(ctx, audio, PublishCreatedEvent)
	if err != nil {
		return njnerror.Wrapf("audioservice.PublishCreated: failed to publish created: %w", err)
	}

	return nil
}

// PublishUpdated publishes an updated event for the given audio. Returns the first error encountered.
func (adapter *publishAdapter) PublishUpdated(ctx context.Context, audio AudioPublishPayload) error {
	err := adapter.createTransactionalJob(ctx, audio, PublishUpdatedEvent)
	if err != nil {
		return njnerror.Wrapf("audioservice.PublishUpdated: failed to publish updated: %w", err)
	}

	return nil
}

// PublishDeleted publishes a deleted event for the given audio. Returns the first error encountered.
func (adapter *publishAdapter) PublishDeleted(ctx context.Context, audio AudioPublishPayload) error {
	err := adapter.createTransactionalJob(ctx, audio, PublishDeletedEvent)
	if err != nil {
		return njnerror.Wrapf("audioservice.PublishDeleted: failed to publish deleted: %w", err)
	}

	return nil
}

// createTransactionalJob creates a new transactional job for the given audio and event type. Returns the first error encountered.
func (adapter *publishAdapter) createTransactionalJob(ctx context.Context, audio AudioPublishPayload, eventType publishEventType) error {
	payload, err := adapter.createTransactionalJobPayload(eventType, audio)
	if err != nil {
		return njnerror.Wrapf("audioservice.createTransactionalJob: failed to create publish payload: %w", err)
	}

	transactionalJob := transactionaljob.TransactionalJob{
		ID:               uuid.New().String(),
		CallbackType:     publishCallbackType,
		CallbackResource: adapter.snsConfig.TopicArn,
		Payload:          payload,
		ClaimTimeout:     publishClaimTimeoutSeconds,
		RetrySeconds:     publishRetrySeconds,
		RetryBackoff:     publishRetryBackoff,
		MaxAttempts:      publishMaxAttempts,
	}

	err = adapter.transactionalJobClient.CreateTransactionalJob(ctx, transactionalJob)
	if err != nil {
		return njnerror.Wrapf("audioservice.createTransactionalJob: failed to create transactional job: %w", err)
	}

	return nil
}

// createTransactionalJobPayload creates a new transactional job payload for the given audio and event type. Returns the first error encountered.
func (adapter *publishAdapter) createTransactionalJobPayload(eventType publishEventType, audio AudioPublishPayload) (string, error) {
	payload := map[string]any{
		"id":         audio.ID,
		"title":      *audio.Title,
		"file_url":   *audio.FileURL,
		"version":    audio.Version,
		"status":     audio.Status,
		"event_type": eventType,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", njnerror.Wrapf("audioservice.createTransactionalJobPayload: failed to create publish payload: %w", err)
	}

	return string(payloadJSON), nil
}
