package stand

import (
	"fmt"
	"strings"

	"github.com/Itros97/MUSA/money"
	"github.com/Itros97/MUSA/sales"
)

// Basket is what one customer is taking, in the order it was tapped.
type Basket struct {
	entries []entry
}

type entry struct {
	item     Item
	quantity int
}

// Add puts one more of an item in the basket; added is false when none of it is left to take.
func (basket *Basket) Add(item Item) (added bool) {
	for index := range basket.entries {
		if basket.entries[index].item.VariantID != item.VariantID {
			continue
		}
		if basket.entries[index].quantity >= item.Left {
			return false
		}
		basket.entries[index].quantity++
		return true
	}
	if item.Left <= 0 {
		return false
	}
	basket.entries = append(basket.entries, entry{item: item, quantity: 1})
	return true
}

// Remove takes one of a variant out of the basket.
func (basket *Basket) Remove(variantID string) {
	for index := range basket.entries {
		if basket.entries[index].item.VariantID != variantID {
			continue
		}
		basket.entries[index].quantity--
		if basket.entries[index].quantity == 0 {
			basket.entries = append(basket.entries[:index], basket.entries[index+1:]...)
		}
		return
	}
}

// Clear empties the basket.
func (basket *Basket) Clear() { basket.entries = nil }

// Quantity is how many of a variant are in the basket.
func (basket Basket) Quantity(variantID string) int {
	for _, each := range basket.entries {
		if each.item.VariantID == variantID {
			return each.quantity
		}
	}
	return 0
}

// Count is how many pieces the basket holds.
func (basket Basket) Count() int {
	count := 0
	for _, each := range basket.entries {
		count += each.quantity
	}
	return count
}

// Total is what the basket costs the customer.
func (basket Basket) Total() money.Money {
	return money.Sum(basket.entries, func(each entry) money.Money { return each.item.Price.Times(each.quantity) })
}

// Lines is the basket as the lines of a sale, each with a new id.
func (basket Basket) Lines(newID func() string) []sales.LineDraft {
	lines := make([]sales.LineDraft, len(basket.entries))
	for index, each := range basket.entries {
		lines[index] = sales.LineDraft{ID: newID(), VariantID: each.item.VariantID, Quantity: each.quantity, UnitPrice: each.item.Price}
	}
	return lines
}

// Words lists what the basket holds as the artist reads it: "2 × Luna · Mate, 1 × Rana feliz".
func (basket Basket) Words() string {
	parts := make([]string, len(basket.entries))
	for index, each := range basket.entries {
		parts[index] = fmt.Sprintf("%d × %s", each.quantity, each.item.Label)
	}
	return strings.Join(parts, ", ")
}

// Taken is one line of what a customer took, as a receipt reads it: "2 × Luna · Mate" and what
// those cost together.
type Taken struct {
	Words  string
	Amount money.Money
}

// Taken lists what the basket holds, a line per item, for the receipt.
func (basket Basket) Taken() []Taken {
	taken := make([]Taken, len(basket.entries))
	for index, each := range basket.entries {
		taken[index] = Taken{Words: fmt.Sprintf("%d × %s", each.quantity, each.item.Label), Amount: each.item.Price * money.Money(each.quantity)}
	}
	return taken
}
