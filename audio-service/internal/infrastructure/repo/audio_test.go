package repo

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jtenhave/not-just-noise/audio-service/internal/audio"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type dbMock struct {
	mock.Mock
}

func (m *dbMock) ReadQuery(query string, dest interface{}, args ...interface{}) error {
	a := m.Called(dest, query, args)
	return a.Error(0)
}

func (m *dbMock) WriteQuery(query string, source interface{}) (int64, error) {
	args := m.Called(source, query)
	return args.Get(0).(int64), args.Error(1)
}

func TestGetAudioByID_NoAudioFound(t *testing.T) {
	db := new(dbMock)
	audioRepo := NewAudioRepo(db)

	db.On("ReadQuery", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	_, err := audioRepo.GetAudio("123abc")
	assert.Error(t, err)
	assert.True(t, njnerror.Type(err) == njnerror.NotFound)
}

func TestGetAudioByID_Failure(t *testing.T) {
	db := new(dbMock)
	audioRepo := NewAudioRepo(db)

	db.On("ReadQuery", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to get audio"))

	_, err := audioRepo.GetAudio("123abc")
	assert.Error(t, err)
}

func TestGetAudioByID_Success(t *testing.T) {
	db := new(dbMock)
	audioRepo := NewAudioRepo(db)

	audioRow := AudioDBRow{
		ID:        "123abc",
		CreatorID: "456def",
		Title:     "Test Audio",
		FileURL:   "https://test.com/audio.mp3",
	}

	db.On("ReadQuery", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			dest := args.Get(0).(*[]AudioDBRow)
			*dest = []AudioDBRow{audioRow}
		}).
		Return(nil)

	_, err := audioRepo.GetAudio("123abc")
	assert.NoError(t, err)
	assert.Equal(t, "123abc", audioRow.ID)
	assert.Equal(t, "456def", audioRow.CreatorID)
	assert.Equal(t, "Test Audio", audioRow.Title)
	assert.Equal(t, "https://test.com/audio.mp3", audioRow.FileURL)
}
func TestCreateAudio_Failure(t *testing.T) {
	db := new(dbMock)
	audioRepo := NewAudioRepo(db)

	db.On("WriteQuery", mock.Anything, mock.Anything).Return(int64(0), fmt.Errorf("failed to create audio"))

	err := audioRepo.CreateAudio(audio.Audio{CreatorID: "456def", Title: "Test Audio", FileURL: "https://test.com/audio.mp3"})
	assert.Error(t, err)
}

func TestCreateAudio_Success(t *testing.T) {
	db := new(dbMock)
	audioRepo := NewAudioRepo(db)

	audio := audio.Audio{
		CreatorID: "456def",
		Title:     "Test Audio",
		FileURL:   "https://test.com/audio.mp3",
	}

	var insertedAudioRow *AudioDBRow

	db.On("WriteQuery", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			insertedAudioRow = args.Get(0).(*AudioDBRow)
		}).
		Return(int64(1), nil)

	err := audioRepo.CreateAudio(audio)
	assert.NoError(t, err)
	assert.Equal(t, audio.CreatorID, insertedAudioRow.CreatorID)
	assert.Equal(t, audio.Title, insertedAudioRow.Title)
	assert.Equal(t, audio.FileURL, insertedAudioRow.FileURL)
}

func TestUpdateAudio_NoOperation(t *testing.T) {
	db := new(dbMock)
	audioRepo := NewAudioRepo(db)

	db.On("WriteQuery", mock.Anything, mock.Anything).Return(nil)

	err := audioRepo.UpdateAudio(audio.UpdateAudio{ID: "123abc", Title: nil, FileURL: nil})
	assert.NoError(t, err)
}

func TestUpdateAudio_Failure(t *testing.T) {
	db := new(dbMock)
	audioRepo := NewAudioRepo(db)

	audioRow := AudioDBRow{
		ID:        "123abc",
		CreatorID: "456def",
		Title:     "Test Audio",
		FileURL:   "https://test.com/audio.mp3",
	}

	db.On("ReadQuery", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			dest := args.Get(0).(*[]AudioDBRow)
			*dest = []AudioDBRow{audioRow}
		}).
		Return(nil)

	db.On("WriteQuery", mock.Anything, mock.Anything).Return(int64(0), fmt.Errorf("failed to update audio"))

	title := "New Title"
	err := audioRepo.UpdateAudio(audio.UpdateAudio{ID: "123abc", Title: &title, FileURL: nil})
	assert.Error(t, err)
}

func TestUpdateAudio_NoRowsAffected(t *testing.T) {
	db := new(dbMock)
	audioRepo := NewAudioRepo(db)

	audioRow := AudioDBRow{
		ID:        "123abc",
		CreatorID: "456def",
		Title:     "Test Audio",
		FileURL:   "https://test.com/audio.mp3",
	}

	db.On("ReadQuery", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			dest := args.Get(0).(*[]AudioDBRow)
			*dest = []AudioDBRow{audioRow}
		}).
		Return(nil)

	db.On("WriteQuery", mock.Anything, mock.Anything).Return(int64(0), nil)

	title := "New Title"
	err := audioRepo.UpdateAudio(audio.UpdateAudio{ID: "123abc", Title: &title, FileURL: nil})
	assert.Error(t, err)
}

func TestUpdateAudio_PartialUpdate(t *testing.T) {
	db := new(dbMock)
	audioRepo := NewAudioRepo(db)

	audioRow := AudioDBRow{
		ID:        "123abc",
		CreatorID: "456def",
		Title:     "Test Audio",
		FileURL:   "https://test.com/audio.mp3",
	}

	db.On("ReadQuery", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			dest := args.Get(0).(*[]AudioDBRow)
			*dest = []AudioDBRow{audioRow}
		}).
		Return(nil)

	var query string
	db.On("WriteQuery", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			query = args.Get(1).(string)
		}).
		Return(int64(1), nil)

	title := "New Title"
	err := audioRepo.UpdateAudio(audio.UpdateAudio{ID: "123abc", Title: &title, FileURL: nil})
	assert.NoError(t, err)
	assert.True(t, strings.Contains(query, "title = :title"))
	assert.False(t, strings.Contains(query, "file_url = :file_url"))
}

func TestDeleteAudio_Failure(t *testing.T) {
	db := new(dbMock)
	audioRepo := NewAudioRepo(db)

	audioRow := AudioDBRow{
		ID:        "123abc",
		CreatorID: "456def",
		Title:     "Test Audio",
		FileURL:   "https://test.com/audio.mp3",
	}

	db.On("ReadQuery", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			dest := args.Get(0).(*[]AudioDBRow)
			*dest = []AudioDBRow{audioRow}
		}).
		Return(nil)

	db.On("WriteQuery", mock.Anything, mock.Anything).Return(int64(0), fmt.Errorf("failed to delete audio"))

	err := audioRepo.DeleteAudio("123abc")
	assert.Error(t, err)
}

func TestDeleteAudio_Success(t *testing.T) {
	db := new(dbMock)
	audioRepo := NewAudioRepo(db)

	audioRow := AudioDBRow{
		ID:        "123abc",
		CreatorID: "456def",
		Title:     "Test Audio",
		FileURL:   "https://test.com/audio.mp3",
	}

	db.On("ReadQuery", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			dest := args.Get(0).(*[]AudioDBRow)
			*dest = []AudioDBRow{audioRow}
		}).
		Return(nil)

	db.On("WriteQuery", mock.Anything, mock.Anything).Return(int64(1), nil)

	err := audioRepo.DeleteAudio("123abc")
	assert.NoError(t, err)
}
