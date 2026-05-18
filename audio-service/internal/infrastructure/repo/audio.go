package repo

import (
	"fmt"
	"time"

	"github.com/jtenhave/not-just-noise/audio-service/internal/audio"
)

type audioRepo struct {
	db DB
}

type AudioDBRow struct {
	ID        string    `db:"id"`
	Title     string    `db:"title"`
	Creator   string    `db:"creator"`
	FileURL   string    `db:"file_url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewAudioRepo(db DB) *audioRepo {
	return &audioRepo{
		db: db,
	}
}

func (repo *audioRepo) GetAudio(id string) (audio.Audio, error) {
	var dbRows []AudioDBRow
	err := repo.db.Select(&dbRows, "SELECT id, title, creator, file_url, created_at, updated_at FROM audio WHERE id = ?", id)
	if err != nil {
		return audio.Audio{}, fmt.Errorf("audiorepo: failed to get audio: %w", err)
	}

	if len(dbRows) == 0 {
		return audio.Audio{}, fmt.Errorf("audiorepo: audio not found")
	}

	return dbRows[0].ToAudio(), nil
}

func (repo *audioRepo) CreateAudio(a audio.Audio) error {
	dbRow := toAudioDBRow(a)
	err := repo.db.NamedExec(&dbRow, "INSERT INTO audio (id, title, creator, file_url) VALUES (:id, :title, :creator, :file_url)")
	if err != nil {
		return fmt.Errorf("audiorepo: failed to create audio: %w", err)
	}

	return nil
}

func (a AudioDBRow) ToAudio() audio.Audio {
	return audio.Audio{
		ID:        a.ID,
		Title:     a.Title,
		Creator:   a.Creator,
		FileURL:   a.FileURL,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

func toAudioDBRow(a audio.Audio) AudioDBRow {
	return AudioDBRow{
		ID:      a.ID,
		Title:   a.Title,
		Creator: a.Creator,
		FileURL: a.FileURL,
	}
}
