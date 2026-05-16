package audio

import (
	"time"

	"github.com/google/uuid"
)

type audioService struct{}

func NewAudioService() audioService {
	return audioService{}

}

func (audioService audioService) GetAudio(id string) (Audio, error) {
	// toDO GET FROM THE db
	return Audio{}, nil
}

func (audioService audioService) CreateAudio(audio Audio) (Audio, error) {
	audio.ID = uuid.New().String()
	audio.CreatedAt = time.Now()
	audio.UpdatedAt = time.Now()

	// toDO INSERT IN THE db

	return audio, nil
}

func (audioService audioService) UpdateAudio(audio Audio) (Audio, error) {
	// toDO UPDATE IN THE db
	return audio, nil
}

func (audioService audioService) DeleteAudio(id string) error {
	// toDO DELETE FROM THE db
	return nil
}
