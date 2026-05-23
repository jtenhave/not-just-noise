package mysql

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/jtenhave/not-just-noise/lib/database"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type mysqlconnection struct {
	db *sqlx.DB
}

type MySQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
}

func NewConnectionQueryRunner(config MySQLConfig) (database.ConnectionQueryRunner, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", config.User, config.Password, config.Host, config.Port, config.DBName)
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("libmysql.NewMySQL: failed to open database: %w", err)
	}

	return &mysqlconnection{
		db: db,
	}, nil
}

func (db *mysqlconnection) BeginTx(ctx context.Context) (database.TransactionQueryRunner, error) {
	tx, err := db.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, njnerror.Wrapf("libmysql.BeginTx: failed to begin transaction: %w", err)
	}

	return NewTransactionQueryRunner(tx), nil
}

func (db *mysqlconnection) QueryContext(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	return queryContext(ctx, db.db, query, args...)
}

func (db *mysqlconnection) ExecContext(ctx context.Context, query string, args ...any) (int64, error) {
	return execContext(ctx, db.db, query, args...)
}
