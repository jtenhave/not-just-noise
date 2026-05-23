package repo

import (
	"github.com/jtenhave/not-just-noise/lib/database"
	"github.com/jtenhave/not-just-noise/lib/database/repo"
)

type TransactionRepository interface {
	Repository
	Commit() error
	Rollback() error
}

type transactionrepo struct {
	queryRunner             database.TransactionQueryRunner
	audioRepo               AudioRepo
	transactionalOutboxRepo TransactionalOutboxRepo
}

func NewTransactionRepository(queryRunner database.TransactionQueryRunner) TransactionRepository {
	return &transactionrepo{
		queryRunner:             queryRunner,
		audioRepo:               NewAudioRepo(queryRunner),
		transactionalOutboxRepo: repo.NewTransactionalOutboxRepo(queryRunner),
	}
}

func (r *transactionrepo) Commit() error {
	return r.queryRunner.Commit()
}

func (r *transactionrepo) Rollback() error {
	return r.queryRunner.Rollback()
}

func (r *transactionrepo) AudioRepo() AudioRepo {
	return r.audioRepo
}

func (r *transactionrepo) TransactionalOutboxRepo() TransactionalOutboxRepo {
	return r.transactionalOutboxRepo
}
