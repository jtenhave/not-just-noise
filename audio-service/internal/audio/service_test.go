package audio

import (
	"fmt"
	"testing"

	"github.com/jtenhave/not-just-noise/lib/errorcode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type audioRepoMock struct {
	mock.Mock
}

func (m *audioRepoMock) GetAudioByID(id string) (Audio, error) {
	args := m.Called(id)
	return args.Get(0).(Audio), args.Error(1)
}

func (m *audioRepoMock) GetAudioByCreatorIDAndTitle(creatorID string, title string) (Audio, error) {
	args := m.Called(creatorID, title)
	return args.Get(0).(Audio), args.Error(1)
}

func (m *audioRepoMock) CreateAudio(audio Audio) error {
	args := m.Called(audio)
	return args.Error(0)
}

func TestGetAudio_Failure(t *testing.T) {
	audioRepo := new(audioRepoMock)
	audioService := NewAudioService(audioRepo)

	audioRepo.On("GetAudio", "123abc").Return(Audio{}, fmt.Errorf("failed to get audio"))
	_, err := audioService.GetAudio("123abc")
	assert.Error(t, err)
}

func TestGetAudio_Success(t *testing.T) {
	audioRepo := new(audioRepoMock)
	audioService := NewAudioService(audioRepo)

	audioRepo.On("GetAudio", "123abc").Return(Audio{ID: "123abc", CreatorID: "456def", Title: "Test Audio", FileURL: "https://test.com/audio.mp3"}, nil)

	audio, err := audioService.GetAudio("123abc")
	assert.NoError(t, err)
	assert.Equal(t, "123abc", audio.ID)
	assert.Equal(t, "456def", audio.CreatorID)
	assert.Equal(t, "Test Audio", audio.Title)
	assert.Equal(t, "https://test.com/audio.mp3", audio.FileURL)
}


func TestCreateAudio_AudioAlreadyExists(t *testing.T) {
	audioRepo := new(audioRepoMock)
	audioService := NewAudioService(audioRepo)

	audioRepo.On("GetAudioByCreatorIDAndTitle", "456def", "Test Audio").Return(Audio{ID: "123abc", CreatorID: "456def", Title: "Test Audio", FileURL: "https://test.com/audio.mp3"}, nil)
	audioRepo.On("CreateAudio", mock.MatchedBy(func(audio Audio) bool {
		return audio.CreatorID == "456def" && audio.Title == "Test Audio" && audio.FileURL == "https://test.com/audio.mp3"
	})).Return("123abc", nil)

	_, err := audioService.CreateAudio(Audio{CreatorID: "456def", Title: "Test Audio", FileURL: "https://test.com/audio.mp3"})

	assert.Error(t, err)
	assert.True(t, errorcode.ErrorCode(err) == errorcode.Conflict)
}

func TestCreateAudio_Failure(t *testing.T) {
	audioRepo := new(audioRepoMock)
	audioService := NewAudioService(audioRepo)

	audioRepo.On("CreateAudio", mock.MatchedBy(func(audio Audio) bool {
		return audio.CreatorID == "456def" && audio.Title == "Test Audio" && audio.FileURL == "https://test.com/audio.mp3"
	})).Return(fmt.Errorf("failed to create audio"))

	_, err := audioService.CreateAudio(Audio{CreatorID: "456def", Title: "Test Audio", FileURL: "https://test.com/audio.mp3"})

	assert.Error(t, err)
}

func TestCreateAudio_Success(t *testing.T) {
	audioRepo := new(audioRepoMock)
	audioService := NewAudioService(audioRepo)

	audioRepo.On("CreateAudio", mock.MatchedBy(func(audio Audio) bool {
		return audio.CreatorID == "456def" && audio.Title == "Test Audio" && audio.FileURL == "https://test.com/audio.mp3"
	})).Return(nil)

	id, err := audioService.CreateAudio(Audio{CreatorID: "456def", Title: "Test Audio", FileURL: "https://test.com/audio.mp3"})
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
}
