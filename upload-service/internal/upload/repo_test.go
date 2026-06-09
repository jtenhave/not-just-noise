package upload

import (
	"context"
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

func TestUploadRepo_GetUpload_NotFound(t *testing.T) {
	db := new(mockDB)
	repo := NewUploadRepo(db)

	db.On("QueryContext", mock.Anything, mock.Anything, mock.Anything).Return([]map[string]any{}, nil)

	_, err := repo.GetUpload(context.Background(), "123")
	assert.Error(t, err)
	assert.Equal(t, njnerror.NotFound, njnerror.Type(err))
}

func TestUploadRepo_GetUpload_Success(t *testing.T) {
	db := new(mockDB)
	repo := NewUploadRepo(db)

	now := time.Now()
	expectedValues := []map[string]any{
		{
			"id":         "123",
			"audio_id":   "456",
			"file_url":   "https://test.com/test.mp3",
			"file_hash":  "123abc",
			"version":    int64(1),
			"status":     "active",
			"created_at": now,
			"updated_at": now,
		},
	}

	db.On("QueryContext", mock.Anything, mock.Anything, mock.Anything).Return(expectedValues, nil)

	upload, err := repo.GetUpload(context.Background(), "123")
	assert.NoError(t, err)
	assert.Equal(t, "123", upload.ID)
	assert.Equal(t, "456", upload.AudioID)
	assert.Equal(t, "https://test.com/test.mp3", upload.FileURL)
	assert.Equal(t, "123abc", upload.FileHash)
	assert.Equal(t, int64(1), upload.Version)
	assert.Equal(t, "active", upload.Status)
	assert.Equal(t, now, upload.CreatedAt)
	assert.Equal(t, now, upload.UpdatedAt)
}

func TestUploadRepo_UpdateUpload_NoChanges(t *testing.T) {
	db := new(mockDB)
	repo := NewUploadRepo(db)

	db.On("ExecContext", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), nil)

	err := repo.UpdateUpload(context.Background(), Upload{
		ID:      "123",
		Version: 1,
	})

	assert.NoError(t, err)
	db.AssertNotCalled(t, "ExecContext")
}

func TestAudioRepo_UpdateAudio_NoRowsAffected(t *testing.T) {
	db := new(mockDB)
	repo := NewUploadRepo(db)

	db.On("ExecContext", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), nil)

	err := repo.UpdateUpload(context.Background(), Upload{
		ID:      "123",
		Version: 1,
	})

	assert.Error(t, err)
	assert.Equal(t, njnerror.Conflict, njnerror.Type(err))
}
