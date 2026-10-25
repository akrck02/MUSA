package customers

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Itros97/MUSA/storage"
)

func testStore(test *testing.T) *Store {
	test.Helper()
	database, err := storage.Open(storage.InMemory)
	if err != nil {
		test.Fatal(err)
	}
	test.Cleanup(func() { database.Close() })
	counter := 0
	return NewStore(database, func() time.Time { return time.UnixMilli(0) }, func() string { counter++; return fmt.Sprint("id", counter) })
}

func TestAProfileBecomesALinkWithoutTheAt(test *testing.T) {
	network := Network{LinkPattern: "https://example.net/" + UserMark}
	if link := network.Link(" @ana_draws "); link != "https://example.net/ana_draws" {
		test.Errorf("link %q", link)
	}
}

func TestANewDatabaseKnowsInstagramBlueskyAndX(test *testing.T) {
	networks, err := testStore(test).Networks()
	if err != nil || len(networks) != 3 || networks[0].Name != "Instagram" || networks[2].Name != "X" {
		test.Errorf("networks %+v, %v", networks, err)
	}
}

func TestACustomerKeepsTheirProfilesButNotEmptyOnes(test *testing.T) {
	store := testStore(test)
	saved, problem, err := store.Save(Customer{Name: " Ana ", Profiles: []Profile{{NetworkID: "instagram", Handle: "@ana"}, {NetworkID: "x", Handle: " "}}})
	if err != nil || problem != "" {
		test.Fatal(problem, err)
	}
	customer, exists, err := store.Find(saved.ID)
	if err != nil || !exists || customer.Name != "Ana" || len(customer.Profiles) != 1 {
		test.Fatalf("found %+v %v", customer, err)
	}
	if handle, has := customer.Handle("instagram"); !has || handle != "@ana" {
		test.Errorf("instagram handle %q", handle)
	}
	if _, problem, _ := store.Save(Customer{Name: "  "}); problem != CodeBlankName {
		test.Errorf("a nameless customer: %q", problem)
	}
}

func TestANetworkNeedsWhereTheHandleGoes(test *testing.T) {
	store := testStore(test)
	if problem, _ := store.SaveNetwork(Network{Name: "Mastodon", LinkPattern: "https://mastodon.social/"}); problem != CodeNetworkPattern {
		test.Errorf("a pattern without %s: %q", UserMark, problem)
	}
	if problem, err := store.SaveNetwork(Network{Name: "Mastodon", LinkPattern: "https://mastodon.social/@" + UserMark}); problem != "" || err != nil {
		test.Fatal(problem, err)
	}
	networks, _ := store.Networks()
	if len(networks) != 4 || networks[3].Icon != "at-sign" {
		test.Errorf("networks %+v", networks)
	}
	if err := store.DeleteNetwork("x"); err != nil {
		test.Fatal(err)
	}
	if networks, _ := store.Networks(); len(networks) != 3 {
		test.Errorf("%d networks after deleting one", len(networks))
	}
}

func TestEveryProblemHasWordsAndTheFailureItsCode(test *testing.T) {
	for _, code := range []Code{CodeBlankName, CodeNetworkPattern} {
		if ProblemWords(code) == string(code) {
			test.Errorf("%s has no words", code)
		}
	}
	var failure *Error
	if err := storageFailed("x", errors.New("cause")); !errors.As(err, &failure) || failure.Code != CodeStorageFailed {
		test.Errorf("%v does not carry its code", err)
	}
}
