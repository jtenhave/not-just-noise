package client

import (
	"context"
	"fmt"
	"strings"

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

	insertColumns := []string{"id", "callback_type", "callback_resource", "payload"}
	insertPlaceholders := []string{"?", "?", "?", "?"}
	insertValues := []any{dispatch.ID, dispatch.CallbackType, dispatch.CallbackResource, dispatch.Payload}

	if dispatch.ClaimTimeout != nil {
		insertColumns = append(insertColumns, "claim_timeout")
		insertPlaceholders = append(insertPlaceholders, "?")
		insertValues = append(insertValues, *dispatch.ClaimTimeout)
	}

	if dispatch.RetrySeconds != nil {
		insertColumns = append(insertColumns, "retry_seconds")
		insertPlaceholders = append(insertPlaceholders, "?")
		insertValues = append(insertValues, *dispatch.RetrySeconds)
	}

	if dispatch.RetryBackoff != nil {
		insertColumns = append(insertColumns, "retry_backoff")
		insertPlaceholders = append(insertPlaceholders, "?")
		insertValues = append(insertValues, *dispatch.RetryBackoff)
	}

	if dispatch.MaxAttempts != nil {
		insertColumns = append(insertColumns, "max_attempts")
		insertPlaceholders = append(insertPlaceholders, "?")
		insertValues = append(insertValues, *dispatch.MaxAttempts)
	}

	query := fmt.Sprintf(`INSERT INTO dispatch_service.dispatches (%s) VALUES (%s)`, strings.Join(insertColumns, ", "), strings.Join(insertPlaceholders, ", "))

	_, err := client.db.ExecContext(ctx, query, insertValues...)
	if err != nil {
		return njnerror.Wrapf("dispatchClient.Dispatch: failed to dispatch: %w", err)
	}

	return nil
}
