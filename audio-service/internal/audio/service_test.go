package audio

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type audioRepoMock struct {
	mock.Mock
}

func (m *audioRepoMock) GetAudio(id string) (Audio, error) {
	args := m.Called(id)
	return args.Get(0).(Audio), args.Error(1)
}

func (m *audioRepoMock) CreateAudio(audio Audio) error {
	args := m.Called(audio)
	return args.Error(0)
}

func (m *audioRepoMock) UpdateAudio(audio UpdateAudio) error {
	args := m.Called(audio)
	return args.Error(0)
}

func (m *audioRepoMock) DeleteAudio(id string) error {
	args := m.Called(id)
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

func TestUpdateAudio_Failure(t *testing.T) {
	audioRepo := new(audioRepoMock)
	audioService := NewAudioService(audioRepo)

	audioRepo.On("UpdateAudio", mock.MatchedBy(func(audio UpdateAudio) bool {
		return audio.ID == "123abc"
	})).Return(fmt.Errorf("failed to update audio"))

	title := "New Title"
	err := audioService.UpdateAudio(UpdateAudio{ID: "123abc", Title: &title, FileURL: nil})
	assert.Error(t, err)
}

func TestUpdateAudio_Success(t *testing.T) {
	audioRepo := new(audioRepoMock)
	audioService := NewAudioService(audioRepo)

	var updatedTitle *string
	var updatedFileURL *string
	audioRepo.On("UpdateAudio", mock.MatchedBy(func(audio UpdateAudio) bool {
		return audio.ID == "123abc"
	})).Run(func(args mock.Arguments) {
		audio := args.Get(0).(UpdateAudio)
		updatedTitle = audio.Title
		updatedFileURL = audio.FileURL
	}).Return(nil)

	title := "New Title"
	err := audioService.UpdateAudio(UpdateAudio{ID: "123abc", Title: &title, FileURL: nil})
	assert.NoError(t, err)
	assert.NotNil(t, updatedTitle)
	assert.Equal(t, title, *updatedTitle)
	assert.Nil(t, updatedFileURL)
}

func TestDeleteAudio_Failure(t *testing.T) {
	audioRepo := new(audioRepoMock)
	audioService := NewAudioService(audioRepo)

	audioRepo.On("DeleteAudio", "123abc").Return(fmt.Errorf("failed to delete audio"))
	err := audioService.DeleteAudio("123abc")
	assert.Error(t, err)
}

func TestDeleteAudio_Success(t *testing.T) {
	audioRepo := new(audioRepoMock)
	audioService := NewAudioService(audioRepo)

	audioRepo.On("DeleteAudio", "123abc").Return(nil)
	err := audioService.DeleteAudio("123abc")
	assert.NoError(t, err)
}
