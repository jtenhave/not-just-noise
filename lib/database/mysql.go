package database

import (
	"fmt"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
)

type mysql struct {
	db *sqlx.DB
}

type MySQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
}

// NewMySQL creates a new MySQL client using the given config. Returns a pointer to the mysql client and the first error encountered.
func NewMySQL(config MySQLConfig) (*mysql, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", config.User, config.Password, config.Host, config.Port, config.DBName)
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("libmysql.NewMySQL: failed to open database: %w", err)
	}

	return &mysql{
		db: db,
	}, nil
}

// ReadQuery runs the given query using args and stores the result in the dest variable. Returns the first error encountered.
func (db *mysql) ReadQuery(query string, dest interface{}, args ...interface{}) error {
	return db.db.Select(dest, query, args...)
}

// WriteQuery executes the given query using the given source. Returns the first error encountered.
func (db *mysql) WriteQuery(query string, source interface{}) error {
	result, err := db.db.NamedExec(query, source)
	if err != nil {
		if mysqlError, ok := err.(*mysqldriver.MySQLError); ok {
			if mysqlError.Number == 1062 {
				return njnerror.NewNJNError(njnerror.Conflict, "libmysql.WriteQuery: duplicate entry")
			}
		}

		return fmt.Errorf("libmysql.WriteQuery: failed to execute query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("libmysql.WriteQuery: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return njnerror.NewNJNError(njnerror.NotFound, "libmysql.WriteQuery: not found")
	}

	return nil
}
