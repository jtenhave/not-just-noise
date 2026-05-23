package audioing

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jtenhave/not-just-noise/audio-service/internal/audio"
	"github.com/jtenhave/not-just-noise/lib/http"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type audioServiceMock struct {
	mock.Mock
}

func (m *audioServiceMock) GetAudio(ctx context.Context, id string) (audio.Audio, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(audio.Audio), args.Error(1)
}

func (m *audioServiceMock) CreateAudio(ctx context.Context, a audio.Audio) (string, error) {
	args := m.Called(ctx, a)
	return args.Get(0).(string), args.Error(1)
}

func (m *audioServiceMock) UpdateAudio(ctx context.Context, a audio.UpdateAudio) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m *audioServiceMock) DeleteAudio(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestHandler_GetAudio_IDIsRequired(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: "",
		PathValues: map[string]string{
			"id": "",
		},
	}

	response := getAudioHandler(request, audioService)
	assertErrorResponse(t, response, 400, idIsRequired)
}

func TestHandler_GetAudio_IDNotFound(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: "",
		PathValues: map[string]string{
			"id": "123abc",
		},
	}

	audioService.On("GetAudio", mock.Anything, "123abc").Return(audio.Audio{}, njnerror.NewNJNError(njnerror.NotFound, "audio not found"))
	response := getAudioHandler(request, audioService)
	assertErrorResponse(t, response, 404, audioNotFound)
}

func TestHandler_GetAudio_Failure(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: "",
		PathValues: map[string]string{
			"id": "123abc",
		},
	}

	audioService.On("GetAudio", mock.Anything, "123abc").Return(audio.Audio{}, fmt.Errorf("failed to get audio"))
	response := getAudioHandler(request, audioService)
	assertErrorResponse(t, response, 500, "failed to get audio")
}

func TestHandler_GetAudio_Success(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: "",
		PathValues: map[string]string{
			"id": "123abc",
		},
	}

	audioService.On("GetAudio", mock.Anything, "123abc").Return(audio.Audio{ID: "123abc", CreatorID: "456def", Title: "Test Audio", FileURL: "https://test.com/audio.mp3"}, nil)
	response := getAudioHandler(request, audioService)
	assert.Equal(t, 200, response.Code())

	body, ok := response.Body().(AudioResponse)
	assert.True(t, ok)
	assert.Equal(t, "123abc", body.ID)
	assert.Equal(t, "456def", body.Creator)
	assert.Equal(t, "Test Audio", body.Title)
	assert.Equal(t, "https://test.com/audio.mp3", body.FileURL)
}

func TestHandler_CreateAudio_UnmarshalError(t *testing.T) {
	audioService := new(audioServiceMock)

	body := "invalid object"

	request := http.Request{
		Body: body,
		PathValues: map[string]string{
			"id": "123abc",
		},
	}

	response := createAudioHandler(request, audioService)
	assertErrorResponse(t, response, 400, failedToUnmarshal)
}

func TestHandler_CreateAudio_TitleIsRequired(t *testing.T) {
	audioService := new(audioServiceMock)

	body := CreateAudioRequest{
		Title:     "",
		CreatorID: "456def",
		FileURL:   "https://test.com/audio.mp3",
	}

	request := http.Request{
		Body: body,
		PathValues: map[string]string{
			"id": "123abc",
		},
	}

	response := createAudioHandler(request, audioService)
	assertErrorResponse(t, response, 400, titleIsRequired)
}

func TestHandler_CreateAudio_CreatorIsRequired(t *testing.T) {
	audioService := new(audioServiceMock)

	body := CreateAudioRequest{
		Title:     "Test Audio",
		CreatorID: "",
		FileURL:   "https://test.com/audio.mp3",
	}

	request := http.Request{
		Body: body,
		PathValues: map[string]string{
			"id": "123abc",
		},
	}

	response := createAudioHandler(request, audioService)
	assert.Equal(t, 400, response.Code())
	assertErrorResponse(t, response, 400, creatorIsRequired)
}

func TestHandler_CreateAudio_FileURLIsRequired(t *testing.T) {
	audioService := new(audioServiceMock)

	body := CreateAudioRequest{
		Title:     "Test Audio",
		CreatorID: "456def",
		FileURL:   "",
	}

	request := http.Request{
		Body: body,
		PathValues: map[string]string{
			"id": "123abc",
		},
	}

	response := createAudioHandler(request, audioService)
	assertErrorResponse(t, response, 400, fileURLIsRequired)
}

func TestHandler_CreateAudio_FileURLIsNotValid(t *testing.T) {
	audioService := new(audioServiceMock)
	body := CreateAudioRequest{
		Title:     "Test Audio",
		CreatorID: "456def",
		FileURL:   "invalid url",
	}
	request := http.Request{
		Body: body,
		PathValues: map[string]string{
			"id": "123abc",
		},
	}

	response := createAudioHandler(request, audioService)
	assertErrorResponse(t, response, 400, fileURLIsNotValid)
}

func TestHandler_CreateAudio_Failure(t *testing.T) {
	audioService := new(audioServiceMock)
	body := CreateAudioRequest{
		Title:     "Test Audio",
		CreatorID: "456def",
		FileURL:   "https://test.com/audio.mp3",
	}
	request := http.Request{
		Body: body,
		PathValues: map[string]string{
			"id": "123abc",
		},
	}

	audioService.On("CreateAudio", mock.Anything, audio.Audio{Title: "Test Audio", CreatorID: "456def", FileURL: "https://test.com/audio.mp3"}).Return("", fmt.Errorf("failed to create audio"))
	response := createAudioHandler(request, audioService)
	assertErrorResponse(t, response, 500, "failed to create audio")
}

func TestHandler_CreateAudio_Success(t *testing.T) {
	audioService := new(audioServiceMock)
	body := CreateAudioRequest{
		Title:     "Test Audio",
		CreatorID: "456def",
		FileURL:   "https://test.com/audio.mp3",
	}
	request := http.Request{
		Body: body,
		PathValues: map[string]string{
			"id": "123abc",
		},
	}

	audioService.On("CreateAudio", mock.Anything, audio.Audio{Title: "Test Audio", CreatorID: "456def", FileURL: "https://test.com/audio.mp3"}).Return("123abc", nil)
	response := createAudioHandler(request, audioService)
	assert.Equal(t, 200, response.Code())

	responseBody, ok := response.Body().(CreateAudioResponse)
	assert.True(t, ok)
	assert.Equal(t, "123abc", responseBody.ID)
}

func TestHandler_UpdateAudio_IDIsRequired(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: UpdateAudioRequest{},
		PathValues: map[string]string{
			"id": "",
		},
	}

	response := updateAudioHandler(request, audioService)
	assertErrorResponse(t, response, 400, idIsRequired)
}

func TestHandler_UpdateAudio_TitleIsRequired(t *testing.T) {
	audioService := new(audioServiceMock)

	title := ""
	body := UpdateAudioRequest{
		Title:   &title,
		FileURL: nil,
	}
	request := http.Request{
		Body: body,
		PathValues: map[string]string{
			"id": "123abc",
		},
	}

	response := updateAudioHandler(request, audioService)
	assertErrorResponse(t, response, 400, titleIsRequired)
}

func TestHandler_UpdateAudio_FileURLIsRequired(t *testing.T) {
	audioService := new(audioServiceMock)

	fileURL := ""
	body := UpdateAudioRequest{
		Title:   nil,
		FileURL: &fileURL,
	}
	request := http.Request{
		Body: body,
		PathValues: map[string]string{
			"id": "123abc",
		},
	}

	response := updateAudioHandler(request, audioService)
	assertErrorResponse(t, response, 400, fileURLIsRequired)
}

func TestHandler_UpdateAudio_FileURLIsNotValid(t *testing.T) {
	audioService := new(audioServiceMock)
	fileURL := "invalid url"
	body := UpdateAudioRequest{
		Title:   nil,
		FileURL: &fileURL,
	}
	request := http.Request{
		Body: body,
		PathValues: map[string]string{
			"id": "123abc",
		},
	}

	response := updateAudioHandler(request, audioService)
	assertErrorResponse(t, response, 400, fileURLIsNotValid)
}

func TestHandler_UpdateAudio_Failure(t *testing.T) {
	audioService := new(audioServiceMock)

	title := "Test Audio"
	fileURL := "https://test.com/audio.mp3"
	body := UpdateAudioRequest{
		Title:   &title,
		FileURL: &fileURL,
	}
	request := http.Request{
		Body: body,
		PathValues: map[string]string{
			"id": "123abc",
		},
	}

	audioService.On("UpdateAudio", mock.Anything, mock.MatchedBy(func(a audio.UpdateAudio) bool {
		return a.ID == "123abc"
	})).Return(fmt.Errorf("failed to update audio"))
	response := updateAudioHandler(request, audioService)
	assertErrorResponse(t, response, 500, "failed to update audio")
}

func TestHandler_UpdateAudio_Success(t *testing.T) {
	audioService := new(audioServiceMock)
	title := "Test Audio"
	fileURL := "https://test.com/audio.mp3"
	body := UpdateAudioRequest{
		Title:   &title,
		FileURL: &fileURL,
	}
	request := http.Request{
		Body: body,
		PathValues: map[string]string{
			"id": "123abc",
		},
	}

	audioService.On("UpdateAudio", mock.Anything, mock.MatchedBy(func(a audio.UpdateAudio) bool {
		return a.ID == "123abc"
	})).Return(nil)
	response := updateAudioHandler(request, audioService)
	assert.Equal(t, 204, response.Code())
}

func TestHandler_DeleteAudio_IDIsRequired(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: "",
		PathValues: map[string]string{
			"id": "",
		},
	}

	response := deleteAudioHandler(request, audioService)
	assertErrorResponse(t, response, 400, idIsRequired)
}

func TestHandler_DeleteAudio_Failure(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: "",
		PathValues: map[string]string{
			"id": "123abc",
		},
	}

	audioService.On("DeleteAudio", mock.Anything, "123abc").Return(fmt.Errorf("failed to delete audio"))
	response := deleteAudioHandler(request, audioService)
	assertErrorResponse(t, response, 500, "failed to delete audio")
}

func TestHandler_DeleteAudio_Success(t *testing.T) {
	audioService := new(audioServiceMock)
	request := http.Request{
		Body: "",
		PathValues: map[string]string{
			"id": "123abc",
		},
	}

	audioService.On("DeleteAudio", mock.Anything, "123abc").Return(nil)
	response := deleteAudioHandler(request, audioService)
	assert.Equal(t, 204, response.Code())
}

func assertErrorResponse(t *testing.T, response http.Response, expectedCode int, expectedMessage string) {
	assert.Equal(t, expectedCode, response.Code())
	assert.NotNil(t, response.Body())

	errorResponseBody, ok := response.Body().(map[string]string)
	assert.True(t, ok)
	assert.True(t, strings.Contains(errorResponseBody["error"], expectedMessage))
}
