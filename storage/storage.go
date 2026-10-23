// Package storage opens MUSA's database: one SQLite file on the device, with the tables of every
// feature. Features write their own queries; this package only gives them the connection, the
// schema and transactions.
package storage

import (
	"context"
	"database/sql"
	_ "embed"
	"time"

	_ "modernc.org/sqlite" // the SQLite driver, in pure Go so phones need no C library
)

//go:embed schema.sql
var schema string

// InMemory is the path that opens a database living only in memory, for tests and the demo.
const InMemory = ":memory:"

// Database is an open MUSA database.
type Database struct {
	connection *sql.DB
}

// Querier is what reads and writes rows: the database itself or a transaction in progress.
type Querier interface {
	ExecContext(context context.Context, query string, arguments ...any) (sql.Result, error)
	QueryContext(context context.Context, query string, arguments ...any) (*sql.Rows, error)
	QueryRowContext(context context.Context, query string, arguments ...any) *sql.Row
}

// Open opens the database file at path, making it and its tables when missing and bringing an
// older one up to date.
func Open(path string) (*Database, error) {
	connection, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, openFailed(path, err)
	}
	// One connection: SQLite writes one at a time, and an in-memory database lives in its connection.
	connection.SetMaxOpenConns(1)
	if _, err := connection.Exec(schema); err != nil {
		connection.Close() // the schema error is the one worth telling
		return nil, openFailed(path, err)
	}
	if err := migrate(connection); err != nil {
		connection.Close() // the migration error is the one worth telling
		return nil, openFailed(path, err)
	}
	return &Database{connection: connection}, nil
}

// Close closes the database.
func (database *Database) Close() error { return database.connection.Close() }

// Rows is the database to read and write outside a transaction.
func (database *Database) Rows() Querier { return database.connection }

// InTransaction runs write inside one transaction: everything it does is kept, or nothing is when
// it returns an error.
func (database *Database) InTransaction(write func(rows Querier) error) error {
	transaction, err := database.connection.Begin()
	if err != nil {
		return transactionFailed(err)
	}
	if err := write(transaction); err != nil {
		transaction.Rollback() // the write's error is the one worth telling
		return err
	}
	if err := transaction.Commit(); err != nil {
		return transactionFailed(err)
	}
	return nil
}

// Milliseconds is how a moment is stored: Unix milliseconds.
func Milliseconds(moment time.Time) int64 { return moment.UnixMilli() }

// Moment reads back a stored moment.
func Moment(milliseconds int64) time.Time { return time.UnixMilli(milliseconds) }

// tables lists every table, for wiping them all.
var tables = []string{"product", "digital_product", "physical_product", "product_variant", "physical_variant", "tag",
	"product_tag", "event", "event_stock", "event_expense", "sale", "sale_line"}

// Empty tells whether there are no products and no events, deleted ones included.
func (database *Database) Empty() (bool, error) {
	var count int
	err := database.connection.QueryRow("SELECT (SELECT count(*) FROM product) + (SELECT count(*) FROM event)").Scan(&count)
	if err != nil {
		return false, transactionFailed(err)
	}
	return count == 0, nil
}

// Wipe removes every row of every table, in one transaction. It cannot be undone: only the
// development tools offer it.
func (database *Database) Wipe() error {
	return database.InTransaction(func(rows Querier) error {
		for _, table := range tables {
			if _, err := rows.ExecContext(context.Background(), "DELETE FROM "+table); err != nil {
				return transactionFailed(err)
			}
		}
		return nil
	})
}
