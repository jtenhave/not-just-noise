package audio

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jtenhave/not-just-noise/lib/errorcode"
)

type audioRepo interface {
	GetAudioByID(id string) (Audio, error)
	GetAudioByCreatorIDAndTitle(creatorID string, title string) (Audio, error)
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
	audio, err := audioService.audioRepo.GetAudioByID(id)
	if err != nil {
		return Audio{}, err
	}

	return audio, nil
}

func (audioService audioService) CreateAudio(audio Audio) (string, error) {
	audio.ID = uuid.New().String()

	_, err := audioService.audioRepo.GetAudioByCreatorIDAndTitle(audio.CreatorID, audio.Title)
	if err == nil {
		return "", errorcode.NewErrorCode(errorcode.Conflict, fmt.Sprintf("audio with title '%s' already exists for creator '%s'", audio.Title, audio.CreatorID))
	} else if errorcode.ErrorCode(err) != errorcode.NotFound {
		return "", err
	}

	err = audioService.audioRepo.CreateAudio(audio)
	if err != nil {
		return "", err
	}

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
