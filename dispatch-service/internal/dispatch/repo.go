package dispatch

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

type dispatchRepo struct {
	db db
}

// NewDispatchRepo creates a new dispatch repository using the given database.
func NewDispatchRepo(db db) *dispatchRepo {
	return &dispatchRepo{
		db: db,
	}
}

// GetAvailableDispatches gets the next available dispatches, up to the given limit. Returns the first available dispatches and the first error encountered.
func (repo *dispatchRepo) GetAvailableDispatches(ctx context.Context, limit int) ([]Dispatch, error) {
	if !repo.db.IsTx(ctx) {
		return nil, fmt.Errorf("dispatchrepo.GetAvailableDispatches: not in a transaction")
	}

	dbRows, err := repo.db.QueryContext(ctx,
		`SELECT id, callback_type, callback_resource, payload
		 FROM dispatches
		 WHERE (processing_at IS NULL OR TIMESTAMPDIFF(SECOND, processing_at, NOW()) > claim_timeout) AND attempts < max_attempts AND available_at <= NOW()
		 ORDER BY created_at ASC
		 LIMIT ?
		 FOR UPDATE SKIP LOCKED`, limit)

	if err != nil {
		return nil, njnerror.Wrapf("dispatchrepo.GetAvailableDispatches: failed to get available dispatches: %w", err)
	}

	if len(dbRows) == 0 {
		return nil, nil
	}

	dispatches := make([]Dispatch, len(dbRows))
	for i, row := range dbRows {
		dispatch, err := toDispatch(row)
		if err != nil {
			return nil, njnerror.Wrapf("dispatchrepo.GetAvailableDispatches: failed to convert row to dispatch: %w", err)
		}
		dispatches[i] = dispatch
	}
	return dispatches, nil
}

// ToDispatch converts a row a to an Dispatch.
func toDispatch(row map[string]any) (Dispatch, error) {
	id, ok := row["id"].(string)
	if !ok {
		return Dispatch{}, fmt.Errorf("dispatchrepo.toDispatch: id is not a string")
	}
	callbackType, ok := row["callback_type"].(string)
	if !ok {
		return Dispatch{}, fmt.Errorf("dispatchrepo.toDispatch: callbackType is not a string")
	}
	callbackResource, ok := row["callback_resource"].(string)
	if !ok {
		return Dispatch{}, fmt.Errorf("dispatchrepo.toDispatch: callbackResource is not a string")
	}
	if !ok {
		return Dispatch{}, fmt.Errorf("dispatchrepo.toDispatch: callbackURL is not a string")
	}
	payload, ok := row["payload"].(string)
	if !ok {
		return Dispatch{}, fmt.Errorf("dispatchrepo.toDispatch: payload is not a string")
	}

	return Dispatch{
		ID:               id,
		CallbackType:     callbackType,
		CallbackResource: callbackResource,
		Payload:          payload,
	}, nil
}

// ClaimDispatches claims the given dispatches. Returns the first error encountered.
func (repo *dispatchRepo) ClaimDispatches(ctx context.Context, ids []string) error {
	if !repo.db.IsTx(ctx) {
		return fmt.Errorf("dispatchrepo.ClaimDispatches: not in a transaction")
	}

	placeholders := make([]string, len(ids))
	for i := range ids {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(`UPDATE dispatches 
		SET processing_at = ?, attempts = attempts + 1
		WHERE id in (%s)`, strings.Join(placeholders, ", "))

	args := make([]interface{}, len(ids)+1)
	args[0] = time.Now()
	for i := range ids {
		args[i+1] = ids[i]
	}
	_, err := repo.db.ExecContext(ctx, query, args...)

	if err != nil {
		return njnerror.Wrapf("dispatchrepo.ClaimDispatches: failed to claim dispatches: %w", err)
	}
	return nil
}

// ReleaseDispatches releases the given dispatch. Returns the first error encountered.
func (repo *dispatchRepo) ReleaseDispatch(ctx context.Context, id string, lastError string) error {
	_, err := repo.db.ExecContext(ctx,
		`UPDATE dispatches 
		SET available_at = DATE_ADD(NOW(), INTERVAL (retry_seconds * POWER(retry_backoff, attempts - 1)) SECOND), 
		processing_at = NULL, 
		last_error = ?
		WHERE id = ?`, lastError, id)

	if err != nil {
		return njnerror.Wrapf("dispatchrepo.ReleaseDispatches: failed to release dispatches: %w", err)
	}
	return nil
}

// DeleteDispatch deletes the given dispatch id. Returns the first error encountered.
func (repo *dispatchRepo) DeleteDispatch(ctx context.Context, id string) error {
	query := `DELETE FROM dispatches 
			  WHERE id = ?`

	_, err := repo.db.ExecContext(ctx, query, id)
	if err != nil {
		return njnerror.Wrapf("dispatchrepo.DeleteDispatch: failed to delete dispatch: %w", err)
	}

	return nil
}
