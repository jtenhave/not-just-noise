package audio

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jtenhave/not-just-noise/lib/errorcode"
	"github.com/jtenhave/not-just-noise/lib/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type audioServiceMock struct {
	mock.Mock
}

func (m *audioServiceMock) GetAudio(id string) (Audio, error) {
	args := m.Called(id)
	return args.Get(0).(Audio), args.Error(1)
}

func (m *audioServiceMock) CreateAudio(audio Audio) (string, error) {
	args := m.Called(audio)
	return args.Get(0).(string), args.Error(1)
}

func (m *audioServiceMock) UpdateAudio(audio Audio) (Audio, error) {
	args := m.Called(audio)
	return args.Get(0).(Audio), args.Error(1)
}

func (m *audioServiceMock) DeleteAudio(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestHandler_GetAudio_IDIsRequired(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: "",
		PathValue: func(name string) string {
			return ""
		},
	}

	response := getAudioHandler(request, audioService)
	assert.Equal(t, 400, response.StatusCode)
	assert.True(t, strings.Contains(response.Body, idIsRequired))
}

func TestHandler_GetAudio_IDNotFound(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: "",
		PathValue: func(name string) string {
			return "123abc"
		},
	}

	audioService.On("GetAudio", "123abc").Return(Audio{}, errorcode.NewErrorCode(errorcode.NotFound, "audio not found"))
	response := getAudioHandler(request, audioService)
	assert.Equal(t, 404, response.StatusCode)
	assert.True(t, strings.Contains(response.Body, audioNotFound))
}

func TestHandler_GetAudio_Failure(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: "",
		PathValue: func(name string) string {
			return "123abc"
		},
	}

	audioService.On("GetAudio", "123abc").Return(Audio{}, fmt.Errorf("failed to get audio"))
	response := getAudioHandler(request, audioService)
	assert.Equal(t, 500, response.StatusCode)
}

func TestHandler_GetAudio_Success(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: "",
		PathValue: func(name string) string {
			return "123abc"
		},
	}

	audioService.On("GetAudio", "123abc").Return(Audio{ID: "123abc", CreatorID: "456def", Title: "Test Audio", FileURL: "https://test.com/audio.mp3"}, nil)
	response := getAudioHandler(request, audioService)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, `{"id":"123abc","title":"Test Audio","creator":"456def","file_url":"https://test.com/audio.mp3","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}`, response.Body)
}

func TestHandler_CreateAudio_UnmarshalError(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: "invalid json",
		PathValue: func(name string) string {
			return ""
		},
	}

	response := createAudioHandler(request, audioService)
	assert.Equal(t, 400, response.StatusCode)
	assert.True(t, strings.Contains(response.Body, failedToUnmarshal))
}

func TestHandler_CreateAudio_TitleIsRequired(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: `{"title":"", "creator_id":"456def", "file_url":"https://test.com/audio.mp3"}`,
		PathValue: func(name string) string {
			return ""
		},
	}

	response := createAudioHandler(request, audioService)
	assert.Equal(t, 400, response.StatusCode)
	assert.True(t, strings.Contains(response.Body, titleIsRequired))
}

func TestHandler_CreateAudio_CreatorIsRequired(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: `{"title":"Test Audio", "creator_id":"", "file_url":"https://test.com/audio.mp3"}`,
		PathValue: func(name string) string {
			return ""
		},
	}

	response := createAudioHandler(request, audioService)
	assert.Equal(t, 400, response.StatusCode)
	assert.True(t, strings.Contains(response.Body, creatorIsRequired))
}

func TestHandler_CreateAudio_FileURLIsRequired(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: `{"title":"Test Audio", "creator_id":"456def", "file_url":""}`,
		PathValue: func(name string) string {
			return ""
		},
	}

	response := createAudioHandler(request, audioService)
	assert.Equal(t, 400, response.StatusCode)
	assert.True(t, strings.Contains(response.Body, fileURLIsRequired))
}

func TestHandler_CreateAudio_FileURLIsNotValid(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: `{"title":"Test Audio", "creator_id":"456def", "file_url":"invalid url"}`,
		PathValue: func(name string) string {
			return ""
		},
	}

	response := createAudioHandler(request, audioService)
	assert.Equal(t, 400, response.StatusCode)
	assert.True(t, strings.Contains(response.Body, fileURLIsNotValid))
}

func TestHandler_CreateAudio_Failure(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: `{"title":"Test Audio", "creator_id":"456def", "file_url":"https://test.com/audio.mp3"}`,
		PathValue: func(name string) string {
			return ""
		},
	}

	audioService.On("CreateAudio", Audio{Title: "Test Audio", CreatorID: "456def", FileURL: "https://test.com/audio.mp3"}).Return("", fmt.Errorf("failed to create audio"))
	response := createAudioHandler(request, audioService)
	assert.Equal(t, 500, response.StatusCode)
}

func TestHandler_CreateAudio_FailureWithCode(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: `{"title":"Test Audio", "creator_id":"456def", "file_url":"https://test.com/audio.mp3"}`,
		PathValue: func(name string) string {
			return ""
		},
	}

	audioService.On("CreateAudio", Audio{Title: "Test Audio", CreatorID: "456def", FileURL: "https://test.com/audio.mp3"}).Return("", errorcode.NewErrorCode(409, "audio already exists"))
	response := createAudioHandler(request, audioService)
	assert.Equal(t, 409, response.StatusCode)
	assert.True(t, strings.Contains(response.Body, "audio already exists"))
}

func TestHandler_CreateAudio_Success(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: `{"title":"Test Audio", "creator_id":"456def", "file_url":"https://test.com/audio.mp3"}`,
		PathValue: func(name string) string {
			return ""
		},
	}

	audioService.On("CreateAudio", Audio{Title: "Test Audio", CreatorID: "456def", FileURL: "https://test.com/audio.mp3"}).Return("123abc", nil)
	response := createAudioHandler(request, audioService)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, `{"id":"123abc"}`, response.Body)
}
