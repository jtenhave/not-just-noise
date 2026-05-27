package audio

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jtenhave/not-just-noise/audio-service/internal/adapters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func (m *audioRepoMock) DeleteAudio(ctx context.Context, id string) error {
	arguments := m.Called(ctx, id)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

type audioPublishAdapterMock struct {
	mock.Mock
}

func (m *audioPublishAdapterMock) PublishCreated(ctx context.Context, audio adapters.AudioPublishPayload) error {
	arguments := m.Called(ctx, audio)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

func (m *audioPublishAdapterMock) PublishUpdated(ctx context.Context, audio adapters.AudioPublishPayload) error {
	arguments := m.Called(ctx, audio)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

func (m *audioPublishAdapterMock) PublishDeleted(ctx context.Context, audio adapters.AudioPublishPayload) error {
	arguments := m.Called(ctx, audio)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

func TestAudioService_GetAudio_RepoError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	publishAdapter := new(audioPublishAdapterMock)

	service := NewAudioService(tm, repo, publishAdapter)

	repo.On("GetAudio", mock.Anything, "123").Return(Audio{}, errors.New("db down"))

	_, err := service.GetAudio(context.Background(), "123")
	assert.Error(t, err)

	repo.AssertExpectations(t)
}

func TestAudioService_GetAudio_Success(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	publishAdapter := new(audioPublishAdapterMock)

	service := NewAudioService(tm, repo, publishAdapter)

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
	publishAdapter := new(audioPublishAdapterMock)

	service := NewAudioService(tm, repo, publishAdapter)

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

	publishAdapter.AssertNotCalled(t, "PublishCreated")
	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestAudioService_CreateAudio_PublishError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	publishAdapter := new(audioPublishAdapterMock)

	service := NewAudioService(tm, repo, publishAdapter)

	title := "Test Title"
	fileURL := "https://test.com/test.mp3"
	audio := Audio{
		Title:     &title,
		CreatorID: "creator-1",
		FileURL:   &fileURL,
	}

	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("CreateAudio", mock.Anything, mock.Anything).Return(nil)
	publishAdapter.On("PublishCreated", mock.Anything, mock.Anything).Return(errors.New("publish failed"))

	_, err := service.CreateAudio(context.Background(), audio)
	assert.Error(t, err)

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
	publishAdapter.AssertExpectations(t)
}

func TestAudioService_CreateAudio_Success(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	publishAdapter := new(audioPublishAdapterMock)

	service := NewAudioService(tm, repo, publishAdapter)

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

	publishAdapter.On("PublishCreated", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			payload := args.Get(1).(adapters.AudioPublishPayload)
			assert.Equal(t, id, payload.ID)
			assert.Equal(t, title, *payload.Title)
			assert.Equal(t, fileURL, *payload.FileURL)
		}).
		Return(nil)

	id, err := service.CreateAudio(context.Background(), audio)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
	publishAdapter.AssertExpectations(t)
}

func TestAudioService_UpdateAudio_GetError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	publishAdapter := new(audioPublishAdapterMock)

	service := NewAudioService(tm, repo, publishAdapter)

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
	publishAdapter := new(audioPublishAdapterMock)

	service := NewAudioService(tm, repo, publishAdapter)

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

	repo.On("GetAudio", mock.Anything, "123").Return(existingAudio, nil)
	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("UpdateAudio", mock.Anything, updateAudio).Return(errors.New("update failed"))

	err := service.UpdateAudio(context.Background(), updateAudio)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to update audio"))

	publishAdapter.AssertNotCalled(t, "PublishUpdated")
	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestAudioService_UpdateAudio_PublishError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	publishAdapter := new(audioPublishAdapterMock)

	service := NewAudioService(tm, repo, publishAdapter)

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

	repo.On("GetAudio", mock.Anything, "123").Return(existingAudio, nil)
	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("UpdateAudio", mock.Anything, updateAudio).Return(nil)
	publishAdapter.On("PublishUpdated", mock.Anything, mock.Anything).Return(errors.New("publish failed"))

	err := service.UpdateAudio(context.Background(), updateAudio)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to publish updated"))

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
	publishAdapter.AssertExpectations(t)
}

func TestAudioService_UpdateAudio_Success(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	publishAdapter := new(audioPublishAdapterMock)

	service := NewAudioService(tm, repo, publishAdapter)

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

	repo.On("GetAudio", mock.Anything, "123").Return(existingAudio, nil)
	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("UpdateAudio", mock.Anything, updateAudio).Return(nil)

	publishAdapter.On("PublishUpdated", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		payload := args.Get(1).(adapters.AudioPublishPayload)
		assert.Equal(t, "123", payload.ID)
		assert.Equal(t, newTitle, *payload.Title)
		assert.Equal(t, int64(2), payload.Version)
	}).Return(nil)

	err := service.UpdateAudio(context.Background(), updateAudio)
	assert.NoError(t, err)

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
	publishAdapter.AssertExpectations(t)
}

func TestAudioService_DeleteAudio_GetError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	publishAdapter := new(audioPublishAdapterMock)

	service := NewAudioService(tm, repo, publishAdapter)

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
	publishAdapter := new(audioPublishAdapterMock)

	service := NewAudioService(tm, repo, publishAdapter)

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
	repo.On("DeleteAudio", mock.Anything, "123").Return(errors.New("delete failed"))

	err := service.DeleteAudio(context.Background(), "123")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to delete audio"))

	publishAdapter.AssertNotCalled(t, "PublishDeleted")
	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestAudioService_DeleteAudio_PublishError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	publishAdapter := new(audioPublishAdapterMock)

	service := NewAudioService(tm, repo, publishAdapter)

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
	repo.On("DeleteAudio", mock.Anything, "123").Return(nil)
	publishAdapter.On("PublishDeleted", mock.Anything, mock.Anything).Return(errors.New("publish failed"))

	err := service.DeleteAudio(context.Background(), "123")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to publish deleted"))

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
	publishAdapter.AssertExpectations(t)
}

func TestAudioService_DeleteAudio_Success(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(audioRepoMock)
	publishAdapter := new(audioPublishAdapterMock)

	service := NewAudioService(tm, repo, publishAdapter)

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
	repo.On("DeleteAudio", mock.Anything, "123").Return(nil)
	publishAdapter.On("PublishDeleted", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		payload := args.Get(1).(adapters.AudioPublishPayload)
		assert.Equal(t, "123", payload.ID)
		assert.Equal(t, "deleted", payload.Status)
		assert.Equal(t, int64(2), payload.Version)
	}).Return(nil)

	err := service.DeleteAudio(context.Background(), "123")
	assert.NoError(t, err)

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
	publishAdapter.AssertExpectations(t)
}
