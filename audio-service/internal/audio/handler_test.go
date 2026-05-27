package audio

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/jtenhave/not-just-noise/lib/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type serviceMock struct {
	mock.Mock
}

func (m *serviceMock) GetAudio(ctx context.Context, id string) (Audio, error) {
	arguments := m.Called(ctx, id)

	var err error
	if e := arguments.Get(1); e != nil {
		err = e.(error)
	}

	return arguments.Get(0).(Audio), err
}

func (m *serviceMock) CreateAudio(ctx context.Context, audio Audio) (string, error) {
	arguments := m.Called(ctx, audio)

	var err error
	if e := arguments.Get(1); e != nil {
		err = e.(error)
	}

	return arguments.Get(0).(string), err
}

func (m *serviceMock) UpdateAudio(ctx context.Context, audio Audio) error {
	arguments := m.Called(ctx, audio)

	var err error
	if e := arguments.Get(1); e != nil {
		err = e.(error)
	}

	return err
}

func (m *serviceMock) DeleteAudio(ctx context.Context, id string) error {
	arguments := m.Called(ctx, id)

	var err error
	if e := arguments.Get(1); e != nil {
		err = e.(error)
	}

	return err
}

func (m *serviceMock) PublishAudio(ctx context.Context, audioPublishEvent AudioPublishEvent) error {
	arguments := m.Called(ctx, audioPublishEvent)
	var err error
	if e := arguments.Get(1); e != nil {
		err = e.(error)
	}
	return err
}

func TestAudioHandler_GetAudio_IDIsRequired(t *testing.T) {
	mockService := new(serviceMock)

	request := http.Request{
		Context: context.Background(),
		PathValues: map[string]string{
			"id": "",
		},
		Body: "",
	}

	response := getAudioHandler(request, mockService)

	assert.Equal(t, 400, response.Code)
	mockService.AssertNotCalled(t, "GetAudio")
}

func TestAudioHandler_GetAudio_Success(t *testing.T) {
	mockService := new(serviceMock)

	request := http.Request{
		Context: context.Background(),
		PathValues: map[string]string{
			"id": "123",
		},
		Body: "",
	}

	title := "Test Title"
	fileURL := "https://test.com/test.mp3"
	expectedAudio := Audio{
		ID:        "123",
		Title:     &title,
		CreatorID: "123",
		FileURL:   &fileURL,
	}

	mockService.On("GetAudio", context.Background(), "123").Return(expectedAudio, nil)

	response := getAudioHandler(request, mockService)

	assert.Equal(t, 200, response.Code)

	body := AudioResponse{}
	err := json.Unmarshal([]byte(*response.Body), &body)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	assert.Equal(t, expectedAudio.ID, body.ID)
	assert.Equal(t, *expectedAudio.Title, body.Title)
	assert.Equal(t, expectedAudio.CreatorID, body.Creator)
	assert.Equal(t, *expectedAudio.FileURL, body.FileURL)
}

func TestAudioHandler_CreateAudio_TitleIsRequired(t *testing.T) {
	requestBody := CreateAudioRequest{
		Title:     "",
		CreatorID: "123",
		FileURL:   "https://test.com/test.mp3",
	}

	errors := requestBody.Validate()

	assert.Equal(t, 1, len(errors))
	assert.Equal(t, titleIsRequired, errors[0])
}

func TestAudioHandler_CreateAudio_CreatorIDIsRequired(t *testing.T) {
	requestBody := CreateAudioRequest{
		Title:     "Test Title",
		CreatorID: "",
		FileURL:   "https://test.com/test.mp3",
	}

	errors := requestBody.Validate()

	assert.Equal(t, 1, len(errors))
	assert.Equal(t, creatorIsRequired, errors[0])
}

func TestAudioHandler_CreateAudio_FileURLIsRequired(t *testing.T) {
	requestBody := CreateAudioRequest{
		Title:     "Test Title",
		CreatorID: "123",
		FileURL:   "",
	}

	errors := requestBody.Validate()

	assert.Equal(t, 1, len(errors))
	assert.Equal(t, fileURLIsRequired, errors[0])
}

func TestAudioHandler_CreateAudio_FileURLIsValid(t *testing.T) {
	requestBody := CreateAudioRequest{
		Title:     "Test Title",
		CreatorID: "123",
		FileURL:   "not-a-valid-url",
	}

	errors := requestBody.Validate()

	assert.Equal(t, 1, len(errors))
	assert.Equal(t, fileURLIsNotValid, errors[0])
}

func TestAudioHandler_CreateAudio_Success(t *testing.T) {
	mockService := new(serviceMock)
	requestBody := CreateAudioRequest{
		Title:     "Test Title",
		CreatorID: "123",
		FileURL:   "https://test.com/test.mp3",
	}

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	mockService.On("CreateAudio", context.Background(), requestBody.ToAudio()).Return("123", nil)

	request := http.Request{
		Context: context.Background(),
		PathValues: map[string]string{
			"id": "123",
		},
		Body: string(requestBodyJSON),
	}

	response := createAudioHandler(request, mockService)

	assert.Equal(t, 200, response.Code)

	body := CreateAudioResponse{}
	err = json.Unmarshal([]byte(*response.Body), &body)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	assert.Equal(t, "123", body.ID)
}

func TestAudioHandler_UpdateAudio_IDIsRequired(t *testing.T) {
	mockService := new(serviceMock)
	request := http.Request{
		Context: context.Background(),
		PathValues: map[string]string{
			"id": "",
		},
		Body: "",
	}

	response := updateAudioHandler(request, mockService)

	assert.Equal(t, 400, response.Code)
	assert.True(t, strings.Contains(*response.Body, idIsRequired))
	mockService.AssertNotCalled(t, "UpdateAudio")
}

func TestAudioHandler_UpdateAudio_TitleIsRequired(t *testing.T) {
	title := ""
	requestBody := UpdateAudioRequest{
		Title:   &title,
		FileURL: nil,
	}

	errors := requestBody.Validate()
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, titleIsRequired, errors[0])
}

func TestAudioHandler_UpdateAudio_FileURLIsRequired(t *testing.T) {
	fileURL := ""
	requestBody := UpdateAudioRequest{
		Title:   nil,
		FileURL: &fileURL,
	}

	errors := requestBody.Validate()
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, fileURLIsRequired, errors[0])
}

func TestAudioHandler_UpdateAudio_FileURLIsNotValid(t *testing.T) {
	fileURL := "not-a-valid-url"
	requestBody := UpdateAudioRequest{
		Title:   nil,
		FileURL: &fileURL,
	}

	errors := requestBody.Validate()
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, fileURLIsNotValid, errors[0])
}

func TestAudioHandler_DeleteAudio_IDIsRequired(t *testing.T) {
	mockService := new(serviceMock)

	request := http.Request{
		Context: context.Background(),
		PathValues: map[string]string{
			"id": "",
		},
		Body: "",
	}

	response := deleteAudioHandler(request, mockService)

	assert.Equal(t, 400, response.Code)
	mockService.AssertNotCalled(t, "DeleteAudio")
}
