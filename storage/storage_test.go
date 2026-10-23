package storage

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestARowSurvivesClosingAndReopeningTheDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "musa.db")
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.Rows().ExecContext(t.Context(),
		"INSERT INTO tag(id, name, normalized_name, color, created_at, updated_at) VALUES ('1', 'Cats', 'cats', 'ROSE', 0, 0)")
	if err != nil {
		t.Fatal(err)
	}
	database.Close()

	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	var name string
	if err := reopened.Rows().QueryRowContext(t.Context(), "SELECT name FROM tag WHERE id = '1'").Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "Cats" {
		t.Errorf("name = %q", name)
	}
}

func TestAFailedTransactionKeepsNothing(t *testing.T) {
	database, err := Open(InMemory)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	stop := errors.New("stop")
	err = database.InTransaction(func(rows Querier) error {
		if _, err := rows.ExecContext(t.Context(),
			"INSERT INTO tag(id, name, normalized_name, color, created_at, updated_at) VALUES ('1', 'Cats', 'cats', 'ROSE', 0, 0)"); err != nil {
			return err
		}
		return stop
	})
	if !errors.Is(err, stop) {
		t.Fatalf("err = %v", err)
	}
	var count int
	database.Rows().QueryRowContext(t.Context(), "SELECT count(*) FROM tag").Scan(&count)
	if count != 0 {
		t.Errorf("%d tags kept", count)
	}
}

func TestEachErrorCarriesItsOwnCode(t *testing.T) {
	cause := errors.New("cause")
	cases := map[Code]error{CodeOpenFailed: openFailed("x", cause), CodeTransactionFailed: transactionFailed(cause)}
	for code, err := range cases {
		var failure *Error
		if !errors.As(err, &failure) || failure.Code != code {
			t.Errorf("%v does not carry %s", err, code)
		}
	}
}
