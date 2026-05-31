package client

import (
	"context"
	"fmt"

	"github.com/jtenhave/not-just-noise/contracts/dispatch"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type db interface {
	ExecContext(ctx context.Context, query string, args ...any) (int64, error)
	IsTx(ctx context.Context) bool
}

type dispatchClient struct {
	db db
}

// NewDispatchClient creates a new dispatch client using the given database.
func NewDispatchClient(db db) *dispatchClient {
	return &dispatchClient{
		db: db,
	}
}

// Dispatch dispatches a new dispatch. Returns the first error encountered.
func (client *dispatchClient) Dispatch(ctx context.Context, dispatch dispatch.Dispatch) error {
	if !client.db.IsTx(ctx) {
		return fmt.Errorf("dispatchClient.Dispatch: not in a transaction")
	}

	_, err := client.db.ExecContext(ctx,
		`INSERT INTO transactional_job_service.transactional_jobs (id, callback_type, callback_resource, payload, claim_timeout, retry_seconds, retry_backoff, max_attempts) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		dispatch.ID, dispatch.CallbackType, dispatch.CallbackResource, dispatch.Payload, dispatch.ClaimTimeout, dispatch.RetrySeconds, dispatch.RetryBackoff, dispatch.MaxAttempts)

	if err != nil {
		return njnerror.Wrapf("dispatchClient.Dispatch: failed to dispatch: %w", err)
	}

	return nil
}
