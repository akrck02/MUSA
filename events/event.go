// Package events is the fairs and markets the artist sells at: the stock taken from the catalogue
// to the stand, the sales made there, what attending cost, and the profit left. Closing a fair
// brings what was not sold back to the catalogue.
package events

import (
	"strings"
	"time"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/money"
	"github.com/Itros97/MUSA/sales"
)

// Status tells whether a fair still takes stock and sales, or is finished and read-only.
type Status string

// The statuses. The words are the ones stored in the database.
const (
	Open   Status = "OPEN"
	Closed Status = "CLOSED"
)

// Event is a fair or a market.
type Event struct {
	ID        string
	Name      string
	Location  string
	StartDate calendar.Date
	EndDate   calendar.Date
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Draft is what the event editor saves; status and times are worked out on saving.
type Draft struct {
	ID        string
	Name      string
	Location  string
	StartDate calendar.Date
	EndDate   calendar.Date
}

// Detail is a fair with the stock taken to it, the sales made there and what it cost to attend.
type Detail struct {
	Event    Event
	Stock    []StockLine
	Sales    []sales.Sale
	Expenses []Expense
}

// StockLine is a physical variant taken to a fair: the units moved from the catalogue and the units
// sold there.
type StockLine struct {
	VariantID   string
	ProductName string
	VariantName string
	Price       money.Money
	Cost        money.Money
	Taken       int
	Sold        int
}

// Remaining is the units still on the stand.
func (line StockLine) Remaining() int { return line.Taken - line.Sold }

// LineOf finds the stock line of a variant; found is false when it was not taken.
func (detail Detail) LineOf(variantID string) (line StockLine, found bool) {
	for _, line := range detail.Stock {
		if line.VariantID == variantID {
			return line, true
		}
	}
	return StockLine{}, false
}

// Expense is money spent to attend a fair: the stand, travel, lodging.
type Expense struct {
	ID        string
	EventID   string
	Concept   string
	Amount    money.Money
	CreatedAt time.Time
}

// Summary is the totals of a fair.
type Summary struct {
	UnitsSold   int
	Gross       money.Money
	CostOfGoods money.Money
	Expenses    money.Money
}

// Profit is gross minus the cost of goods minus the expenses; negative when the fair lost money.
func (summary Summary) Profit() money.Money {
	return summary.Gross - summary.CostOfGoods - summary.Expenses
}

// Summarise adds up the sales and expenses of a fair.
func Summarise(detail Detail) Summary {
	totals := sales.Summarise(detail.Sales)
	return Summary{
		UnitsSold:   totals.Units,
		Gross:       totals.Revenue,
		CostOfGoods: totals.Cost,
		Expenses:    money.Sum(detail.Expenses, func(expense Expense) money.Money { return expense.Amount }),
	}
}

// Validate lists every problem of a draft; none means it can be saved.
func Validate(draft Draft) []Code {
	var problems []Code
	if strings.TrimSpace(draft.Name) == "" {
		problems = append(problems, CodeBlankName)
	}
	if draft.EndDate.Before(draft.StartDate) {
		problems = append(problems, CodeEndBeforeStart)
	}
	return problems
}

// WhenAndWhere is "18–19 abr 2026 · Pabellón 2", without the place when it is empty.
func (event Event) WhenAndWhere() string {
	words := calendar.RangeWords(event.StartDate, event.EndDate)
	if strings.TrimSpace(event.Location) == "" {
		return words
	}
	return words + " · " + event.Location
}
