package upload

import (
	"context"
	"time"

	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type db interface {
	ExecContext(ctx context.Context, query string, args ...any) (int64, error)
	QueryContext(ctx context.Context, query string, args ...any) ([]map[string]any, error)
}

type uploadRepo struct {
	db db
}

// NewUploadRepo creates a new upload repository using the given database.
func NewUploadRepo(db db) *uploadRepo {
	return &uploadRepo{
		db: db,
	}
}

// GetUpload gets an upload record using the given audioID. Returns the upload record and the first error encountered.
func (repo *uploadRepo) GetUpload(ctx context.Context, audioID string) (Upload, error) {
	dbRows, err := repo.db.QueryContext(ctx,
		`SELECT id, audio_id, version, status, created_at, updated_at 
		 FROM uploads 
		 WHERE audio_id = ? AND status = 'active'`, audioID)

	if err != nil {
		return Upload{}, njnerror.Wrapf("uploadrepo.GetUpload: failed to get upload: %w", err)
	}

	if len(dbRows) == 0 {
		return Upload{}, njnerror.NewNJNError(njnerror.NotFound, "uploadrepo.GetUpload: upload not found")
	}

	return toUpload(dbRows[0]), nil
}

// ToUpload converts a row a to an Upload.
func toUpload(row map[string]any) Upload {
	return Upload{
		ID:        row["id"].(string),
		AudioID:   row["audio_id"].(string),
		Version:   row["version"].(int64),
		Status:    row["status"].(string),
		CreatedAt: row["created_at"].(time.Time),
		UpdatedAt: row["updated_at"].(time.Time),
	}
}

// CreateUpload creates a new upload record using the given upload. Returns the first error encountered.
func (repo *uploadRepo) CreateUpload(ctx context.Context, upload Upload) error {
	_, err := repo.db.ExecContext(ctx,
		`INSERT INTO uploads (id, audio_id, file_url, file_hash) 
		VALUES (?, ?, ?, ?)`, upload.ID, upload.AudioID, upload.FileURL, upload.FileHash)

	if err != nil {
		return njnerror.Wrapf("audiorepo.CreateAudio: failed to create audio: %w", err)
	}

	return nil
}

// UpdateUpload updates an upload record using the given upload. Returns the first error encountered.
func (repo *uploadRepo) UpdateUpload(ctx context.Context, upload Upload) error {
	query := `UPDATE uploads 
			  SET file_url = ?, file_hash = ?, version = ?
	          WHERE id = ? AND version < ? AND status = 'active'`

	rowsAffected, err := repo.db.ExecContext(ctx, query, upload.FileURL, upload.FileHash, upload.Version, upload.ID, upload.Version)
	if err != nil {
		return njnerror.Wrapf("uploadrepo.UpdateUpload: failed to update upload: %w", err)
	}

	if rowsAffected == 0 {
		return njnerror.NewNJNError(njnerror.Conflict, "uploadrepo.UpdateUpload: upload record out of date")
	}

	return nil
}

// DeleteUpload deletes an upload record using the given id. Returns the first error encountered.
func (repo *uploadRepo) DeleteUpload(ctx context.Context, id string, version int64) error {
	// Delete should always succeed, so we don't check for version in this case.
	query := `UPDATE uploads SET 
		version = ?,
		status = 'deleted'
		WHERE id = ?`

	_, err := repo.db.ExecContext(ctx, query, version, id)
	if err != nil {
		return njnerror.Wrapf("uploadrepo.DeleteUpload: failed to delete upload: %w", err)
	}

	return nil
}
