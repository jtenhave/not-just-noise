package repo

import (
	"fmt"
	"testing"

	"github.com/jtenhave/not-just-noise/audio-service/internal/audio"
	"github.com/jtenhave/not-just-noise/lib/errorcode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type dbMock struct {
	mock.Mock
}

func (m *dbMock) Select(dest interface{}, query string, args ...interface{}) error {
	a := m.Called(dest, query, args)
	return a.Error(0)
}

func (m *dbMock) NamedExec(source interface{}, query string) error {
	args := m.Called(source, query)
	return args.Error(0)
}

func TestGetAudioByID_NoAudioFound(t *testing.T) {
	db := new(dbMock)
	audioRepo := NewAudioRepo(db)

	db.On("Select", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	_, err := audioRepo.GetAudioByID("123abc")
	assert.Error(t, err)
	assert.True(t, errorcode.ErrorCode(err) == errorcode.NotFound)
}

func TestGetAudioByID_Failure(t *testing.T) {
	db := new(dbMock)
	audioRepo := NewAudioRepo(db)

	db.On("Select", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to get audio"))

	_, err := audioRepo.GetAudioByID("123abc")
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

	db.On("Select", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			dest := args.Get(0).(*[]AudioDBRow)
			*dest = []AudioDBRow{audioRow}
		}).
		Return(nil)

	_, err := audioRepo.GetAudioByID("123abc")
	assert.NoError(t, err)
	assert.Equal(t, "123abc", audioRow.ID)
	assert.Equal(t, "456def", audioRow.CreatorID)
	assert.Equal(t, "Test Audio", audioRow.Title)
	assert.Equal(t, "https://test.com/audio.mp3", audioRow.FileURL)
}

func TestGetAudioByCreatorIDAndTitle_NoAudioFound(t *testing.T) {
	db := new(dbMock)
	audioRepo := NewAudioRepo(db)

	db.On("Select", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	_, err := audioRepo.GetAudioByCreatorIDAndTitle("456def", "Test Audio")
	assert.Error(t, err)
	assert.True(t, errorcode.ErrorCode(err) == errorcode.NotFound)
}

func TestGetAudioByCreatorIDAndTitle_Failure(t *testing.T) {
	db := new(dbMock)
	audioRepo := NewAudioRepo(db)

	db.On("Select", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to get audio"))

	_, err := audioRepo.GetAudioByCreatorIDAndTitle("456def", "Test Audio")
	assert.Error(t, err)
}

func TestGetAudioByCreatorIDAndTitle_Success(t *testing.T) {
	db := new(dbMock)
	audioRepo := NewAudioRepo(db)

	audioRow := AudioDBRow{
		ID:        "123abc",
		CreatorID: "456def",
		Title:     "Test Audio",
		FileURL:   "https://test.com/audio.mp3",
	}

	db.On("Select", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			dest := args.Get(0).(*[]AudioDBRow)
			*dest = []AudioDBRow{audioRow}
		}).
		Return(nil)

	_, err := audioRepo.GetAudioByCreatorIDAndTitle("456def", "Test Audio")
	assert.NoError(t, err)
	assert.Equal(t, "123abc", audioRow.ID)
	assert.Equal(t, "456def", audioRow.CreatorID)
	assert.Equal(t, "Test Audio", audioRow.Title)
	assert.Equal(t, "https://test.com/audio.mp3", audioRow.FileURL)
}

func TestCreateAudio_Failure(t *testing.T) {
	db := new(dbMock)
	audioRepo := NewAudioRepo(db)

	db.On("NamedExec", mock.Anything, mock.Anything).Return(fmt.Errorf("failed to create audio"))

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

	db.On("NamedExec", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			insertedAudioRow = args.Get(0).(*AudioDBRow)
		}).
		Return(nil)

	err := audioRepo.CreateAudio(audio)
	assert.NoError(t, err)
	assert.Equal(t, audio.CreatorID, insertedAudioRow.CreatorID)
	assert.Equal(t, audio.Title, insertedAudioRow.Title)
	assert.Equal(t, audio.FileURL, insertedAudioRow.FileURL)
}
