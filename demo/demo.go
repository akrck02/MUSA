// Package demo fills an empty database with a made-up studio, to try MUSA without typing a
// catalogue first: prints, stickers, merch and two commissions, a closed fair from last month, one
// happening today, one coming up and four orders. It goes through the same stores as the screens.
package demo

import (
	"fmt"
	"time"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/customers"
	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/money"
	"github.com/Itros97/MUSA/projects"
	"github.com/Itros97/MUSA/sales"
	"github.com/Itros97/MUSA/storage"
)

// Stores are the stores the demo writes through.
type Stores struct {
	Database  *storage.Database
	Products  *catalogue.Store
	Sales     *sales.Store
	Events    *events.Store
	Customers *customers.Store
	Projects  *projects.Store
	Now       func() time.Time
	NewID     func() string
}

// CodeNotEmpty is the code of filling a database that already has data: its SKUs would clash.
const CodeNotEmpty = "demo.not-empty"

// CodeRefused is the code of demo data a store turned down, which is a bug in the demo.
const CodeRefused = "demo.refused"

// Fill loads the demo data, but only into an empty database.
func Fill(stores Stores) error {
	empty, err := stores.Database.Empty()
	if err != nil {
		return err
	}
	if !empty {
		return fmt.Errorf("%s: the database already has products or events", CodeNotEmpty)
	}
	filler := &filler{stores: stores, variants: map[string]string{}}
	return filler.fill()
}

// filler remembers the variants it made by SKU (or "product · variant" for digital ones), to sell
// them later; the first error stops it.
type filler struct {
	stores   Stores
	variants map[string]string
	err      error
}

func (filler *filler) fill() error {
	filler.catalogue()
	today := calendar.Of(filler.stores.Now())

	autumn := filler.event("Mercadillo de otoño", "Plaza Mayor", today.AddDays(-30), today.AddDays(-29))
	filler.take(autumn, map[string]int{"LAM-GAT-A4": 6, "PEG-LUNA": 15, "CHA-RANA": 10})
	filler.sell(autumn, "LAM-GAT-A4", 4, 1200)
	filler.sell(autumn, "LAM-GAT-A4", 1, 1000)
	filler.sell(autumn, "PEG-LUNA", 9, 300)
	filler.sell(autumn, "CHA-RANA", 6, 250)
	filler.spend(autumn, "Stand", 4000)
	filler.spend(autumn, "Gasolina", 2200)
	filler.refused(filler.stores.Events.Close(autumn))

	comic := filler.event("Feria del cómic", "Pabellón 3", today, today.AddDays(1))
	filler.take(comic, map[string]int{"LAM-GAT-A4": 5, "LAM-BOS-A3": 2, "PEG-LUNA": 12, "LLA-DRAG": 6})
	filler.sell(comic, "LAM-GAT-A4", 2, 1200)
	filler.sell(comic, "PEG-LUNA", 3, 300)
	filler.sell(comic, "LLA-DRAG", 1, 800)
	filler.sell(comic, "LAM-BOS-A3", 1, 1800)
	filler.spend(comic, "Stand", 6000)
	filler.spend(comic, "Comida", 1450)

	filler.event("Mercado de Navidad", "Centro cultural", today.AddDays(50), today.AddDays(52))

	filler.orders(today)
	filler.projects(today)
	return filler.err
}

func (filler *filler) orders(today calendar.Date) {
	daysAgo := func(count int) time.Time { return sales.SoldAt(today.AddDays(-count), filler.stores.Now()) }
	filler.order(sales.Draft{Channel: sales.InPerson, Customer: "Vecina del taller", SoldAt: daysAgo(0),
		Lines: filler.lines("LLA-DRAG", 2, 800)})
	filler.order(sales.Draft{Channel: sales.OnlineShop, Customer: "@lectora_de_comics", Note: "Pedido web 0007", SoldAt: daysAgo(3),
		Lines: append(filler.lines("LAM-BOS-A4", 2, 1200), filler.lines("PEG-HONG", 1, 600)...), ShippingCharged: 450, ShippingCost: 380})
	filler.order(sales.Draft{Channel: sales.Social, Customer: "@ranita_ilustra", Note: "Encargo por DM", SoldAt: daysAgo(10),
		Lines: filler.lines("Retrato digital · Busto", 1, 3500)})
	filler.order(sales.Draft{Channel: sales.Marketplace, Note: "Etsy", SoldAt: daysAgo(40),
		Lines: filler.lines("LAM-GAT-A3", 1, 2000), ShippingCharged: 500, ShippingCost: 420})
}

func (filler *filler) event(name, location string, start, end calendar.Date) string {
	id := filler.stores.NewID()
	filler.refused(filler.stores.Events.Save(events.Draft{ID: id, Name: name, Location: location, StartDate: start, EndDate: end}))
	return id
}

func (filler *filler) take(eventID string, units map[string]int) {
	for sku, quantity := range units {
		filler.refused(filler.stores.Events.MoveStock(eventID, filler.variants[sku], quantity))
	}
}

func (filler *filler) sell(eventID, sku string, quantity int, cents money.Money) {
	filler.refused(filler.stores.Events.Sell(eventID, filler.variants[sku], quantity, cents))
}

func (filler *filler) spend(eventID, concept string, cents money.Money) {
	filler.refused(filler.stores.Events.AddExpense(eventID, concept, cents))
}

func (filler *filler) lines(key string, quantity int, cents money.Money) []sales.LineDraft {
	return []sales.LineDraft{{ID: filler.stores.NewID(), VariantID: filler.variants[key], Quantity: quantity, UnitPrice: cents}}
}

func (filler *filler) order(draft sales.Draft) {
	if filler.err != nil {
		return
	}
	_, problems, err := filler.stores.Sales.Record(draft)
	filler.stop(len(problems), problems, err)
}

func (filler *filler) refused(refusals []events.Refusal, err error) {
	filler.stop(len(refusals), refusals, err)
}

// stop keeps the first error, or the first refusal as one: the demo's data must never be refused.
func (filler *filler) stop(refusedCount int, refusals any, err error) {
	if filler.err != nil {
		return
	}
	if err != nil {
		filler.err = err
		return
	}
	if refusedCount > 0 {
		filler.err = fmt.Errorf("%s: %v", CodeRefused, refusals)
	}
}
