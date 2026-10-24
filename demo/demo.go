// Package demo fills an empty database with a made-up studio, to try MUSA without typing a
// catalogue first: the works, fairs and orders of the author's design, a closed fair from last
// month, one happening today and three to come. It goes through the same stores as the screens.
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
	filler := &filler{stores: stores, variants: map[string]string{}, finalStock: map[string]int{}}
	return filler.fill()
}

// filler remembers the variants it made by SKU (or "product · variant" for digital ones), to sell
// them later; the first error stops it.
type filler struct {
	stores   Stores
	variants map[string]string
	// finalStock is the stock each version, by its SKU, should end with.
	finalStock map[string]int
	err        error
}

func (filler *filler) fill() error {
	filler.catalogue()
	today := calendar.Of(filler.stores.Now())
	filler.fairs(today)
	filler.orders(today)
	filler.projects(today)
	filler.settleStock()
	return filler.err
}

// piece is some units of a version, by its SKU, in the order a stand shows them.
type piece struct {
	sku   string
	units int
}

// fairs are a closed one from last month, the one on today with its first sales, and three to come.
func (filler *filler) fairs(today calendar.Date) {
	summer := filler.event("Mercadillo de verano", "Plaza Mayor", today.AddDays(-30), today.AddDays(-29))
	filler.take(summer, []piece{{"LAM-GAT-A5", 6}, {"PEG-LUN-B", 15}, {"PEG-NUB", 10}})
	filler.sell(summer, "LAM-GAT-A5", 5, 800)
	filler.sell(summer, "PEG-LUN-B", 9, 300)
	filler.sell(summer, "PEG-NUB", 6, 300)
	filler.spend(summer, "Stand", 4000)
	filler.spend(summer, "Gasolina", 2200)
	filler.refused(filler.stores.Events.Close(summer))

	// Forty-two pieces, as the author's design counts them.
	autumn := filler.event("Mercado de Otoño", "Paseo del Río", today, today)
	filler.take(autumn, []piece{{"LAM-GAT-A5", 4}, {"PEG-LUN-M", 3}, {"LLA-RAN", 4}, {"PEG-OLA", 6},
		{"PEG-CAF", 2}, {"PEG-PLA", 5}, {"LAM-ZOR-A5", 3}, {"PEG-NUB", 8}, {"PEG-CER", 4}, {"PEG-CAC", 3}})
	filler.sellLines(autumn, []piece{{"LAM-GAT-A5", 2}, {"LLA-RAN", 1}})
	filler.sellLines(autumn, []piece{{"PEG-LUN-M", 1}})
	filler.spend(autumn, "Stand", 3500)

	harbour := filler.event("Feria del Puerto", "Muelle Norte", today.AddDays(3), today.AddDays(3))
	filler.take(harbour, []piece{{"LAM-BOS-A4", 6}, {"PEG-NUB", 10}, {"LLA-TAC", 8}, {"LAM-KOI-A4", 8}, {"LAM-DON-A4", 10}})
	filler.event("Feria del Libro", "Recinto Ferial", today.AddDays(17), today.AddDays(18))
	filler.event("Mercadillo de Invierno", "Centro cultural", today.AddDays(38), today.AddDays(39))
}

// orders are two to prepare, two on their way and the month's history: a shop's order already
// delivered and two sales in hand.
func (filler *filler) orders(today calendar.Date) {
	daysAgo := func(count int) time.Time { return sales.SoldAt(today.AddDays(-count), filler.stores.Now()) }
	etsy := func(customer string, days int, lines []sales.LineDraft) sales.Draft {
		return sales.Draft{Channel: sales.Marketplace, Note: "Etsy", Customer: customer, SoldAt: daysAgo(days), Lines: lines}
	}
	filler.order(etsy("Lucía", 1, append(filler.lines("LAM-ZOR-A4", 1, 1500), filler.lines("LAM-GAT-A5", 1, 800)...)))
	filler.order(etsy("Marcos", 0, filler.lines("LAM-GAT-A4", 1, 1200)))
	filler.sent(etsy("Ana", 3, filler.lines("LAM-BOS-A4", 1, 1000)))
	filler.sent(etsy("Pablo", 4, filler.lines("LAM-ZOR-A5", 1, 800)))
	filler.delivered(etsy("Sara", 1, filler.lines("LAM-BOS-A4", 1, 1000)))
	filler.order(sales.Draft{Channel: sales.InPerson, Customer: "Hugo", SoldAt: daysAgo(2), Lines: filler.lines("PEG-OLA", 2, 300)})
	filler.order(sales.Draft{Channel: sales.InPerson, Customer: "Taller Norte", SoldAt: daysAgo(3),
		Lines: append(append(filler.lines("LAM-GAT-A3", 1, 1500), filler.lines("LAM-BOS-A4", 1, 1000)...), filler.lines("LLA-RAN", 1, 600)...)})
	filler.delivered(etsy("Irene", 40, filler.lines("LAM-GAT-A4", 1, 1200)))
}

// sent records an order and marks it prepared, so it is on its way.
func (filler *filler) sent(draft sales.Draft) {
	filler.advanced(draft, 1)
}

// delivered records an order and moves it along until it is delivered.
func (filler *filler) delivered(draft sales.Draft) {
	filler.advanced(draft, 2)
}

func (filler *filler) advanced(draft sales.Draft, steps int) {
	if filler.err != nil {
		return
	}
	sale, problems, err := filler.stores.Sales.Record(draft)
	filler.stop(len(problems), problems, err)
	for step := 0; step < steps && filler.err == nil; step++ {
		problems, err := filler.stores.Sales.Advance(sale.ID)
		filler.stop(len(problems), problems, err)
	}
}

func (filler *filler) event(name, location string, start, end calendar.Date) string {
	id := filler.stores.NewID()
	filler.refused(filler.stores.Events.Save(events.Draft{ID: id, Name: name, Location: location, StartDate: start, EndDate: end}))
	return id
}

func (filler *filler) take(eventID string, pieces []piece) {
	for _, each := range pieces {
		filler.refused(filler.stores.Events.MoveStock(eventID, filler.variants[each.sku], each.units))
	}
}

// sellLines sells some pieces at a fair to one customer, at the stand's prices.
func (filler *filler) sellLines(eventID string, pieces []piece) {
	if filler.err != nil {
		return
	}
	detail, err := filler.stores.Events.Detail(eventID)
	filler.stop(0, nil, err)
	var lines []sales.LineDraft
	for _, each := range pieces {
		line, _ := detail.LineOf(filler.variants[each.sku])
		lines = append(lines, sales.LineDraft{ID: filler.stores.NewID(), VariantID: line.VariantID, Quantity: each.units, UnitPrice: line.Price})
	}
	filler.refused(filler.stores.Events.SellLines(eventID, lines, sales.Cash))
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
