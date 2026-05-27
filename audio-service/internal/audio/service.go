package audio

import (
	"context"

	"github.com/google/uuid"
	"github.com/jtenhave/not-just-noise/audio-service/internal/adapters"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type TransactionManager interface {
	WithinTx(ctx context.Context, transaction func(context.Context) error) error
}

type AudioRepo interface {
	GetAudio(ctx context.Context, id string) (Audio, error)
	CreateAudio(ctx context.Context, audio Audio) error
	UpdateAudio(ctx context.Context, audio Audio) error
	DeleteAudio(ctx context.Context, id string) error
}

type AudioPublishAdpater interface {
	PublishCreated(ctx context.Context, audio adapters.AudioPublishPayload) error
	PublishUpdated(ctx context.Context, audio adapters.AudioPublishPayload) error
	PublishDeleted(ctx context.Context, audio adapters.AudioPublishPayload) error
}

type audioService struct {
	transactionManager  TransactionManager
	audioRepo           AudioRepo
	audioPublishAdapter AudioPublishAdpater
}

// NewAudioService creates a new audioService using the given audioRepo and audioNotifier.
func NewAudioService(transactionManager TransactionManager, audioRepo AudioRepo, audioPublishAdapter AudioPublishAdpater) audioService {
	return audioService{
		transactionManager:  transactionManager,
		audioRepo:           audioRepo,
		audioPublishAdapter: audioPublishAdapter,
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

	audioPublishPayload := audioService.createAudioPublishPayload(audio)
	err := audioService.transactionManager.WithinTx(ctx, func(ctx context.Context) error {
		err := audioService.audioRepo.CreateAudio(ctx, audio)
		if err != nil {
			return njnerror.Wrapf("audioservice.CreateAudio: failed to create audio: %w", err)
		}

		err = audioService.audioPublishAdapter.PublishCreated(ctx, audioPublishPayload)
		if err != nil {
			return njnerror.Wrapf("audioservice.CreateAudio: failed to publish created: %w", err)
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
	audioToUpdate, err := audioService.audioRepo.GetAudio(ctx, audio.ID)
	if err != nil {
		return njnerror.Wrapf("audioservice.UpdateAudio: failed to get audio: %w", err)
	}

	if audio.Title != nil {
		audioToUpdate.Title = audio.Title
	}

	if audio.FileURL != nil {
		audioToUpdate.FileURL = audio.FileURL
	}

	audioToUpdate.Version++

	audioPublishPayload := audioService.createAudioPublishPayload(audioToUpdate)
	err = audioService.transactionManager.WithinTx(ctx, func(ctx context.Context) error {
		err = audioService.audioRepo.UpdateAudio(ctx, audioToUpdate)
		if err != nil {
			return njnerror.Wrapf("audioservice.UpdateAudio: failed to update audio: %w", err)
		}

		err = audioService.audioPublishAdapter.PublishUpdated(ctx, audioPublishPayload)
		if err != nil {
			return njnerror.Wrapf("audioservice.UpdateAudio: failed to publish updated: %w", err)
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
	audioToDelete, err := audioService.audioRepo.GetAudio(ctx, id)
	if err != nil {
		return njnerror.Wrapf("audioservice.DeleteAudio: failed to get audio: %w", err)
	}

	audioToDelete.Version++
	audioToDelete.Status = "deleted"

	audioPublishPayload := audioService.createAudioPublishPayload(audioToDelete)
	err = audioService.transactionManager.WithinTx(ctx, func(ctx context.Context) error {
		err = audioService.audioRepo.DeleteAudio(ctx, id)
		if err != nil {
			return njnerror.Wrapf("audioservice.DeleteAudio: failed to delete audio: %w", err)
		}

		err = audioService.audioPublishAdapter.PublishDeleted(ctx, audioPublishPayload)
		if err != nil {
			return njnerror.Wrapf("audioservice.DeleteAudio: failed to publish deleted: %w", err)
		}
		return nil
	})

	if err != nil {
		return njnerror.Wrapf("audioservice.DeleteAudio: delete audio transaction failed: %w", err)
	}

	return nil
}

// createAudioPublishPayload creates a new audio publish payload for the given audio.
func (audioService audioService) createAudioPublishPayload(audio Audio) adapters.AudioPublishPayload {
	status := "active"
	if audio.Status != "" {
		status = audio.Status
	}

	return adapters.AudioPublishPayload{
		ID:      audio.ID,
		Title:   audio.Title,
		FileURL: audio.FileURL,
		Version: audio.Version,
		Status:  status,
	}
}
