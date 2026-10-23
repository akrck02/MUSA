package money

// percentInWhole is how many percent make the whole price.
const percentInWhole = 100

// Margin is what the artist earns per unit: price minus cost, and that profit as a share of the price.
type Margin struct {
	Amount Money
	// Percent is the profit as a whole percentage of the price; HasPercent is false when the price is zero.
	Percent    int
	HasPercent bool
}

// MarginOf is the margin of selling at price what cost cost.
func MarginOf(price, cost Money) Margin {
	amount := price - cost
	if price <= 0 {
		return Margin{Amount: amount}
	}
	return Margin{Amount: amount, Percent: int(amount * percentInWhole / price), HasPercent: true}
}

// Negative tells whether selling loses money.
func (margin Margin) Negative() bool { return margin.Amount < 0 }
