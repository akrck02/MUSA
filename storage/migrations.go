package storage

import (
	"context"
	"database/sql"
	"fmt"
)

// migrations bring a database made by an older MUSA up to today's tables, in order. The schema
// file makes the tables of the first version; the database's user_version counts how many of
// these steps it has taken. A new step goes at the end, and is never edited once released.
var migrations = []string{
	// Orders have a status. Sales recorded before it are taken as delivered: they were done.
	"ALTER TABLE sale ADD COLUMN status TEXT NOT NULL DEFAULT 'DELIVERED'",
}

// migrate takes the steps a database has not taken yet, each in its own transaction.
func migrate(connection *sql.DB) error {
	var version int
	if err := connection.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	for step := version; step < len(migrations); step++ {
		if err := takeStep(connection, step); err != nil {
			return err
		}
	}
	return nil
}

func takeStep(connection *sql.DB, step int) error {
	transaction, err := connection.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	if _, err := transaction.Exec(migrations[step]); err != nil {
		transaction.Rollback() // the step's error is the one worth telling
		return err
	}
	// PRAGMA takes no parameters; step is a number this package counts.
	if _, err := transaction.Exec(fmt.Sprintf("PRAGMA user_version = %d", step+1)); err != nil {
		transaction.Rollback() // the pragma's error is the one worth telling
		return err
	}
	return transaction.Commit()
}
