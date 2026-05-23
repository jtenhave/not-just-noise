package database

import "context"

type QueryRunner interface {
	ExecContext(ctx context.Context, query string, args ...any) (int64, error)
	QueryContext(ctx context.Context, query string, args ...any) ([]map[string]any, error)
}

type TransactionQueryRunner interface {
	QueryRunner
	Commit() error
	Rollback() error
}

type ConnectionQueryRunner interface {
	QueryRunner
	BeginTx(ctx context.Context) (TransactionQueryRunner, error)
}