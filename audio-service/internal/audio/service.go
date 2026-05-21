package audio

import (
	"github.com/google/uuid"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type audioRepo interface {
	GetAudio(id string) (Audio, error)
	CreateAudio(audio Audio) error
	UpdateAudio(audio UpdateAudio) error
	DeleteAudio(id string) error
}

type audioService struct {
	audioRepo audioRepo
}

func NewAudioService(audioRepo audioRepo) audioService {
	return audioService{
		audioRepo: audioRepo,
	}
}

func (audioService audioService) GetAudio(id string) (Audio, error) {
	audio, err := audioService.audioRepo.GetAudio(id)
	if err != nil {
		return Audio{}, njnerror.Wrapf("audioservice.GetAudio: failed to get audio: %w", err)
	}

	return audio, nil
}

func (audioService audioService) CreateAudio(audio Audio) (string, error) {
	audio.ID = uuid.New().String()
	err := audioService.audioRepo.CreateAudio(audio)
	if err != nil {
		return "", njnerror.Wrapf("audioservice.CreateAudio: failed to create audio: %w", err)
	}

	return audio.ID, nil
}

func (audioService audioService) UpdateAudio(audio UpdateAudio) error {
	err := audioService.audioRepo.UpdateAudio(audio)
	if err != nil {
		return njnerror.Wrapf("audioservice.UpdateAudio: failed to update audio: %w", err)
	}

	return nil
}

func (audioService audioService) DeleteAudio(id string) error {
	err := audioService.audioRepo.DeleteAudio(id)
	if err != nil {
		return njnerror.Wrapf("audioservice.DeleteAudio: failed to delete audio: %w", err)
	}

	return nil
}
