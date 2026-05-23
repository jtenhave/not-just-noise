package audio

import (
	"context"
	"strings"
	"time"

	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type db interface {
	ExecContext(ctx context.Context, query string, args ...any) (int64, error)
	QueryContext(ctx context.Context, query string, args ...any) ([]map[string]any, error)
}

type audioRepo struct {
	db db
}

// NewAudioRepo creates a new audio repository using the given database.
func NewAudioRepo(db db) *audioRepo {
	return &audioRepo{
		db: db,
	}
}

// GetAudio gets an audio record using the given id. Returns the audio record and the first error encountered.
func (repo *audioRepo) GetAudio(ctx context.Context, id string) (Audio, error) {
	dbRows, err := repo.db.QueryContext(ctx,
		`SELECT id, title, creator_id, file_url, version, status, created_at, updated_at 
		 FROM audio 
		 WHERE id = ? AND status = 'active'`, id)

	if err != nil {
		return Audio{}, njnerror.Wrapf("audiorepo.GetAudio: failed to get audio: %w", err)
	}

	if len(dbRows) == 0 {
		return Audio{}, njnerror.NewNJNError(njnerror.NotFound, "audiorepo.GetAudio: audio not found")
	}

	return toAudio(dbRows[0]), nil
}

// ToAudio converts a row a to an Audio.
func toAudio(row map[string]any) Audio {
	title := row["title"].(string)
	fileURL := row["file_url"].(string)

	return Audio{
		ID:        row["id"].(string),
		CreatorID: row["creator_id"].(string),
		Title:     &title,
		FileURL:   &fileURL,
		Version:   row["version"].(int64),
		Status:    row["status"].(string),
		CreatedAt: row["created_at"].(time.Time),
		UpdatedAt: row["updated_at"].(time.Time),
	}
}

// CreateAudio creates a new audio record using the given audio. Returns the first error encountered.
func (repo *audioRepo) CreateAudio(ctx context.Context, audio Audio) error {
	_, err := repo.db.ExecContext(ctx,
		`INSERT INTO audio (id, title, creator_id, file_url) 
		VALUES (?, ?, ?, ?)`, audio.ID, audio.Title, audio.CreatorID, audio.FileURL)

	if err != nil {
		return njnerror.Wrapf("audiorepo.CreateAudio: failed to create audio: %w", err)
	}

	return nil
}

// UpdateAudio updates an audio record using the given audio. Returns the first error encountered.
func (repo *audioRepo) UpdateAudio(ctx context.Context, audio Audio, version int64) error {
	if audio.Title == nil && audio.FileURL == nil {
		return nil
	}

	query := "UPDATE audio SET "
	args := make([]any, 0)
	updates := make([]string, 0)

	if audio.Title != nil {
		updates = append(updates, "title = ?")
		args = append(args, *audio.Title)
	}

	if audio.FileURL != nil {
		updates = append(updates, "file_url = ?")
		args = append(args, *audio.FileURL)
	}

	args = append(args, audio.ID, version)

	updates = append(updates, "version = version + 1")
	query += strings.Join(updates, ", ") + " WHERE id = ? AND version = ? AND status = 'active'"

	rowsAffected, err := repo.db.ExecContext(ctx, query, args...)
	if err != nil {
		return njnerror.Wrapf("audiorepo.UpdateAudio: failed to update audio: %w", err)
	}

	if rowsAffected == 0 {
		return njnerror.NewNJNError(njnerror.Conflict, "audiorepo.UpdateAudio: audio record out of date")
	}

	return nil
}

// DeleteAudio deletes an audio record using the given id. Returns the first error encountered.
func (repo *audioRepo) DeleteAudio(ctx context.Context, id string) error {
	// Delete should always succeed, so we don't check for version in this case.
	query := `UPDATE audio SET 
		version = version + 1,
		status = 'deleted'
		WHERE id = ?`

	_, err := repo.db.ExecContext(ctx, query, id)
	if err != nil {
		return njnerror.Wrapf("audiorepo.DeleteAudio: failed to delete audio: %w", err)
	}

	return nil
}
