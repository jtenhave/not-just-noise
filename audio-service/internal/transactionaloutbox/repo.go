package transactionaloutbox

import (
	"context"

	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type db interface {
	ExecContext(ctx context.Context, query string, args ...any) (int64, error)
	QueryContext(ctx context.Context, query string, args ...any) ([]map[string]any, error)
}

type transactionalOutboxRepo struct {
	db db
}

// NewAudioRepo creates a new audio repository using the given database.
func NewTransactionalOutboxRepo(db db) *transactionalOutboxRepo {
	return &transactionalOutboxRepo{
		db: db,
	}
}

func (repo *transactionalOutboxRepo) CreateTransactionalOutboxRecord(ctx context.Context, transactionalOutboxRecrod TransactionalOutboxRecord) error {
	_, err := repo.db.ExecContext(ctx,
		`INSERT INTO transactional_outbox (id, payload) 
		VALUES (?, ?)`, transactionalOutboxRecrod.ID, transactionalOutboxRecrod.Payload)

	if err != nil {
		return njnerror.Wrapf("transactionaloutboxrepo.CreateTransactionalOutboxRecord: failed to create transactional outbox record: %w", err)
	}

	return nil
}
