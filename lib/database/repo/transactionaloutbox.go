package repo

import (
	"context"

	"github.com/jtenhave/not-just-noise/lib/database"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
	"github.com/jtenhave/not-just-noise/lib/notify"
)

type transactionalOutboxRepo struct {
	db database.QueryRunner
}

// NewAudioRepo creates a new audio repository using the given database.
func NewTransactionalOutboxRepo(db database.QueryRunner) *transactionalOutboxRepo {
	return &transactionalOutboxRepo{
		db: db,
	}
}

func (repo *transactionalOutboxRepo) CreateTransactionalOutboxRecord(ctx context.Context, transactionalOutboxRecrod notify.TransactionalOutboxRecord) error {
	_, err := repo.db.ExecContext(ctx,
		`INSERT INTO transactional_outbox (id, payload) 
		VALUES (?, ?)`, transactionalOutboxRecrod.ID, transactionalOutboxRecrod.Payload)

	if err != nil {
		return njnerror.Wrapf("transactionaloutboxrepo.CreateTransactionalOutboxRecord: failed to create transactional outbox record: %w", err)
	}

	return nil
}
