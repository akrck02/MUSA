// Package catalogue is what the artist sells: products, their variants with prices and stock, and
// the tags that group them. It checks drafts, filters the list, stores products and draws the
// catalogue's screens.
package catalogue

import (
	"time"

	"github.com/Itros97/MUSA/money"
)

// Kind is one of the two kinds of product the artist sells.
type Kind string

// The kinds of product. The words are the ones stored in the database.
const (
	// Physical products are stock the artist keeps and ships: prints, stickers, merch.
	Physical Kind = "PHYSICAL"
	// Digital products are made to order, such as commissions.
	Digital Kind = "DIGITAL"
)

// Status is where a product is in its life in the catalogue.
type Status string

// The statuses of a product. The words are the ones stored in the database.
const (
	Draft    Status = "DRAFT"
	Active   Status = "ACTIVE"
	Archived Status = "ARCHIVED"
)

// Product is anything the artist sells. It always has at least one variant, which carries the
// price and the cost.
type Product struct {
	ID          string
	Kind        Kind
	Name        string
	Description string
	Tags        []Tag
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Variants    []Variant

	// Material is what a physical product is made of ("cotton paper").
	Material string
	// DeliveryFormat is how a digital product is delivered ("PNG 4000 px").
	DeliveryFormat string
	// TurnaroundDays is how long a digital product takes to make; zero when not said.
	TurnaroundDays int
}

// Variant is a version of a product that is sold on its own (a size, a format, a level of
// detail), with its own price and cost. Stock, SKU and measures are only used by physical products;
// zero means "not said" for each of them except Stock.
type Variant struct {
	ID    string
	Name  string
	Price money.Money
	Cost  money.Money

	SKU               string
	Stock             int
	MinimumStock      int
	WeightGrams       int
	WidthMillimetres  int
	HeightMillimetres int
	DepthMillimetres  int
}

// LowStock tells whether the variant is sold out or down to the minimum the artist wants to keep.
func (variant Variant) LowStock() bool {
	return variant.Stock == 0 || (variant.MinimumStock > 0 && variant.Stock <= variant.MinimumStock)
}

// TotalStock is the units in stock across every variant.
func (product Product) TotalStock() int {
	total := 0
	for _, variant := range product.Variants {
		total += variant.Stock
	}
	return total
}

// LowStock tells whether a physical product has any variant that needs restocking.
func (product Product) LowStock() bool {
	if product.Kind != Physical {
		return false
	}
	for _, variant := range product.Variants {
		if variant.LowStock() {
			return true
		}
	}
	return false
}

// LowestPrice is the cheapest price among the variants, shown as "from …" in the list.
func (product Product) LowestPrice() money.Money {
	if len(product.Variants) == 0 {
		return 0
	}
	lowest := product.Variants[0].Price
	for _, variant := range product.Variants[1:] {
		lowest = min(lowest, variant.Price)
	}
	return lowest
}

// HasPriceRange tells whether the variants have different prices.
func (product Product) HasPriceRange() bool {
	for _, variant := range product.Variants {
		if variant.Price != product.LowestPrice() {
			return true
		}
	}
	return false
}

// Variant finds one of the product's variants by its id.
func (product Product) Variant(id string) (Variant, bool) {
	for _, variant := range product.Variants {
		if variant.ID == id {
			return variant, true
		}
	}
	return Variant{}, false
}

// ProductDraft is what the editor saves: the product as the artist left it, the tags by their
// names, and the stock each variant had when the editor opened (StockBaseline), so stock that moved
// elsewhere meanwhile is kept.
type ProductDraft struct {
	Product       Product
	TagNames      []string
	StockBaseline map[string]int
}
