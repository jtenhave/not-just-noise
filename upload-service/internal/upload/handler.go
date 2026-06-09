package upload

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/jtenhave/not-just-noise/contracts/audio"
	"github.com/jtenhave/not-just-noise/lib/log"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type UploadService interface {
	StartUpload(ctx context.Context, upload Upload) error
	DeleteUpload(ctx context.Context, upload Upload) error
}

type uploadHandler struct {
	uploadService UploadService
}

func NewUploadHandler(uploadService UploadService) uploadHandler {
	return uploadHandler{
		uploadService: uploadService,
	}
}

func (handler *uploadHandler) HandleAudioChangedEventMessage(ctx context.Context, message string) error {
	audioChangedEvent, err := handler.toAudioChangedEvent(message)
	if err != nil {
		return njnerror.Wrapf("uploadhandler.handleAudioChangedEventMessage: failed to convert message to audio changed event: %w", err)
	}

	log.Logger(ctx).Info("received audio changed event", "audio_changed_event", audioChangedEvent)

	if !handler.isValidAudioChangedEvent(audioChangedEvent) {
		log.Logger(ctx).Error("invalid audio changed event", "audio_changed_event", audioChangedEvent)
		return nil
	}

	upload := handler.toUpload(audioChangedEvent)

	if audioChangedEvent.Status == "deleted" {
		err = handler.uploadService.DeleteUpload(ctx, upload)
		if err != nil {
			return njnerror.Wrapf("uploadhandler.handleAudioChangedEventMessage: failed to delete upload: %w", err)
		}
		return nil
	} else if audioChangedEvent.FileChanged {
		err = handler.uploadService.StartUpload(ctx, upload)
		if err != nil {
			return njnerror.Wrapf("uploadhandler.handleAudioChangedEventMessage: failed to start upload: %w", err)
		}
	}

	return nil
}

func (handler *uploadHandler) toAudioChangedEvent(payload string) (audio.AudioChangedEvent, error) {
	audioChangedEvent := audio.AudioChangedEvent{}
	err := json.Unmarshal([]byte(payload), &audioChangedEvent)
	if err != nil {
		return audio.AudioChangedEvent{}, njnerror.Wrapf("uploadhandler.toAudioChangedEvent: failed to unmarshal audio changed event: %w", err)
	}

	return audioChangedEvent, nil
}

func (handler *uploadHandler) toUpload(audioChangedEvent audio.AudioChangedEvent) Upload {
	upload := Upload{
		AudioID: audioChangedEvent.ID,
		FileURL: *audioChangedEvent.FileURL,
		Version: audioChangedEvent.Version,
	}

	return upload
}

// isValid checks if the given audioChangedEvent is valid.
func (handler *uploadHandler) isValidAudioChangedEvent(audioChangedEvent audio.AudioChangedEvent) bool {
	return audioChangedEvent.FileURL != nil && strings.HasSuffix(*audioChangedEvent.FileURL, ".mp3")
}
