package storage

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func TestARowSurvivesClosingAndReopeningTheDatabase(test *testing.T) {
	path := filepath.Join(test.TempDir(), "musa.db")
	database, err := Open(path)
	if err != nil {
		test.Fatal(err)
	}
	_, err = database.Rows().ExecContext(test.Context(),
		"INSERT INTO tag(id, name, normalized_name, color, created_at, updated_at) VALUES ('1', 'Cats', 'cats', 'ROSE', 0, 0)")
	if err != nil {
		test.Fatal(err)
	}
	database.Close()

	reopened, err := Open(path)
	if err != nil {
		test.Fatal(err)
	}
	defer reopened.Close()
	var name string
	if err := reopened.Rows().QueryRowContext(test.Context(), "SELECT name FROM tag WHERE id = '1'").Scan(&name); err != nil {
		test.Fatal(err)
	}
	if name != "Cats" {
		test.Errorf("name = %q", name)
	}
}

func TestAFailedTransactionKeepsNothing(test *testing.T) {
	database, err := Open(InMemory)
	if err != nil {
		test.Fatal(err)
	}
	defer database.Close()
	stop := errors.New("stop")
	err = database.InTransaction(func(rows Querier) error {
		if _, err := rows.ExecContext(test.Context(),
			"INSERT INTO tag(id, name, normalized_name, color, created_at, updated_at) VALUES ('1', 'Cats', 'cats', 'ROSE', 0, 0)"); err != nil {
			return err
		}
		return stop
	})
	if !errors.Is(err, stop) {
		test.Fatalf("err = %v", err)
	}
	var count int
	database.Rows().QueryRowContext(test.Context(), "SELECT count(*) FROM tag").Scan(&count)
	if count != 0 {
		test.Errorf("%d tags kept", count)
	}
}

func TestEachErrorCarriesItsOwnCode(test *testing.T) {
	cause := errors.New("cause")
	cases := map[Code]error{CodeOpenFailed: openFailed("x", cause), CodeTransactionFailed: transactionFailed(cause)}
	for code, err := range cases {
		var failure *Error
		if !errors.As(err, &failure) || failure.Code != code {
			test.Errorf("%v does not carry %s", err, code)
		}
	}
}

func TestAnOlderDatabaseGainsTheOrderStatusAndKeepsItsSales(test *testing.T) {
	path := filepath.Join(test.TempDir(), "musa.db")
	older, err := sql.Open("sqlite", path)
	if err != nil {
		test.Fatal(err)
	}
	if _, err := older.Exec(schema); err != nil {
		test.Fatal(err)
	}
	_, err = older.Exec(`INSERT INTO sale(id, channel, customer, note, shipping_charged_cents, shipping_cost_cents,
		sold_at, updated_at) VALUES ('s1', 'ONLINE_SHOP', '', '', 0, 0, 0, 0)`)
	if err != nil {
		test.Fatal(err)
	}
	older.Close()

	database, err := Open(path)
	if err != nil {
		test.Fatal(err)
	}
	defer database.Close()
	var status string
	if err := database.Rows().QueryRowContext(test.Context(), "SELECT status FROM sale WHERE id = 's1'").Scan(&status); err != nil {
		test.Fatal(err)
	}
	if status != "DELIVERED" {
		test.Errorf("an old sale is %q, not delivered", status)
	}
}

func TestOpeningTwiceTakesEachMigrationOnce(test *testing.T) {
	path := filepath.Join(test.TempDir(), "musa.db")
	for range 2 {
		database, err := Open(path)
		if err != nil {
			test.Fatal(err)
		}
		database.Close()
	}
}

func TestASettingIsKeptAcrossOpeningsAndSurvivesWiping(test *testing.T) {
	path := filepath.Join(test.TempDir(), "musa.db")
	database, err := Open(path)
	if err != nil {
		test.Fatal(err)
	}
	if _, found, err := database.Setting("theme"); found || err != nil {
		test.Fatalf("a new database already has a theme (found %v, err %v)", found, err)
	}
	if err := database.SetSetting("theme", "light"); err != nil {
		test.Fatal(err)
	}
	if err := database.SetSetting("theme", "dark"); err != nil {
		test.Fatal(err)
	}
	if err := database.Wipe(); err != nil {
		test.Fatal(err)
	}
	database.Close()
	reopened, err := Open(path)
	if err != nil {
		test.Fatal(err)
	}
	defer reopened.Close()
	if value, found, err := reopened.Setting("theme"); value != "dark" || !found || err != nil {
		test.Errorf("theme is %q (found %v, err %v), not the last one set", value, found, err)
	}
}
