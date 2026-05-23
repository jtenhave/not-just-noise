package mysql

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/jtenhave/not-just-noise/lib/database"
)

type mysqlTx struct {
	tx *sqlx.Tx
}

func NewTransactionQueryRunner(tx *sqlx.Tx) database.TransactionQueryRunner {
	return &mysqlTx{
		tx: tx,
	}
}

func (tx *mysqlTx) Commit() error {
	return tx.tx.Commit()
}

func (tx *mysqlTx) Rollback() error {
	return tx.tx.Rollback()
}

func (tx *mysqlTx) QueryContext(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	return queryContext(ctx, tx.tx, query, args...)
}

func (tx *mysqlTx) ExecContext(ctx context.Context, query string, args ...any) (int64, error) {
	return execContext(ctx, tx.tx, query, args...)
}
