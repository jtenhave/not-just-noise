package repo

import (
	"strings"
	"time"

	"github.com/jtenhave/not-just-noise/audio-service/internal/audio"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type audioRepo struct {
	db DB
}

type AudioDBRow struct {
	ID        string    `db:"id"`
	CreatorID string    `db:"creator_id"`
	Title     string    `db:"title"`
	FileURL   string    `db:"file_url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// NewAudioRepo creates a new audio repository using the given database.
func NewAudioRepo(db DB) *audioRepo {
	return &audioRepo{
		db: db,
	}
}

// GetAudio gets an audio record using the given id. Returns the audio record and the first error encountered.
func (repo *audioRepo) GetAudio(id string) (audio.Audio, error) {
	var dbRows []AudioDBRow
	err := repo.db.ReadQuery("SELECT id, title, creator_id, file_url, created_at, updated_at FROM audio WHERE id = ?", &dbRows, id)
	if err != nil {
		return audio.Audio{}, njnerror.Wrapf("audiorepo.GetAudioByID: failed to get audio: %w", err)
	}

	if len(dbRows) == 0 {
		return audio.Audio{}, njnerror.NewNJNError(njnerror.NotFound, "audiorepo.GetAudioByID: audio not found")
	}

	return dbRows[0].ToAudio(), nil
}

// ToAudio converts the AudioDBRow a to an Audio.
func (a AudioDBRow) ToAudio() audio.Audio {
	return audio.Audio{
		ID:        a.ID,
		CreatorID: a.CreatorID,
		Title:     a.Title,
		FileURL:   a.FileURL,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

// CreateAudio creates a new audio record using the given audio a. Returns the first error encountered.
func (repo *audioRepo) CreateAudio(a audio.Audio) error {
	dbRow := toAudioDBRow(a)
	err := repo.db.WriteQuery("INSERT INTO audio (id, title, creator_id, file_url) VALUES (:id, :title, :creator_id, :file_url)", &dbRow)
	if err != nil {
		return njnerror.Wrapf("audiorepo.CreateAudio: failed to create audio: %w", err)
	}

	return nil
}

// ToAudioDBRow converts the Audio a to an AudioDBRow.
func toAudioDBRow(a audio.Audio) AudioDBRow {
	return AudioDBRow{
		ID:        a.ID,
		CreatorID: a.CreatorID,
		Title:     a.Title,
		FileURL:   a.FileURL,
	}
}

func (repo *audioRepo) UpdateAudio(audio audio.UpdateAudio) error {
	if audio.Title == nil && audio.FileURL == nil {
		return nil
	}

	dbRow := toUpdateAudioDBRow(audio)
	query := "UPDATE audio SET "

	updates := make([]string, 0)
	if audio.Title != nil {
		updates = append(updates, "title = :title")
	}
	if audio.FileURL != nil {
		updates = append(updates, "file_url = :file_url")
	}
	query += strings.Join(updates, ", ") + " WHERE id = :id"

	err := repo.db.WriteQuery(query, &dbRow)
	if err != nil {
		return njnerror.Wrapf("audiorepo.UpdateAudio: failed to update audio: %w", err)
	}

	return nil
}

func toUpdateAudioDBRow(a audio.UpdateAudio) AudioDBRow {
	title := ""
	if a.Title != nil {
		title = *a.Title
	}

	fileURL := ""
	if a.FileURL != nil {
		fileURL = *a.FileURL
	}

	return AudioDBRow{
		ID:      a.ID,
		Title:   title,
		FileURL: fileURL,
	}
}

func (repo *audioRepo) DeleteAudio(id string) error {
	err := repo.db.WriteQuery("DELETE FROM audio WHERE id = :id", &AudioDBRow{ID: id})
	if err != nil {
		return njnerror.Wrapf("audiorepo.DeleteAudio: failed to delete audio: %w", err)
	}

	return nil
}
