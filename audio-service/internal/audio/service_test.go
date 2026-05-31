package audio

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	audioContract "github.com/jtenhave/not-just-noise/contracts/audio"
	"github.com/jtenhave/not-just-noise/contracts/dispatch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testDispatchDestination = "audio-changed"

type transactionManagerMock struct {
	mock.Mock
}

func (m *transactionManagerMock) WithinTx(ctx context.Context, transaction func(context.Context) error) error {
	arguments := m.Called(ctx, transaction)

	var callbackErr error
	if transaction != nil {
		callbackErr = transaction(ctx)
	}

	if len(arguments) > 0 && arguments.Get(0) != nil {
		return arguments.Get(0).(error)
	}

	return callbackErr
}

type audioRepoMock struct {
	mock.Mock
}

func (m *audioRepoMock) GetAudio(ctx context.Context, id string) (Audio, error) {
	arguments := m.Called(ctx, id)

	var err error
	if e := arguments.Get(1); e != nil {
		err = e.(error)
	}

	return arguments.Get(0).(Audio), err
}

func (m *audioRepoMock) CreateAudio(ctx context.Context, audio Audio) error {
	arguments := m.Called(ctx, audio)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

func (m *audioRepoMock) UpdateAudio(ctx context.Context, audio Audio) error {
	arguments := m.Called(ctx, audio)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

func (m *audioRepoMock) DeleteAudio(ctx context.Context, id string) (int64, error) {
	arguments := m.Called(ctx, id)

	var err error
	if e := arguments.Get(1); e != nil {
		err = e.(error)
	}

	return arguments.Get(0).(int64), err
}

type dispatchClientMock struct {
	mock.Mock
}

func (m *dispatchClientMock) Dispatch(ctx context.Context, dispatch dispatch.Dispatch) error {
	arguments := m.Called(ctx, dispatch)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

func TestAudioService_GetAudio_RepoError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	dispatchClient := new(dispatchClientMock)

	service := NewAudioService(tm, repo, dispatchClient, testDispatchDestination)

	repo.On("GetAudio", mock.Anything, "123").Return(Audio{}, errors.New("db down"))

	_, err := service.GetAudio(context.Background(), "123")
	assert.Error(t, err)

	repo.AssertExpectations(t)
}

func TestAudioService_GetAudio_Success(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	dispatchClient := new(dispatchClientMock)

	service := NewAudioService(tm, repo, dispatchClient, testDispatchDestination)

	title := "Test Title"
	fileURL := "https://test.com/test.mp3"
	expectedAudio := Audio{
		ID:        "123",
		Title:     &title,
		CreatorID: "creator-1",
		FileURL:   &fileURL,
		Version:   1,
		Status:    "active",
	}

	repo.On("GetAudio", mock.Anything, "123").Return(expectedAudio, nil)

	audio, err := service.GetAudio(context.Background(), "123")
	assert.NoError(t, err)
	assert.Equal(t, expectedAudio, audio)

	repo.AssertExpectations(t)
}

func TestAudioService_CreateAudio_RepoError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	dispatchClient := new(dispatchClientMock)

	service := NewAudioService(tm, repo, dispatchClient, testDispatchDestination)

	title := "Test Title"
	fileURL := "https://test.com/test.mp3"
	audio := Audio{
		Title:     &title,
		CreatorID: "creator-1",
		FileURL:   &fileURL,
	}

	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("CreateAudio", mock.Anything, mock.Anything).Return(errors.New("insert failed"))

	_, err := service.CreateAudio(context.Background(), audio)
	assert.Error(t, err)

	dispatchClient.AssertNotCalled(t, "Dispatch")
	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestAudioService_CreateAudio_DispatchError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	dispatchClient := new(dispatchClientMock)

	service := NewAudioService(tm, repo, dispatchClient, testDispatchDestination)

	title := "Test Title"
	fileURL := "https://test.com/test.mp3"
	audio := Audio{
		Title:     &title,
		CreatorID: "creator-1",
		FileURL:   &fileURL,
	}

	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("CreateAudio", mock.Anything, mock.Anything).Return(nil)
	dispatchClient.On("Dispatch", mock.Anything, mock.Anything).Return(errors.New("dispatch failed"))

	_, err := service.CreateAudio(context.Background(), audio)
	assert.Error(t, err)

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
	dispatchClient.AssertExpectations(t)
}

func TestAudioService_CreateAudio_Success(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	dispatchClient := new(dispatchClientMock)

	service := NewAudioService(tm, repo, dispatchClient, testDispatchDestination)

	title := "Test Title"
	fileURL := "https://test.com/test.mp3"
	audio := Audio{
		Title:     &title,
		CreatorID: "creator-1",
		FileURL:   &fileURL,
	}

	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)

	var id string
	repo.On("CreateAudio", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			audio := args.Get(1).(Audio)
			assert.Equal(t, "creator-1", audio.CreatorID)
			assert.Equal(t, title, *audio.Title)
			assert.Equal(t, fileURL, *audio.FileURL)
			assert.NotEmpty(t, audio.ID)
			id = audio.ID
		}).
		Return(nil)

	dispatchClient.On("Dispatch", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			d := args.Get(1).(dispatch.Dispatch)
			assert.Equal(t, dispatch.CallbackTypeNotify, d.CallbackType)
			assert.Equal(t, testDispatchDestination, d.CallbackResource)

			var event audioContract.AudioChangedEvent
			assert.NoError(t, json.Unmarshal([]byte(d.Payload), &event))
			assert.Equal(t, audioContract.AudioChangedEventTypeCreated, event.EventType)
			assert.Equal(t, id, event.ID)
			assert.Equal(t, title, *event.Title)
			assert.Equal(t, fileURL, *event.FileURL)
		}).
		Return(nil)

	id, err := service.CreateAudio(context.Background(), audio)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
	dispatchClient.AssertExpectations(t)
}

func TestAudioService_UpdateAudio_GetError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	dispatchClient := new(dispatchClientMock)

	service := NewAudioService(tm, repo, dispatchClient, testDispatchDestination)

	newTitle := "New Title"
	updateAudio := Audio{
		ID:    "123",
		Title: &newTitle,
	}

	repo.On("GetAudio", mock.Anything, "123").Return(Audio{}, errors.New("not found"))

	err := service.UpdateAudio(context.Background(), updateAudio)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to get audio"))

	tm.AssertNotCalled(t, "WithinTx")
	repo.AssertExpectations(t)
}

func TestAudioService_UpdateAudio_RepoError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	dispatchClient := new(dispatchClientMock)

	service := NewAudioService(tm, repo, dispatchClient, testDispatchDestination)

	existingTitle := "Old Title"
	existingFileURL := "https://test.com/old.mp3"
	newTitle := "New Title"
	existingAudio := Audio{
		ID:        "123",
		Title:     &existingTitle,
		CreatorID: "creator-1",
		FileURL:   &existingFileURL,
		Version:   1,
		Status:    "active",
	}
	updateAudio := Audio{
		ID:    "123",
		Title: &newTitle,
	}

	expectedAudioToUpdate := existingAudio
	expectedAudioToUpdate.Title = &newTitle
	expectedAudioToUpdate.Version = 1

	repo.On("GetAudio", mock.Anything, "123").Return(existingAudio, nil)
	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("UpdateAudio", mock.Anything, expectedAudioToUpdate).Return(errors.New("update failed"))

	err := service.UpdateAudio(context.Background(), updateAudio)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to update audio"))

	dispatchClient.AssertNotCalled(t, "Dispatch")
	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestAudioService_UpdateAudio_DispatchError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	dispatchClient := new(dispatchClientMock)

	service := NewAudioService(tm, repo, dispatchClient, testDispatchDestination)

	existingTitle := "Old Title"
	existingFileURL := "https://test.com/old.mp3"
	newTitle := "New Title"
	existingAudio := Audio{
		ID:        "123",
		Title:     &existingTitle,
		CreatorID: "creator-1",
		FileURL:   &existingFileURL,
		Version:   1,
		Status:    "active",
	}
	updateAudio := Audio{
		ID:    "123",
		Title: &newTitle,
	}

	expectedAudioToUpdate := existingAudio
	expectedAudioToUpdate.Title = &newTitle
	expectedAudioToUpdate.Version = 1

	repo.On("GetAudio", mock.Anything, "123").Return(existingAudio, nil)
	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("UpdateAudio", mock.Anything, expectedAudioToUpdate).Return(nil)
	dispatchClient.On("Dispatch", mock.Anything, mock.Anything).Return(errors.New("dispatch failed"))

	err := service.UpdateAudio(context.Background(), updateAudio)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to raise audio updated event"))

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
	dispatchClient.AssertExpectations(t)
}

func TestAudioService_UpdateAudio_Success(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	dispatchClient := new(dispatchClientMock)

	service := NewAudioService(tm, repo, dispatchClient, testDispatchDestination)

	existingTitle := "Old Title"
	existingFileURL := "https://test.com/old.mp3"
	newTitle := "New Title"
	existingAudio := Audio{
		ID:        "123",
		Title:     &existingTitle,
		CreatorID: "creator-1",
		FileURL:   &existingFileURL,
		Version:   1,
		Status:    "active",
	}
	updateAudio := Audio{
		ID:    "123",
		Title: &newTitle,
	}

	expectedAudioToUpdate := existingAudio
	expectedAudioToUpdate.Title = &newTitle
	expectedAudioToUpdate.Version = 1

	repo.On("GetAudio", mock.Anything, "123").Return(existingAudio, nil)
	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("UpdateAudio", mock.Anything, expectedAudioToUpdate).Return(nil)

	dispatchClient.On("Dispatch", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		d := args.Get(1).(dispatch.Dispatch)
		assert.Equal(t, dispatch.CallbackTypeNotify, d.CallbackType)
		assert.Equal(t, testDispatchDestination, d.CallbackResource)

		var event audioContract.AudioChangedEvent
		assert.NoError(t, json.Unmarshal([]byte(d.Payload), &event))
		assert.Equal(t, audioContract.AudioChangedEventTypeUpdated, event.EventType)
		assert.Equal(t, "123", event.ID)
		assert.Equal(t, newTitle, *event.Title)
		assert.Equal(t, int64(2), event.Version)
	}).Return(nil)

	err := service.UpdateAudio(context.Background(), updateAudio)
	assert.NoError(t, err)

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
	dispatchClient.AssertExpectations(t)
}

func TestAudioService_DeleteAudio_GetError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	dispatchClient := new(dispatchClientMock)

	service := NewAudioService(tm, repo, dispatchClient, testDispatchDestination)

	repo.On("GetAudio", mock.Anything, "123").Return(Audio{}, errors.New("not found"))

	err := service.DeleteAudio(context.Background(), "123")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to get audio"))

	tm.AssertNotCalled(t, "WithinTx")
	repo.AssertNotCalled(t, "DeleteAudio")
	repo.AssertExpectations(t)
}

func TestAudioService_DeleteAudio_RepoError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	dispatchClient := new(dispatchClientMock)

	service := NewAudioService(tm, repo, dispatchClient, testDispatchDestination)

	title := "Test Title"
	fileURL := "https://test.com/test.mp3"
	existingAudio := Audio{
		ID:        "123",
		Title:     &title,
		CreatorID: "creator-1",
		FileURL:   &fileURL,
		Version:   1,
		Status:    "active",
	}

	repo.On("GetAudio", mock.Anything, "123").Return(existingAudio, nil)
	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("DeleteAudio", mock.Anything, "123").Return(int64(0), errors.New("delete failed"))

	err := service.DeleteAudio(context.Background(), "123")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to delete audio"))

	dispatchClient.AssertNotCalled(t, "Dispatch")
	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestAudioService_DeleteAudio_DispatchError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	dispatchClient := new(dispatchClientMock)

	service := NewAudioService(tm, repo, dispatchClient, testDispatchDestination)

	title := "Test Title"
	fileURL := "https://test.com/test.mp3"
	existingAudio := Audio{
		ID:        "123",
		Title:     &title,
		CreatorID: "creator-1",
		FileURL:   &fileURL,
		Version:   1,
		Status:    "active",
	}

	repo.On("GetAudio", mock.Anything, "123").Return(existingAudio, nil)
	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("DeleteAudio", mock.Anything, "123").Return(int64(2), nil)
	dispatchClient.On("Dispatch", mock.Anything, mock.Anything).Return(errors.New("dispatch failed"))

	err := service.DeleteAudio(context.Background(), "123")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to raise audio deleted event"))

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
	dispatchClient.AssertExpectations(t)
}

func TestAudioService_DeleteAudio_Success(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	dispatchClient := new(dispatchClientMock)

	service := NewAudioService(tm, repo, dispatchClient, testDispatchDestination)

	title := "Test Title"
	fileURL := "https://test.com/test.mp3"
	existingAudio := Audio{
		ID:        "123",
		Title:     &title,
		CreatorID: "creator-1",
		FileURL:   &fileURL,
		Version:   1,
		Status:    "active",
	}

	repo.On("GetAudio", mock.Anything, "123").Return(existingAudio, nil)
	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("DeleteAudio", mock.Anything, "123").Return(int64(2), nil)
	dispatchClient.On("Dispatch", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		d := args.Get(1).(dispatch.Dispatch)
		assert.Equal(t, dispatch.CallbackTypeNotify, d.CallbackType)
		assert.Equal(t, testDispatchDestination, d.CallbackResource)

		var event audioContract.AudioChangedEvent
		assert.NoError(t, json.Unmarshal([]byte(d.Payload), &event))
		assert.Equal(t, audioContract.AudioChangedEventTypeDeleted, event.EventType)
		assert.Equal(t, "123", event.ID)
		assert.Equal(t, "deleted", event.Status)
		assert.Equal(t, int64(2), event.Version)
	}).Return(nil)

	err := service.DeleteAudio(context.Background(), "123")
	assert.NoError(t, err)

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
	dispatchClient.AssertExpectations(t)
}
