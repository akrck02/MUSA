package customers

import (
	"context"
	"strings"
	"time"

	"github.com/Itros97/MUSA/storage"
)

// Store keeps customers and networks in the database.
type Store struct {
	database *storage.Database
	now      func() time.Time
	newID    func() string
}

// NewStore keeps customers in database; tests pass a fixed clock and fixed ids.
func NewStore(database *storage.Database, now func() time.Time, newID func() string) *Store {
	return &Store{database: database, now: now, newID: newID}
}

// Networks reads the networks, in the artist's order.
func (store *Store) Networks() ([]Network, error) {
	result, err := store.database.Rows().QueryContext(context.Background(),
		"SELECT id, name, link_pattern, icon FROM social_network ORDER BY position, name")
	if err != nil {
		return nil, storageFailed("read networks", err)
	}
	defer result.Close()
	var networks []Network
	for result.Next() {
		var network Network
		if err := result.Scan(&network.ID, &network.Name, &network.LinkPattern, &network.Icon); err != nil {
			return nil, storageFailed("read networks", err)
		}
		networks = append(networks, network)
	}
	return networks, result.Err()
}

// SaveNetwork adds a network or changes one; it gives back the problem that stops it, if any.
func (store *Store) SaveNetwork(network Network) (Code, error) {
	if strings.TrimSpace(network.Name) == "" {
		return CodeBlankName, nil
	}
	if !strings.Contains(network.LinkPattern, UserMark) {
		return CodeNetworkPattern, nil
	}
	if network.ID == "" {
		network.ID = store.newID()
	}
	if network.Icon == "" {
		network.Icon = "at-sign"
	}
	_, err := store.database.Rows().ExecContext(context.Background(), `INSERT INTO social_network(id, name, link_pattern, icon, position)
		VALUES (?, ?, ?, ?, (SELECT count(*) FROM social_network))
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, link_pattern = excluded.link_pattern, icon = excluded.icon`,
		network.ID, strings.TrimSpace(network.Name), strings.TrimSpace(network.LinkPattern), network.Icon)
	if err != nil {
		return "", storageFailed("save network", err)
	}
	return "", nil
}

// DeleteNetwork removes a network and every handle on it.
func (store *Store) DeleteNetwork(id string) error {
	return store.database.InTransaction(func(rows storage.Querier) error {
		for _, statement := range []string{"DELETE FROM customer_profile WHERE network_id = ?", "DELETE FROM social_network WHERE id = ?"} {
			if _, err := rows.ExecContext(context.Background(), statement, id); err != nil {
				return storageFailed("delete network", err)
			}
		}
		return nil
	})
}
