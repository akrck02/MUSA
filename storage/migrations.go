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
	// Sales say how they were paid; older ones did not say.
	"ALTER TABLE sale ADD COLUMN paid_with TEXT NOT NULL DEFAULT ''",
	// Customers, the social networks the artist reaches them on, and commissioned projects.
	`CREATE TABLE social_network (
		id TEXT NOT NULL PRIMARY KEY,
		name TEXT NOT NULL,
		link_pattern TEXT NOT NULL,
		icon TEXT NOT NULL,
		position INTEGER NOT NULL
	);
	INSERT INTO social_network VALUES
		('instagram', 'Instagram', 'https://instagram.com/{user}', 'camera', 0),
		('bluesky', 'Bluesky', 'https://bsky.app/profile/{user}', 'cloud', 1),
		('x', 'X', 'https://x.com/{user}', 'x', 2);
	CREATE TABLE customer (
		id TEXT NOT NULL PRIMARY KEY,
		name TEXT NOT NULL,
		contact TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	CREATE TABLE customer_profile (
		customer_id TEXT NOT NULL,
		network_id TEXT NOT NULL,
		handle TEXT NOT NULL,
		PRIMARY KEY (customer_id, network_id)
	);
	CREATE TABLE project (
		id TEXT NOT NULL PRIMARY KEY,
		title TEXT NOT NULL,
		customer_id TEXT NOT NULL,
		stage TEXT NOT NULL,
		deadline TEXT NOT NULL,
		price_cents INTEGER NOT NULL,
		variant_id TEXT NOT NULL,
		sale_id TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		deleted_at INTEGER
	);`,
	// The artist's settings, by name: the theme they chose, for now.
	"CREATE TABLE setting (name TEXT PRIMARY KEY, value TEXT NOT NULL)",
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
