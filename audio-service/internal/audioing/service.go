package audioing

import (
	"context"

	"github.com/google/uuid"
	"github.com/jtenhave/not-just-noise/audio-service/internal/audio"
	"github.com/jtenhave/not-just-noise/audio-service/internal/infrastructure/repo"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
	"github.com/jtenhave/not-just-noise/lib/notify"
)

type audioService struct {
	connectionRepo       repo.ConnectionRepository
	audioNotifyFormatter AudioNotifyFormatter
}

type AudioNotifyFormatter interface {
	NotifyAudioChangedPayload(eventType audio.AudioNotifierEventType, audio audio.Audio) (string, error)
}

// NewAudioService creates a new audioService using the given audioRepo and audioNotifier.
func NewAudioService(connectionRepo repo.ConnectionRepository, audioNotifyFormatter AudioNotifyFormatter) audioService {
	return audioService{
		connectionRepo:       connectionRepo,
		audioNotifyFormatter: audioNotifyFormatter,
	}
}

// GetAudio gets an audio record using the given id. Returns the audio record and the first error encountered.
func (audioService audioService) GetAudio(context context.Context, id string) (audio.Audio, error) {
	a, err := audioService.connectionRepo.AudioRepo().GetAudio(context, id)
	if err != nil {
		return audio.Audio{}, njnerror.Wrapf("audioservice.GetAudio: failed to get audio: %w", err)
	}

	return a, nil
}

// CreateAudio creates a new audio record using the given audio. Returns the id of the created audio and the first error encountered.
func (audioService audioService) CreateAudio(ctx context.Context, a audio.Audio) (string, error) {
	a.ID = uuid.New().String()

	notifyPayload, err := audioService.audioNotifyFormatter.NotifyAudioChangedPayload(audio.AudioCreatedEvent, a)
	if err != nil {
		return "", njnerror.Wrapf("audioservice.CreateAudio: failed to create notify payload: %w", err)
	}

	tx, err := audioService.connectionRepo.BeginTx(ctx)
	if err != nil {
		return "", njnerror.Wrapf("audioservice.CreateAudio: failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	err = tx.AudioRepo().CreateAudio(ctx, a)
	if err != nil {
		return "", njnerror.Wrapf("audioservice.CreateAudio: failed to create audio: %w", err)
	}

	transactionalOutboxRecord := notify.TransactionalOutboxRecord{
		ID:      uuid.New().String(),
		Payload: notifyPayload,
	}

	err = tx.TransactionalOutboxRepo().CreateTransactionalOutboxRecord(ctx, transactionalOutboxRecord)
	if err != nil {
		return "", njnerror.Wrapf("audioservice.CreateAudio: failed to create transactional outbox record: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return "", njnerror.Wrapf("audioservice.CreateAudio: failed to commit transaction: %w", err)
	}

	return a.ID, nil
}

// UpdateAudio updates an audio record using the given audio. Returns the first error encountered.
func (audioService audioService) UpdateAudio(ctx context.Context, a audio.UpdateAudio) error {
	storedAudio, err := audioService.connectionRepo.AudioRepo().GetAudio(ctx, a.ID)
	if err != nil {
		return njnerror.Wrapf("audiorepo.UpdateAudio: failed to get audio: %w", err)
	}

	payloadAudio := storedAudio
	payloadAudio.Version++
	if a.Title != nil {
		payloadAudio.Title = *a.Title
	}

	if a.FileURL != nil {
		payloadAudio.FileURL = *a.FileURL
	}

	notifyPayload, err := audioService.audioNotifyFormatter.NotifyAudioChangedPayload(audio.AudioUpdatedEvent, payloadAudio)
	if err != nil {
		return njnerror.Wrapf("audioservice.UpdateAudio: failed to create notify payload: %w", err)
	}

	tx, err := audioService.connectionRepo.BeginTx(ctx)
	if err != nil {
		return njnerror.Wrapf("audioservice.CreateAudio: failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	err = tx.AudioRepo().UpdateAudio(ctx, a, storedAudio.Version)
	if err != nil {
		return njnerror.Wrapf("audioservice.CreateAudio: failed to create audio: %w", err)
	}

	transactionalOutboxRecord := notify.TransactionalOutboxRecord{
		ID:      uuid.New().String(),
		Payload: notifyPayload,
	}

	err = tx.TransactionalOutboxRepo().CreateTransactionalOutboxRecord(ctx, transactionalOutboxRecord)
	if err != nil {
		return njnerror.Wrapf("audioservice.CreateAudio: failed to create transactional outbox record: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return njnerror.Wrapf("audioservice.CreateAudio: failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteAudio deletes an audio record using the given id. Returns the first error encountered.
func (audioService audioService) DeleteAudio(ctx context.Context, id string) error {
	storedAudio, err := audioService.connectionRepo.AudioRepo().GetAudio(ctx, id)
	if err != nil {
		return njnerror.Wrapf("audiorepo.DeleteAudio: failed to get audio: %w", err)
	}

	payloadAudio := storedAudio
	payloadAudio.Version++
	payloadAudio.Status = "deleted"

	notifyPayload, err := audioService.audioNotifyFormatter.NotifyAudioChangedPayload(audio.AudioDeletedEvent, payloadAudio)
	if err != nil {
		return njnerror.Wrapf("audioservice.DeleteAudio: failed to create notify payload: %w", err)
	}

	tx, err := audioService.connectionRepo.BeginTx(ctx)
	if err != nil {
		return njnerror.Wrapf("audioservice.DeleteAudio: failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	err = tx.AudioRepo().DeleteAudio(ctx, id)
	if err != nil {
		return njnerror.Wrapf("audioservice.DeleteAudio: failed to delete audio: %w", err)
	}

	transactionalOutboxRecord := notify.TransactionalOutboxRecord{
		ID:      uuid.New().String(),
		Payload: notifyPayload,
	}

	err = tx.TransactionalOutboxRepo().CreateTransactionalOutboxRecord(ctx, transactionalOutboxRecord)
	if err != nil {
		return njnerror.Wrapf("audioservice.DeleteAudio: failed to create transactional outbox record: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return njnerror.Wrapf("audioservice.DeleteAudio: failed to commit transaction: %w", err)
	}

	return nil
}
