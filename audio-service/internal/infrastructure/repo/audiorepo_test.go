package repo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jtenhave/not-just-noise/audio-service/internal/audio"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type queryRunnerMock struct {
	mock.Mock
}

func (m *queryRunnerMock) ExecContext(ctx context.Context, query string, args ...any) (int64, error) {
	callArgs := m.Called(append([]any{ctx, query}, args...)...)
	return callArgs.Get(0).(int64), callArgs.Error(1)
}

func (m *queryRunnerMock) QueryContext(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	callArgs := m.Called(append([]any{ctx, query}, args...)...)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).([]map[string]any), callArgs.Error(1)
}

func TestToAudio(t *testing.T) {
	now := time.Now()
	row := map[string]any{
		"id":         "123abc",
		"title":      "Test Audio",
		"creator_id": "456def",
		"file_url":   "https://test.com/audio.mp3",
		"version":    int64(2),
		"status":     "active",
		"created_at": now,
		"updated_at": now,
	}

	got := toAudio(row)

	assert.Equal(t, "123abc", got.ID)
	assert.Equal(t, "456def", got.CreatorID)
	assert.Equal(t, "Test Audio", got.Title)
	assert.Equal(t, "https://test.com/audio.mp3", got.FileURL)
	assert.Equal(t, int64(2), got.Version)
	assert.Equal(t, "active", got.Status)
	assert.Equal(t, now, got.CreatedAt)
	assert.Equal(t, now, got.UpdatedAt)
}

func TestGetAudio_Success(t *testing.T) {
	db := new(queryRunnerMock)
	repo := NewAudioRepo(db)
	ctx := context.Background()
	now := time.Now()

	row := map[string]any{
		"id":         "123abc",
		"title":      "Test Audio",
		"creator_id": "456def",
		"file_url":   "https://test.com/audio.mp3",
		"version":    int64(1),
		"status":     "active",
		"created_at": now,
		"updated_at": now,
	}

	db.On("QueryContext", ctx, mock.AnythingOfType("string"), "123abc").
		Return([]map[string]any{row}, nil)

	got, err := repo.GetAudio(ctx, "123abc")

	assert.NoError(t, err)
	assert.Equal(t, toAudio(row), got)
	db.AssertExpectations(t)
}

func TestGetAudio_NotFound(t *testing.T) {
	db := new(queryRunnerMock)
	repo := NewAudioRepo(db)
	ctx := context.Background()

	db.On("QueryContext", ctx, mock.AnythingOfType("string"), "123abc").
		Return([]map[string]any{}, nil)

	_, err := repo.GetAudio(ctx, "123abc")

	assert.Error(t, err)
	assert.Equal(t, njnerror.NotFound, njnerror.Type(err))
	db.AssertExpectations(t)
}

func TestGetAudio_QueryError(t *testing.T) {
	db := new(queryRunnerMock)
	repo := NewAudioRepo(db)
	ctx := context.Background()

	db.On("QueryContext", ctx, mock.AnythingOfType("string"), "123abc").
		Return(nil, fmt.Errorf("database error"))

	_, err := repo.GetAudio(ctx, "123abc")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audiorepo.GetAudio: failed to get audio")
	db.AssertExpectations(t)
}

func TestCreateAudio_Success(t *testing.T) {
	db := new(queryRunnerMock)
	repo := NewAudioRepo(db)
	ctx := context.Background()

	audioRec := audio.Audio{
		ID:        "123abc",
		CreatorID: "456def",
		Title:     "Test Audio",
		FileURL:   "https://test.com/audio.mp3",
	}

	db.On("ExecContext", ctx, mock.AnythingOfType("string"), "123abc", "Test Audio", "456def", "https://test.com/audio.mp3").
		Return(int64(1), nil)

	err := repo.CreateAudio(ctx, audioRec)

	assert.NoError(t, err)
	db.AssertExpectations(t)
}

func TestCreateAudio_Error(t *testing.T) {
	db := new(queryRunnerMock)
	repo := NewAudioRepo(db)
	ctx := context.Background()

	audioRec := audio.Audio{
		ID:        "123abc",
		CreatorID: "456def",
		Title:     "Test Audio",
		FileURL:   "https://test.com/audio.mp3",
	}

	db.On("ExecContext", ctx, mock.AnythingOfType("string"), "123abc", "Test Audio", "456def", "https://test.com/audio.mp3").
		Return(int64(0), fmt.Errorf("database error"))

	err := repo.CreateAudio(ctx, audioRec)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audiorepo.CreateAudio: failed to create audio")
	db.AssertExpectations(t)
}

func TestUpdateAudio_NoFields(t *testing.T) {
	db := new(queryRunnerMock)
	repo := NewAudioRepo(db)
	ctx := context.Background()

	err := repo.UpdateAudio(ctx, audio.UpdateAudio{ID: "123abc"}, 1)

	assert.NoError(t, err)
	db.AssertNotCalled(t, "ExecContext")
}

func TestUpdateAudio_TitleOnly_Success(t *testing.T) {
	db := new(queryRunnerMock)
	repo := NewAudioRepo(db)
	ctx := context.Background()
	title := "New Title"

	expectedQuery := "UPDATE audio SET title = ?, version = version + 1 WHERE id = ? AND version = ? AND status = 'active'"
	db.On("ExecContext", ctx, expectedQuery, title, "123abc", int64(1)).
		Return(int64(1), nil)

	err := repo.UpdateAudio(ctx, audio.UpdateAudio{ID: "123abc", Title: &title}, 1)

	assert.NoError(t, err)
	db.AssertExpectations(t)
}

func TestUpdateAudio_FileURLOnly_Success(t *testing.T) {
	db := new(queryRunnerMock)
	repo := NewAudioRepo(db)
	ctx := context.Background()
	fileURL := "https://test.com/new.mp3"

	expectedQuery := "UPDATE audio SET file_url = ?, version = version + 1 WHERE id = ? AND version = ? AND status = 'active'"
	db.On("ExecContext", ctx, expectedQuery, fileURL, "123abc", int64(2)).
		Return(int64(1), nil)

	err := repo.UpdateAudio(ctx, audio.UpdateAudio{ID: "123abc", FileURL: &fileURL}, 2)

	assert.NoError(t, err)
	db.AssertExpectations(t)
}

func TestUpdateAudio_TitleAndFileURL_Success(t *testing.T) {
	db := new(queryRunnerMock)
	repo := NewAudioRepo(db)
	ctx := context.Background()
	title := "New Title"
	fileURL := "https://test.com/new.mp3"

	expectedQuery := "UPDATE audio SET title = ?, file_url = ?, version = version + 1 WHERE id = ? AND version = ? AND status = 'active'"
	db.On("ExecContext", ctx, expectedQuery, title, fileURL, "123abc", int64(3)).
		Return(int64(1), nil)

	err := repo.UpdateAudio(ctx, audio.UpdateAudio{ID: "123abc", Title: &title, FileURL: &fileURL}, 3)

	assert.NoError(t, err)
	db.AssertExpectations(t)
}

func TestUpdateAudio_Conflict(t *testing.T) {
	db := new(queryRunnerMock)
	repo := NewAudioRepo(db)
	ctx := context.Background()
	title := "New Title"

	db.On("ExecContext", ctx, mock.AnythingOfType("string"), title, "123abc", int64(1)).
		Return(int64(0), nil)

	err := repo.UpdateAudio(ctx, audio.UpdateAudio{ID: "123abc", Title: &title}, 1)

	assert.Error(t, err)
	assert.Equal(t, njnerror.Conflict, njnerror.Type(err))
	db.AssertExpectations(t)
}

func TestUpdateAudio_Error(t *testing.T) {
	db := new(queryRunnerMock)
	repo := NewAudioRepo(db)
	ctx := context.Background()
	title := "New Title"

	db.On("ExecContext", ctx, mock.AnythingOfType("string"), title, "123abc", int64(1)).
		Return(int64(0), fmt.Errorf("database error"))

	err := repo.UpdateAudio(ctx, audio.UpdateAudio{ID: "123abc", Title: &title}, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audiorepo.UpdateAudio: failed to update audio")
	db.AssertExpectations(t)
}

func TestDeleteAudio_Success(t *testing.T) {
	db := new(queryRunnerMock)
	repo := NewAudioRepo(db)
	ctx := context.Background()

	expectedQuery := `UPDATE audio SET 
		version = version + 1,
		status = 'deleted'
		WHERE id = ?`
	db.On("ExecContext", ctx, expectedQuery, "123abc").
		Return(int64(1), nil)

	err := repo.DeleteAudio(ctx, "123abc")

	assert.NoError(t, err)
	db.AssertExpectations(t)
}

func TestDeleteAudio_Error(t *testing.T) {
	db := new(queryRunnerMock)
	repo := NewAudioRepo(db)
	ctx := context.Background()

	db.On("ExecContext", ctx, mock.AnythingOfType("string"), "123abc").
		Return(int64(0), fmt.Errorf("database error"))

	err := repo.DeleteAudio(ctx, "123abc")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audiorepo.DeleteAudio: failed to delete audio")
	db.AssertExpectations(t)
}
