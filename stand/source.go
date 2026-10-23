// Package stand is the stand: the screen the artist sells from, at a fair or at home, with a
// basket, the total and «Cobrar». What it sells comes from a Source: the stock taken to one fair,
// or the whole catalogue.
package stand

import (
	"time"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/money"
	"github.com/Itros97/MUSA/sales"
)

// Item is one thing on the stand: a variant, its price and how many are left to sell.
type Item struct {
	VariantID string
	Label     string // the product and the version, for a sale's line: "Luna · Mate"
	// Name is the product's: it paints the item's picture until products have photos, the same
	// for every version, so the artist learns it.
	Name    string
	Version string // empty for a product with one version
	Price   money.Money
	Left    int
	Tags    []string
}

// Source is where a stand takes what it sells from, and where it records a sale. Refusal is the
// artist's words when the sale is turned down (not enough left, a closed fair); err is a failure.
type Source interface {
	Name() string
	Items() ([]Item, error)
	Sell(lines []sales.LineDraft, method sales.Method) (refusal string, err error)
}

// Fair sells the stock taken to one open fair; its sales belong to the fair.
type Fair struct {
	Event    events.Event
	Events   *events.Store
	Products *catalogue.Store
}

// Name is the fair's name.
func (fair Fair) Name() string { return fair.Event.Name }

// Items is what is on the fair's stand, with the tags of each product.
func (fair Fair) Items() ([]Item, error) {
	detail, err := fair.Events.Detail(fair.Event.ID)
	if err != nil {
		return nil, err
	}
	var items []Item
	for _, line := range detail.Stock {
		product, _, err := fair.Products.FindByVariant(line.VariantID)
		if err != nil {
			return nil, err
		}
		items = append(items, Item{VariantID: line.VariantID, Label: sales.ItemLabel(line.ProductName, line.VariantName),
			Name: line.ProductName, Version: line.VariantName,
			Price: line.Price, Left: line.Remaining(), Tags: tagNames(product.Tags)})
	}
	return items, nil
}

// Sell records the basket as one of the fair's sales.
func (fair Fair) Sell(lines []sales.LineDraft, method sales.Method) (string, error) {
	refused, err := fair.Events.SellLines(fair.Event.ID, lines, method)
	if err != nil || len(refused) == 0 {
		return "", err
	}
	return events.RefusalWords(refused[0]), nil
}

// WholeStock sells straight from the catalogue, at home or in a shop: a local sale.
type WholeStock struct {
	Sales    *sales.Store
	Products *catalogue.Store
	Now      func() time.Time
}

// Name is what the artist calls selling from the whole catalogue.
func (WholeStock) Name() string { return "Todo el stock" }

// Items is every active physical variant with stock in the catalogue.
func (stock WholeStock) Items() ([]Item, error) {
	products, err := stock.Products.All()
	if err != nil {
		return nil, err
	}
	var items []Item
	for _, product := range products {
		if product.Kind != catalogue.Physical || product.Status != catalogue.Active {
			continue
		}
		for _, variant := range product.Variants {
			if variant.Stock > 0 {
				items = append(items, Item{VariantID: variant.ID, Label: sales.ItemLabel(product.Name, variant.Name),
					Name: product.Name, Version: variant.Name,
					Price: variant.Price, Left: variant.Stock, Tags: tagNames(product.Tags)})
			}
		}
	}
	return items, nil
}

// Sell records the basket as a local sale, delivered at once.
func (stock WholeStock) Sell(lines []sales.LineDraft, method sales.Method) (string, error) {
	now := stock.Now()
	_, problems, err := stock.Sales.Record(sales.Draft{Channel: sales.InPerson, PaidWith: method, Lines: lines,
		SoldAt: sales.SoldAt(calendar.Of(now), now)})
	if err != nil || len(problems) == 0 {
		return "", err
	}
	return sales.ProblemWords(problems[0]), nil
}

func tagNames(tags []catalogue.Tag) []string {
	names := make([]string, len(tags))
	for index, tag := range tags {
		names[index] = tag.Name
	}
	return names
}
