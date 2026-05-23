package audio

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jtenhave/not-just-noise/audio-service/internal/transactionaloutbox"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type TransactionManager interface {
	WithinTx(ctx context.Context, transaction func(context.Context) error) error
}

type AudioRepo interface {
	GetAudio(ctx context.Context, id string) (Audio, error)
	CreateAudio(ctx context.Context, audio Audio) error
	UpdateAudio(ctx context.Context, audio Audio, version int64) error
	DeleteAudio(ctx context.Context, id string) error
}

type TransactionalOutboxRepo interface {
	CreateTransactionalOutboxRecord(ctx context.Context, transactionalOutboxRecord transactionaloutbox.TransactionalOutboxRecord) error
}

type audioNotifierEventType string

const (
	AudioCreatedEvent audioNotifierEventType = "CREATED"
	AudioUpdatedEvent audioNotifierEventType = "UPDATED"
	AudioDeletedEvent audioNotifierEventType = "DELETED"
)

type audioService struct {
	transactionManager      TransactionManager
	audioRepo               AudioRepo
	transactionalOutboxRepo TransactionalOutboxRepo
}

// NewAudioService creates a new audioService using the given audioRepo and audioNotifier.
func NewAudioService(transactionManager TransactionManager, audioRepo AudioRepo, transactionalOutboxRepo TransactionalOutboxRepo) audioService {
	return audioService{
		transactionManager:      transactionManager,
		audioRepo:               audioRepo,
		transactionalOutboxRepo: transactionalOutboxRepo,
	}
}

// GetAudio gets an audio record using the given id. Returns the audio record and the first error encountered.
func (audioService audioService) GetAudio(context context.Context, id string) (Audio, error) {
	audio, err := audioService.audioRepo.GetAudio(context, id)
	if err != nil {
		return Audio{}, njnerror.Wrapf("audioservice.GetAudio: failed to get audio: %w", err)
	}

	return audio, nil
}

// CreateAudio creates a new audio record using the given audio. Returns the id of the created audio and the first error encountered.
func (audioService audioService) CreateAudio(ctx context.Context, audio Audio) (string, error) {
	audio.ID = uuid.New().String()

	transactionalOutboxRecord, err := audioService.createTransactionalOutboxRecord(AudioCreatedEvent, audio)
	if err != nil {
		return "", njnerror.Wrapf("audioservice.CreateAudio: failed to create transactional outbox record: %w", err)
	}

	err = audioService.transactionManager.WithinTx(ctx, func(ctx context.Context) error {
		err = audioService.audioRepo.CreateAudio(ctx, audio)
		if err != nil {
			return njnerror.Wrapf("audioservice.CreateAudio: failed to create audio: %w", err)
		}

		err = audioService.transactionalOutboxRepo.CreateTransactionalOutboxRecord(ctx, transactionalOutboxRecord)
		if err != nil {
			return njnerror.Wrapf("audioservice.CreateAudio: failed to create transactional outbox record: %w", err)
		}
		return nil
	})

	if err != nil {
		return "", njnerror.Wrapf("audioservice.CreateAudio: create audio transaction failed: %w", err)
	}

	return audio.ID, nil
}

// UpdateAudio updates an audio record using the given audio. Returns the first error encountered.
func (audioService audioService) UpdateAudio(ctx context.Context, audio Audio) error {
	storedAudio, err := audioService.audioRepo.GetAudio(ctx, audio.ID)
	if err != nil {
		return njnerror.Wrapf("audioservice.UpdateAudio: failed to get audio: %w", err)
	}

	storedAudio.Version++
	if audio.Title != nil {
		storedAudio.Title = audio.Title
	}

	if audio.FileURL != nil {
		storedAudio.FileURL = audio.FileURL
	}

	transactionalOutboxRecord, err := audioService.createTransactionalOutboxRecord(AudioUpdatedEvent, storedAudio)
	if err != nil {
		return njnerror.Wrapf("audioservice.UpdateAudio: failed to create transactional outbox record: %w", err)
	}

	err = audioService.transactionManager.WithinTx(ctx, func(ctx context.Context) error {
		err = audioService.audioRepo.UpdateAudio(ctx, audio, storedAudio.Version)
		if err != nil {
			return njnerror.Wrapf("audioservice.UpdateAudio: failed to update audio: %w", err)
		}

		err = audioService.transactionalOutboxRepo.CreateTransactionalOutboxRecord(ctx, transactionalOutboxRecord)
		if err != nil {
			return njnerror.Wrapf("audioservice.UpdateAudio: failed to create transactional outbox record: %w", err)
		}
		return nil
	})

	if err != nil {
		return njnerror.Wrapf("audioservice.UpdateAudio: update audio transaction failed: %w", err)
	}

	return nil
}

// DeleteAudio deletes an audio record using the given id. Returns the first error encountered.
func (audioService audioService) DeleteAudio(ctx context.Context, id string) error {
	audio, err := audioService.audioRepo.GetAudio(ctx, id)
	if err != nil {
		return njnerror.Wrapf("audioservice.DeleteAudio: failed to get audio: %w", err)
	}

	audio.Version++
	audio.Status = "deleted"

	transactionalOutboxRecord, err := audioService.createTransactionalOutboxRecord(AudioDeletedEvent, audio)
	if err != nil {
		return njnerror.Wrapf("audioservice.DeleteAudio: failed to create transactional outbox record: %w", err)
	}

	err = audioService.transactionManager.WithinTx(ctx, func(ctx context.Context) error {
		err = audioService.audioRepo.DeleteAudio(ctx, id)
		if err != nil {
			return njnerror.Wrapf("audioservice.DeleteAudio: failed to delete audio: %w", err)
		}

		err = audioService.transactionalOutboxRepo.CreateTransactionalOutboxRecord(ctx, transactionalOutboxRecord)
		if err != nil {
			return njnerror.Wrapf("audioservice.DeleteAudio: failed to create transactional outbox record: %w", err)
		}
		return nil
	})

	if err != nil {
		return njnerror.Wrapf("audioservice.DeleteAudio: delete audio transaction failed: %w", err)
	}

	return nil
}

func (audioService audioService) createTransactionalOutboxRecord(eventType audioNotifierEventType, audio Audio) (transactionaloutbox.TransactionalOutboxRecord, error) {
	payload := map[string]any{
		"event_type": eventType,
		"audio_id":   audio.ID,
		"title":      *audio.Title,
		"file_url":   *audio.FileURL,
		"version":    audio.Version,
		"status":     audio.Status,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return transactionaloutbox.TransactionalOutboxRecord{}, njnerror.Wrapf("audioservice.CreateTransactionalOutboxRecord: failed to create notify payload: %w", err)
	}

	return transactionaloutbox.TransactionalOutboxRecord{
		ID:      uuid.New().String(),
		Payload: string(payloadJSON),
	}, nil
}
