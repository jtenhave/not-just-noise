package transactionaljob

import (
	"context"
	"fmt"

	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type db interface {
	ExecContext(ctx context.Context, query string, args ...any) (int64, error)
	IsTx(ctx context.Context) bool
}

type transactionalJobClient struct {
	db db
}

// NewTransactionalJobClient creates a new transactional job client using the given database.
func NewTransactionalJobClient(db db) *transactionalJobClient {
	return &transactionalJobClient{
		db: db,
	}
}

// CreateTransactionalJob creates a new transactional job. Returns the first error encountered.
func (client *transactionalJobClient) CreateTransactionalJob(ctx context.Context, transactionalJob TransactionalJob) error {
	if !client.db.IsTx(ctx) {
		return fmt.Errorf("transactionaljobrepo.CreateTransactionalJob: not in a transaction")
	}

	_, err := client.db.ExecContext(ctx,
		`INSERT INTO transactional_job_service.transactional_jobs (id, callback_url, payload, claim_timeout, retry_seconds, retry_backoff, max_attempts) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		transactionalJob.ID, transactionalJob.CallbackURL, transactionalJob.Payload, transactionalJob.ClaimTimeout, transactionalJob.RetrySeconds, transactionalJob.RetryBackoff, transactionalJob.MaxAttempts)

	if err != nil {
		return njnerror.Wrapf("transactionaljobrepo.CreateTransactionalJob: failed to create transactional job: %w", err)
	}

	return nil
}
