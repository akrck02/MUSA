package storage

import (
	"database/sql"
	"errors"
)

// Setting reads one of the artist's settings by name; found is false when it was never set.
func (database *Database) Setting(name string) (value string, found bool, err error) {
	err = database.connection.QueryRow("SELECT value FROM setting WHERE name = ?", name).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, transactionFailed(err)
	}
	return value, true, nil
}

// SetSetting keeps one of the artist's settings, replacing what it held. Settings are not data:
// Wipe leaves them alone.
func (database *Database) SetSetting(name, value string) error {
	_, err := database.connection.Exec("INSERT INTO setting(name, value) VALUES (?, ?) ON CONFLICT(name) DO UPDATE SET value = excluded.value", name, value)
	if err != nil {
		return transactionFailed(err)
	}
	return nil
}
