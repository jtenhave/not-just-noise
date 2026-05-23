package repo

import (
	"context"

	"github.com/jtenhave/not-just-noise/lib/database"
	"github.com/jtenhave/not-just-noise/lib/database/repo"
)

type ConnectionRepository interface {
	Repository
	BeginTx(ctx context.Context) (TransactionRepository, error)
}

type connectionrepo struct {
	queryRunner             database.ConnectionQueryRunner
	audioRepo               AudioRepo
	transactionalOutboxRepo TransactionalOutboxRepo
}

func NewConnectionRepository(queryRunner database.ConnectionQueryRunner) ConnectionRepository {
	return &connectionrepo{
		queryRunner:             queryRunner,
		audioRepo:               NewAudioRepo(queryRunner),
		transactionalOutboxRepo: repo.NewTransactionalOutboxRepo(queryRunner),
	}
}

func (r *connectionrepo) BeginTx(ctx context.Context) (TransactionRepository, error) {
	tx, err := r.queryRunner.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	return NewTransactionRepository(tx), nil
}

func (r *connectionrepo) AudioRepo() AudioRepo {
	return r.audioRepo
}

func (r *connectionrepo) TransactionalOutboxRepo() TransactionalOutboxRepo {
	return r.transactionalOutboxRepo
}
