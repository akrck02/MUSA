package customers

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Itros97/MUSA/storage"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	database, err := storage.Open(storage.InMemory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	counter := 0
	return NewStore(database, func() time.Time { return time.UnixMilli(0) }, func() string { counter++; return fmt.Sprint("id", counter) })
}

func TestAProfileBecomesALinkWithoutTheAt(t *testing.T) {
	network := Network{LinkPattern: "https://example.net/" + UserMark}
	if link := network.Link(" @ana_draws "); link != "https://example.net/ana_draws" {
		t.Errorf("link %q", link)
	}
}

func TestANewDatabaseKnowsInstagramBlueskyAndX(t *testing.T) {
	networks, err := testStore(t).Networks()
	if err != nil || len(networks) != 3 || networks[0].Name != "Instagram" || networks[2].Name != "X" {
		t.Errorf("networks %+v, %v", networks, err)
	}
}

func TestACustomerKeepsTheirProfilesButNotEmptyOnes(t *testing.T) {
	store := testStore(t)
	saved, problem, err := store.Save(Customer{Name: " Ana ", Profiles: []Profile{{NetworkID: "instagram", Handle: "@ana"}, {NetworkID: "x", Handle: " "}}})
	if err != nil || problem != "" {
		t.Fatal(problem, err)
	}
	found, ok, err := store.Find(saved.ID)
	if err != nil || !ok || found.Name != "Ana" || len(found.Profiles) != 1 {
		t.Fatalf("found %+v %v", found, err)
	}
	if handle, ok := found.Handle("instagram"); !ok || handle != "@ana" {
		t.Errorf("instagram handle %q", handle)
	}
	if _, problem, _ := store.Save(Customer{Name: "  "}); problem != CodeBlankName {
		t.Errorf("a nameless customer: %q", problem)
	}
}

func TestANetworkNeedsWhereTheHandleGoes(t *testing.T) {
	store := testStore(t)
	if problem, _ := store.SaveNetwork(Network{Name: "Mastodon", LinkPattern: "https://mastodon.social/"}); problem != CodeNetworkPattern {
		t.Errorf("a pattern without %s: %q", UserMark, problem)
	}
	if problem, err := store.SaveNetwork(Network{Name: "Mastodon", LinkPattern: "https://mastodon.social/@" + UserMark}); problem != "" || err != nil {
		t.Fatal(problem, err)
	}
	networks, _ := store.Networks()
	if len(networks) != 4 || networks[3].Icon != "at-sign" {
		t.Errorf("networks %+v", networks)
	}
	if err := store.DeleteNetwork("x"); err != nil {
		t.Fatal(err)
	}
	if networks, _ := store.Networks(); len(networks) != 3 {
		t.Errorf("%d networks after deleting one", len(networks))
	}
}

func TestEveryProblemHasWordsAndTheFailureItsCode(t *testing.T) {
	for _, code := range []Code{CodeBlankName, CodeNetworkPattern} {
		if ProblemWords(code) == string(code) {
			t.Errorf("%s has no words", code)
		}
	}
	var failure *Error
	if err := storageFailed("x", errors.New("cause")); !errors.As(err, &failure) || failure.Code != CodeStorageFailed {
		t.Errorf("%v does not carry its code", err)
	}
}
