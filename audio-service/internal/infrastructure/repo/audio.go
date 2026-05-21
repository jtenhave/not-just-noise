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
	Version   int64     `db:"version"`
	Status    string    `db:"status"`
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
	err := repo.db.ReadQuery("SELECT id, title, creator_id, file_url, version, status, created_at, updated_at FROM audio WHERE id = ? AND status = 'active'", &dbRows, id)
	if err != nil {
		return audio.Audio{}, njnerror.Wrapf("audiorepo.GetAudio: failed to get audio: %w", err)
	}

	if len(dbRows) == 0 {
		return audio.Audio{}, njnerror.NewNJNError(njnerror.NotFound, "audiorepo.GetAudio: audio not found")
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
		Version:   a.Version,
		Status:    a.Status,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

// CreateAudio creates a new audio record using the given audio. Returns the first error encountered.
func (repo *audioRepo) CreateAudio(audio audio.Audio) error {
	dbRow := toAudioDBRow(audio)
	_, err := repo.db.WriteQuery("INSERT INTO audio (id, title, creator_id, file_url) VALUES (:id, :title, :creator_id, :file_url)", &dbRow)
	if err != nil {
		return njnerror.Wrapf("audiorepo.CreateAudio: failed to create audio: %w", err)
	}

	return nil
}

// ToAudioDBRow converts the given audio to an AudioDBRow.
func toAudioDBRow(audio audio.Audio) AudioDBRow {
	return AudioDBRow{
		ID:        audio.ID,
		CreatorID: audio.CreatorID,
		Title:     audio.Title,
		FileURL:   audio.FileURL,
	}
}

// UpdateAudio updates an audio record using the given audio. Returns the first error encountered.
func (repo *audioRepo) UpdateAudio(audio audio.UpdateAudio) error {
	if audio.Title == nil && audio.FileURL == nil {
		return nil
	}

	storedAudio, err := repo.GetAudio(audio.ID)
	if err != nil {
		return njnerror.Wrapf("audiorepo.UpdateAudio: failed to get audio: %w", err)
	}

	dbRow := toUpdateAudioDBRow(audio, storedAudio)

	query := "UPDATE audio SET "
	updates := make([]string, 0)
	if audio.Title != nil {
		updates = append(updates, "title = :title")
	}
	if audio.FileURL != nil {
		updates = append(updates, "file_url = :file_url")
	}

	updates = append(updates, "version = version + 1")
	query += strings.Join(updates, ", ") + " WHERE id = :id AND version = :version AND status = 'active'"

	rowsAffected, err := repo.db.WriteQuery(query, &dbRow)
	if err != nil {
		return njnerror.Wrapf("audiorepo.UpdateAudio: failed to update audio: %w", err)
	}

	if rowsAffected == 0 {
		return njnerror.NewNJNError(njnerror.Conflict, "audiorepo.UpdateAudio: audio version out of date")
	}

	return nil
}

// toUpdateAudioDBRow converts the given audio to an AudioDBRow, using storedAudio.
func toUpdateAudioDBRow(audio audio.UpdateAudio, storedAudio audio.Audio) AudioDBRow {
	title := ""
	if audio.Title != nil {
		title = *audio.Title
	}

	fileURL := ""
	if audio.FileURL != nil {
		fileURL = *audio.FileURL
	}

	return AudioDBRow{
		ID:      audio.ID,
		Title:   title,
		FileURL: fileURL,
		Version: storedAudio.Version,
	}
}

// DeleteAudio deletes an audio record using the given id. Returns the first error encountered.
func (repo *audioRepo) DeleteAudio(id string) error {
	_, err := repo.GetAudio(id)
	if err != nil {
		return njnerror.Wrapf("audiorepo.DeleteAudio: failed to get audio: %w", err)
	}

	dbRow := AudioDBRow{
		ID:     id,
		Status: "deleted",
	}

	// Delete should always succeed, so we don't check for version in this case.
	query := `UPDATE audio SET 
		version = version + 1,
		status = 'deleted'
		WHERE id = :id`

	_, err = repo.db.WriteQuery(query, &dbRow)
	if err != nil {
		return njnerror.Wrapf("audiorepo.DeleteAudio: failed to delete audio: %w", err)
	}

	return nil
}
