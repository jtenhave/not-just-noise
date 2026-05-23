package mysql

import (
	"context"
	"database/sql"
	"fmt"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type queryRunner interface {
	QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func queryContext(ctx context.Context, runner queryRunner, query string, args ...any) ([]map[string]any, error) {
	rows, err := runner.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, njnerror.Wrapf("libmysql.QueryContext: failed to query: %w", err)
	}

	rawRows := make([]map[string]any, 0)
	for rows.Next() {
		row := make(map[string]any)
		err = rows.StructScan(&row)
		if err != nil {
			return nil, njnerror.Wrapf("libmysql.QueryContext: failed to scan row: %w", err)
		}

		rawRows = append(rawRows, row)
	}

	return rawRows, nil
}

func execContext(ctx context.Context, db queryRunner, query string, args ...any) (int64, error) {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		if mysqlError, ok := err.(*mysqldriver.MySQLError); ok {
			if mysqlError.Number == 1062 {
				return 0, njnerror.NewNJNError(njnerror.Conflict, "libmysql.WriteQuery: duplicate entry")
			}
		}

		return 0, fmt.Errorf("libmysql.namedExecContext: failed to execute query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("libmysql.namedExecContext: failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
