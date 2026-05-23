package database

import (
	"context"
	"database/sql"
	"fmt"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

const txID = "txID"

type mysqlconnection struct {
	db           *sqlx.DB
	transactions map[string]*sqlx.Tx
}

type MySQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
}

type queryRunner interface {
	QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func NewMySQLConnection(config MySQLConfig) (*mysqlconnection, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", config.User, config.Password, config.Host, config.Port, config.DBName)
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("libmysql.NewMySQL: failed to open database: %w", err)
	}

	return &mysqlconnection{
		db: db,
	}, nil
}

func (db *mysqlconnection) WithinTx(ctx context.Context, transaction func(context.Context) error) error {
	tx, err := db.db.BeginTxx(ctx, nil)
	if err != nil {
		return njnerror.Wrapf("libmysql.BeginTx: failed to begin transaction: %w", err)
	}

	defer tx.Rollback()

	ctx = context.WithValue(ctx, txID, tx)

	err = transaction(ctx)
	if err != nil {
		return njnerror.Wrapf("libmysql.BeginTx: failed to execute transaction: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return njnerror.Wrapf("libmysql.BeginTx: failed to commit transaction: %w", err)
	}

	return nil
}

func (db *mysqlconnection) QueryContext(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	var runner queryRunner = db.db
	if tx, ok := ctx.Value(txID).(*sqlx.Tx); ok {
		runner = tx
	}

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

func (db *mysqlconnection) ExecContext(ctx context.Context, query string, args ...any) (int64, error) {
	var runner queryRunner = db.db
	if tx, ok := ctx.Value(txID).(*sqlx.Tx); ok {
		runner = tx
	}

	result, err := runner.ExecContext(ctx, query, args...)
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
