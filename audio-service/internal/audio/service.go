package audio

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	audioContract "github.com/jtenhave/not-just-noise/contracts/audio"
	"github.com/jtenhave/not-just-noise/contracts/dispatch"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type TxManager interface {
	WithinTx(ctx context.Context, transaction func(context.Context) error) error
}

type AudioRepo interface {
	GetAudio(ctx context.Context, id string) (Audio, error)
	CreateAudio(ctx context.Context, audio Audio) error
	UpdateAudio(ctx context.Context, audio Audio) error
	DeleteAudio(ctx context.Context, id string) (int64, error)
}

type DispatchClient interface {
	Dispatch(ctx context.Context, dispatch dispatch.Dispatch) error
}

type audioService struct {
	txManager           TxManager
	audioRepo           AudioRepo
	dispatchClient      DispatchClient
	dispatchDestination string
}

// NewAudioService creates a new audioService using the given audioRepo and audioNotifier.
func NewAudioService(txManager TxManager, audioRepo AudioRepo, dispatchClient DispatchClient, dispatchDestination string) audioService {
	return audioService{
		txManager:           txManager,
		audioRepo:           audioRepo,
		dispatchClient:      dispatchClient,
		dispatchDestination: dispatchDestination,
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

	err := audioService.txManager.WithinTx(ctx, func(ctx context.Context) error {
		err := audioService.audioRepo.CreateAudio(ctx, audio)
		if err != nil {
			return njnerror.Wrapf("audioservice.CreateAudio: failed to create audio: %w", err)
		}

		err = audioService.raiseAudioChangedEvent(ctx, audioContract.AudioChangedEventTypeCreated, audio)
		if err != nil {
			return njnerror.Wrapf("audioservice.CreateAudio: failed to raise audio created event: %w", err)
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

	err = audioService.txManager.WithinTx(ctx, func(ctx context.Context) error {
		err = audioService.audioRepo.UpdateAudio(ctx, audioToUpdate)
		if err != nil {
			return njnerror.Wrapf("audioservice.UpdateAudio: failed to update audio: %w", err)
		}

		audioToUpdate.Version++

		err = audioService.raiseAudioChangedEvent(ctx, audioContract.AudioChangedEventTypeUpdated, audioToUpdate)
		if err != nil {
			return njnerror.Wrapf("audioservice.UpdateAudio: failed to raise audio updated event: %w", err)
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

	err = audioService.txManager.WithinTx(ctx, func(ctx context.Context) error {
		version, err := audioService.audioRepo.DeleteAudio(ctx, id)
		if err != nil {
			return njnerror.Wrapf("audioservice.DeleteAudio: failed to delete audio: %w", err)
		}
		
		audioToDelete.Version = version

		err = audioService.raiseAudioChangedEvent(ctx, audioContract.AudioChangedEventTypeDeleted, audioToDelete)
		if err != nil {
			return njnerror.Wrapf("audioservice.DeleteAudio: failed to raise audio deleted event: %w", err)
		}
		return nil
	})

	if err != nil {
		return njnerror.Wrapf("audioservice.DeleteAudio: delete audio transaction failed: %w", err)
	}

	return nil
}

// raiseAudioChangedEvent raises a new audio changed event for the given audio and event type. Returns the first error encountered.
func (audioService *audioService) raiseAudioChangedEvent(ctx context.Context, eventType audioContract.AudioChangedEventType, audio Audio) error {
	audioChangedEvent := audioContract.AudioChangedEvent{
		ID:        audio.ID,
		Title:     audio.Title,
		FileURL:   audio.FileURL,
		Version:   audio.Version,
		Status:    audio.Status,
		EventType: eventType,
	}

	payloadJSON, err := json.Marshal(audioChangedEvent)
	if err != nil {
		return njnerror.Wrapf("audioservice.raiseAudioChangedEvent: failed to create audio changed event payload: %w", err)
	}

	dispatch := dispatch.Dispatch{
		ID:               uuid.New().String(),
		CallbackType:     dispatch.CallbackTypeNotify,
		CallbackResource: audioService.dispatchDestination,
		Payload:          string(payloadJSON),
	}

	err = audioService.dispatchClient.Dispatch(ctx, dispatch)
	if err != nil {
		return njnerror.Wrapf("audioservice.raiseAudioChangedEvent: failed to dispatch audio changed event: %w", err)
	}

	return nil
}
