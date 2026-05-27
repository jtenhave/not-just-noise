package audio

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jtenhave/not-just-noise/lib/njnerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDB struct {
	mock.Mock
}

func (m *mockDB) QueryContext(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	arguments := m.Called(ctx, query, args)

	var err error
	if e := arguments.Get(1); e != nil {
		err = e.(error)
	}

	return arguments.Get(0).([]map[string]any), err
}

func (m *mockDB) ExecContext(ctx context.Context, query string, args ...any) (int64, error) {
	arguments := m.Called(ctx, query, args)

	var err error
	if e := arguments.Get(1); e != nil {
		err = e.(error)
	}

	return arguments.Get(0).(int64), err
}

func TestAudioRepo_GetAudio_NotFound(t *testing.T) {
	db := new(mockDB)
	repo := NewAudioRepo(db)

	db.On("QueryContext", mock.Anything, mock.Anything, mock.Anything).Return([]map[string]any{}, nil)

	_, err := repo.GetAudio(context.Background(), "123")
	assert.Error(t, err)
	assert.Equal(t, njnerror.NotFound, njnerror.Type(err))
}

func TestAudioRepo_GetAudio_Success(t *testing.T) {
	db := new(mockDB)
	repo := NewAudioRepo(db)

	now := time.Now()
	expectedValues := []map[string]any{
		{
			"id":         "123",
			"title":      "Test Title",
			"creator_id": "123",
			"file_url":   "https://test.com/test.mp3",
			"version":    int64(1),
			"status":     "active",
			"created_at": now,
			"updated_at": now,
		},
	}

	db.On("QueryContext", mock.Anything, mock.Anything, mock.Anything).Return(expectedValues, nil)

	audio, err := repo.GetAudio(context.Background(), "123")
	assert.NoError(t, err)
	assert.Equal(t, "123", audio.ID)
	assert.Equal(t, "Test Title", *audio.Title)
	assert.Equal(t, "123", audio.CreatorID)
	assert.Equal(t, "https://test.com/test.mp3", *audio.FileURL)
	assert.Equal(t, int64(1), audio.Version)
	assert.Equal(t, "active", audio.Status)
	assert.Equal(t, now, audio.CreatedAt)
	assert.Equal(t, now, audio.UpdatedAt)
}

func TestAudioRepo_UpdateAudio_NoChanges(t *testing.T) {
	db := new(mockDB)
	repo := NewAudioRepo(db)

	db.On("ExecContext", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), nil)

	err := repo.UpdateAudio(context.Background(), Audio{
		ID:      "123",
		Version: 1,
	})

	assert.NoError(t, err)
	db.AssertNotCalled(t, "ExecContext")
}

func TestAudioRepo_UpdateAudio_TitleChange(t *testing.T) {
	db := new(mockDB)
	repo := NewAudioRepo(db)

	db.On("ExecContext", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		query := args.Get(1).(string)
		assert.True(t, strings.Contains(query, "title ="))
		assert.False(t, strings.Contains(query, "file_url ="))
	}).Return(int64(1), nil)

	newTitle := "New Title"
	err := repo.UpdateAudio(context.Background(), Audio{
		ID:      "123",
		Version: 1,
		Title:   &newTitle,
	})

	assert.NoError(t, err)
	db.AssertExpectations(t)
}

func TestAudioRepo_UpdateAudio_FileURLChange(t *testing.T) {
	db := new(mockDB)
	repo := NewAudioRepo(db)

	db.On("ExecContext", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		query := args.Get(1).(string)
		assert.False(t, strings.Contains(query, "title ="))
		assert.True(t, strings.Contains(query, "file_url ="))
	}).Return(int64(1), nil)

	newFileURL := "https://test.com/new.mp3"
	err := repo.UpdateAudio(context.Background(), Audio{
		ID:      "123",
		Version: 1,
		FileURL: &newFileURL,
	})

	assert.NoError(t, err)
	db.AssertExpectations(t)
}

func TestAudioRepo_UpdateAudio_NoRowsAffected(t *testing.T) {
	db := new(mockDB)
	repo := NewAudioRepo(db)

	db.On("ExecContext", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), nil)

	newTitle := "New Title"
	err := repo.UpdateAudio(context.Background(), Audio{
		ID:      "123",
		Version: 1,
		Title:   &newTitle,
	})

	assert.Error(t, err)
	assert.Equal(t, njnerror.Conflict, njnerror.Type(err))
}
