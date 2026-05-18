package audio

import (
	"github.com/google/uuid"
)

type audioRepo interface {
	GetAudio(id string) (Audio, error)
	CreateAudio(audio Audio) error
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
	// toDO GET FROM THE db
	return Audio{}, nil
}

func (audioService audioService) CreateAudio(audio Audio) (string, error) {
	audio.ID = uuid.New().String()

	// toDO INSERT IN THE db

	return audio.ID, nil
}

func (audioService audioService) UpdateAudio(audio Audio) (Audio, error) {
	// toDO UPDATE IN THE db
	return audio, nil
}

func (audioService audioService) DeleteAudio(id string) error {
	// toDO DELETE FROM THE db
	return nil
}
