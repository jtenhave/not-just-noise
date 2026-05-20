package database

import (
	"fmt"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/jtenhave/not-just-noise/lib/errorcode"
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
		return nil, fmt.Errorf("lib.mysql: failed to open database: %w", err)
	}

	return &mysql{
		db: db,
	}, nil
}

// Select runs the given query using args and stores the result in the dest variable. Returns the first error encountered.
func (db *mysql) Select(dest interface{}, query string, args ...interface{}) error {
	return db.db.Select(dest, query, args...)
}

// NamedExec executes the given query using the given source. Returns the first error encountered.
func (db *mysql) NamedExec(source interface{}, query string) error {
	_, err := db.db.NamedExec(query, source)
	if err != nil {
		if mysqlError, ok := err.(*mysqldriver.MySQLError); ok {
			if mysqlError.Number == 1062 {
				return errorcode.NewErrorCode(errorcode.Conflict, "duplicate entry")
			}
		}

		return fmt.Errorf("lib.mysql: failed to execute named query: %w", err)
	}

	return nil
}
