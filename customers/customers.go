package customers

import (
	"context"
	"strings"

	"github.com/Itros97/MUSA/storage"
)

// Find reads one customer with their profiles; found is false when there is none with id.
func (store *Store) Find(id string) (customer Customer, found bool, err error) {
	all, err := store.read(id)
	if err != nil || len(all) == 0 {
		return Customer{}, false, err
	}
	return all[0], true, nil
}

// All reads every customer with their profiles, by name.
func (store *Store) All() ([]Customer, error) { return store.read("") }

// Save adds a customer or changes one, profiles included: a profile with an empty handle is
// dropped. It gives back the saved customer, or the problem that stops it.
func (store *Store) Save(customer Customer) (Customer, Code, error) {
	customer.Name = strings.TrimSpace(customer.Name)
	if customer.Name == "" {
		return Customer{}, CodeBlankName, nil
	}
	if customer.ID == "" {
		customer.ID = store.newID()
	}
	customer.Contact = strings.TrimSpace(customer.Contact)
	var kept []Profile
	for _, profile := range customer.Profiles {
		if handle := strings.TrimSpace(profile.Handle); handle != "" {
			kept = append(kept, Profile{NetworkID: profile.NetworkID, Handle: handle})
		}
	}
	customer.Profiles = kept
	err := store.database.InTransaction(func(rows storage.Querier) error { return store.write(rows, customer) })
	return customer, "", err
}

func (store *Store) write(rows storage.Querier, customer Customer) error {
	at := storage.Milliseconds(store.now())
	_, err := rows.ExecContext(context.Background(), `INSERT INTO customer(id, name, contact, created_at, updated_at) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, contact = excluded.contact, updated_at = excluded.updated_at`,
		customer.ID, customer.Name, customer.Contact, at, at)
	if err != nil {
		return storageFailed("save customer", err)
	}
	if _, err := rows.ExecContext(context.Background(), "DELETE FROM customer_profile WHERE customer_id = ?", customer.ID); err != nil {
		return storageFailed("save profiles", err)
	}
	for _, profile := range customer.Profiles {
		_, err := rows.ExecContext(context.Background(), "INSERT INTO customer_profile(customer_id, network_id, handle) VALUES (?, ?, ?)",
			customer.ID, profile.NetworkID, profile.Handle)
		if err != nil {
			return storageFailed("save profiles", err)
		}
	}
	return nil
}

func (store *Store) read(onlyID string) ([]Customer, error) {
	rows := store.database.Rows()
	result, err := rows.QueryContext(context.Background(),
		"SELECT id, name, contact FROM customer WHERE ?1 = '' OR id = ?1 ORDER BY name", onlyID)
	if err != nil {
		return nil, storageFailed("read customers", err)
	}
	var customers []Customer
	for result.Next() {
		var customer Customer
		if err := result.Scan(&customer.ID, &customer.Name, &customer.Contact); err != nil {
			result.Close() // the scan error is the one worth telling
			return nil, storageFailed("read customers", err)
		}
		customers = append(customers, customer)
	}
	result.Close() // all read: closing frees the one connection for the profiles
	if err := result.Err(); err != nil {
		return nil, storageFailed("read customers", err)
	}
	for index := range customers {
		profiles, err := store.profiles(customers[index].ID)
		if err != nil {
			return nil, err
		}
		customers[index].Profiles = profiles
	}
	return customers, nil
}

func (store *Store) profiles(customerID string) ([]Profile, error) {
	result, err := store.database.Rows().QueryContext(context.Background(), `SELECT customer_profile.network_id, handle
		FROM customer_profile JOIN social_network ON social_network.id = customer_profile.network_id
		WHERE customer_id = ? ORDER BY social_network.position`, customerID)
	if err != nil {
		return nil, storageFailed("read profiles", err)
	}
	defer result.Close()
	var profiles []Profile
	for result.Next() {
		var profile Profile
		if err := result.Scan(&profile.NetworkID, &profile.Handle); err != nil {
			return nil, storageFailed("read profiles", err)
		}
		profiles = append(profiles, profile)
	}
	return profiles, result.Err()
}
