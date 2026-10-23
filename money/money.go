// Package money keeps amounts of euros as whole cents, so adding prices never rounds, and reads
// and writes them the Spanish way ("1.234,50 €").
package money

// Money is an amount of euros in whole cents: Money(1250) is 12,50 €. Negative amounts are losses.
type Money int64

// Times is the amount multiplied by a number of units.
func (amount Money) Times(quantity int) Money { return amount * Money(quantity) }

// Sum adds up the amount each item gives; an empty list adds up to zero.
func Sum[Item any](items []Item, amountOf func(Item) Money) Money {
	var total Money
	for _, item := range items {
		total += amountOf(item)
	}
	return total
}
