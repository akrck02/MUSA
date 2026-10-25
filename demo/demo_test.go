package demo

import (
	"slices"
	"testing"
	"time"

	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/customers"
	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/projects"
	"github.com/Itros97/MUSA/sales"
	"github.com/Itros97/MUSA/storage"
)

func openStores(test *testing.T) Stores {
	test.Helper()
	database, err := storage.Open(storage.InMemory)
	if err != nil {
		test.Fatal(err)
	}
	test.Cleanup(func() { database.Close() })
	now := func() time.Time { return time.Date(2026, 4, 18, 11, 0, 0, 0, time.Local) }
	products := catalogue.NewStore(database, now, storage.NewID)
	sold := sales.NewStore(database, products, now, storage.NewID)
	people := customers.NewStore(database, now, storage.NewID)
	return Stores{Database: database, Products: products, Sales: sold,
		Events: events.NewStore(database, sold, now, storage.NewID), Customers: people,
		Projects: projects.NewStore(database, products, people, now, storage.NewID), Now: now, NewID: storage.NewID}
}

func TestDemoDataFillsEveryModule(test *testing.T) {
	stores := openStores(test)
	if err := Fill(stores); err != nil {
		test.Fatal(err)
	}
	products, _ := stores.Products.All()
	fairs, _ := stores.Events.All()
	orders, _ := stores.Sales.All()
	commissions, _ := stores.Projects.All()
	// Products: twenty-one physical and three digital. Fairs: one closed, one today, two to come.
	// Sales: three at last month's fair, two today, eight orders, and the delivered project's.
	if len(products) != 21+3 || len(fairs) != 4 || len(orders) != 3+2+8+1 || len(commissions) != 5 {
		test.Errorf("%d products, %d fairs, %d sales, %d projects", len(products), len(fairs), len(orders), len(commissions))
	}
	if !slices.ContainsFunc(fairs, func(fair events.Event) bool { return fair.Status == events.Closed }) {
		test.Error("last month's fair is not closed")
	}
	lowStock := catalogue.Filter(products, catalogue.Query{LowStockOnly: true})
	if len(lowStock) == 0 {
		test.Error("the demo has no low stock to show")
	}
}

func TestTheDemoNeverFillsADatabaseWithData(test *testing.T) {
	stores := openStores(test)
	if err := Fill(stores); err != nil {
		test.Fatal(err)
	}
	if err := Fill(stores); CodeOf(err) != CodeNotEmpty {
		test.Errorf("filling twice: %v", err)
	}
	if err := stores.Database.Wipe(); err != nil {
		test.Fatal(err)
	}
	if empty, _ := stores.Database.Empty(); !empty {
		test.Error("wiping leaves data")
	}
}
