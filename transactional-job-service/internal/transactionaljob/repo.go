package transactionaljob

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type db interface {
	ExecContext(ctx context.Context, query string, args ...any) (int64, error)
	QueryContext(ctx context.Context, query string, args ...any) ([]map[string]any, error)
	IsTx(ctx context.Context) bool
}

type transactionalJobRepo struct {
	db db
}

// NewAudioRepo creates a new audio repository using the given database.
func NewTransactionalJobRepo(db db) *transactionalJobRepo {
	return &transactionalJobRepo{
		db: db,
	}
}

// GetAvailableTransactionalJobs gets the next available transactional jobs, up to the given limit. Returns the first available transactional jobs and the first error encountered.
func (repo *transactionalJobRepo) GetAvailableTransactionalJobs(ctx context.Context, limit int) ([]TransactionalJob, error) {
	if !repo.db.IsTx(ctx) {
		return nil, fmt.Errorf("transactionaljobrepo.GetAvailableTransactionalJobs: not in a transaction")
	}

	dbRows, err := repo.db.QueryContext(ctx,
		`SELECT id, callback_url, payload
		 FROM transactional_jobs
		 WHERE (processing_at IS NULL OR TIMESTAMPDIFF(SECOND, processing_at, NOW()) > claim_timeout) AND attempts < max_attempts AND available_at <= NOW()
		 ORDER BY created_at ASC
		 LIMIT ?
		 FOR UPDATE SKIP LOCKED`, limit)

	if err != nil {
		return nil, njnerror.Wrapf("transactionaljobrepo.GetAvailableTransactionalJobs: failed to get available transactional jobs: %w", err)
	}

	if len(dbRows) == 0 {
		return nil, nil
	}

	transactionalJobs := make([]TransactionalJob, len(dbRows))
	for i, row := range dbRows {
		transactionalJob, err := toTransactionalJob(row)
		if err != nil {
			return nil, njnerror.Wrapf("transactionaljobrepo.GetAvailableTransactionalJobs: failed to convert row to transactional job: %w", err)
		}
		transactionalJobs[i] = transactionalJob
	}
	return transactionalJobs, nil
}

// ToTransactionalJob converts a row a to an TransactionalJob.
func toTransactionalJob(row map[string]any) (TransactionalJob, error) {
	id, ok := row["id"].(string)
	if !ok {
		return TransactionalJob{}, fmt.Errorf("transactionaljobrepo.toTransactionalJob: id is not a string")
	}
	callbackURL, ok := row["callback_url"].(string)
	if !ok {
		return TransactionalJob{}, fmt.Errorf("transactionaljobrepo.toTransactionalJob: callbackURL is not a string")
	}
	payload, ok := row["payload"].(string)
	if !ok {
		return TransactionalJob{}, fmt.Errorf("transactionaljobrepo.toTransactionalJob: payload is not a string")
	}

	return TransactionalJob{
		ID:          id,
		CallbackURL: callbackURL,
		Payload:     payload,
	}, nil
}

// ClaimTransactionalJobs claims the given transactional jobs. Returns the first error encountered.
func (repo *transactionalJobRepo) ClaimTransactionalJobs(ctx context.Context, ids []string) error {
	if !repo.db.IsTx(ctx) {
		return fmt.Errorf("transactionaljobrepo.ClaimTransactionalJob: not in a transaction")
	}

	placeholders := make([]string, len(ids))
	for i := range ids {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(`UPDATE transactional_jobs 
		SET processing_at = ?, attempts = attempts + 1
		WHERE id in (%s)`, strings.Join(placeholders, ", "))

	args := make([]interface{}, len(ids)+1)
	args[0] = time.Now()
	for i := range ids {
		args[i+1] = ids[i]
	}
	_, err := repo.db.ExecContext(ctx, query, args...)

	if err != nil {
		return njnerror.Wrapf("transactionaljobrepo.ClaimTransactionalJob: failed to claim transactional job: %w", err)
	}
	return nil
}

// ReleaseTransactionalJob releases the given transactional job. Returns the first error encountered.
func (repo *transactionalJobRepo) ReleaseTransactionalJob(ctx context.Context, id string, lastError string) error {
	_, err := repo.db.ExecContext(ctx,
		`UPDATE transactional_jobs 
		SET available_at = DATE_ADD(NOW(), INTERVAL (retry_seconds * POWER(retry_backoff, attempts - 1)) SECOND), 
		processing_at = NULL, 
		last_error = ?
		WHERE id = ?`, lastError, id)

	if err != nil {
		return njnerror.Wrapf("transactionaljobrepo.ReleaseTransactionalJob: failed to release transactional job: %w", err)
	}
	return nil
}

// DeleteTransactionalJob deletes the given transactional job id. Returns the first error encountered.
func (repo *transactionalJobRepo) DeleteTransactionalJob(ctx context.Context, id string) error {
	query := `DELETE FROM transactional_jobs 
			  WHERE id = ?`

	_, err := repo.db.ExecContext(ctx, query, id)
	if err != nil {
		return njnerror.Wrapf("transactionaljobrepo.DeleteTransactionalJob: failed to delete transactional job: %w", err)
	}

	return nil
}
