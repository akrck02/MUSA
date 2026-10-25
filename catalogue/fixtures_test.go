package catalogue

import (
	"fmt"
	"testing"
	"time"

	"github.com/Itros97/MUSA/storage"
)

// epoch is the fixed moment the tests' clock starts at.
var epoch = time.UnixMilli(1_700_000_000_000)

func tag(name string) Tag { return Tag{ID: "tag-" + name, Name: name, Color: Teal} }

func physicalVariant(id string) Variant { return Variant{ID: id, Price: 1000, Cost: 300, Stock: 5} }

func physicalProduct(id, name string) Product {
	return Product{ID: id, Kind: Physical, Name: name, Status: Active, Material: "Papel", Variants: []Variant{physicalVariant(id + "-v1")}}
}

func digitalProduct(id, name string) Product {
	return Product{ID: id, Kind: Digital, Name: name, Status: Active, DeliveryFormat: "PNG", TurnaroundDays: 14,
		Variants: []Variant{{ID: id + "-d1", Price: 4000}}}
}

// testStore opens a store on a database in memory, with a clock that moves one minute per reading
// and ids counted from one.
func testStore(test *testing.T) *Store {
	test.Helper()
	database, err := storage.Open(storage.InMemory)
	if err != nil {
		test.Fatal(err)
	}
	test.Cleanup(func() { database.Close() })
	moment, counter := epoch, 0
	clock := func() time.Time { moment = moment.Add(time.Minute); return moment }
	ids := func() string { counter++; return fmt.Sprintf("id-%d", counter) }
	return NewStore(database, clock, ids)
}

func mustSave(test *testing.T, store *Store, draft ProductDraft) Product {
	test.Helper()
	product, problems, err := store.Save(draft)
	if err != nil || len(problems) > 0 {
		test.Fatalf("save: %v %v", problems, err)
	}
	return product
}

func mustFind(test *testing.T, store *Store, id string) Product {
	test.Helper()
	product, found, err := store.Find(id)
	if err != nil || !found {
		test.Fatalf("find %s: %v %v", id, found, err)
	}
	return product
}
