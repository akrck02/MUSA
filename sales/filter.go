package sales

import (
	"slices"
	"strings"
	"time"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/money"
)

// Validate lists every problem of a draft; none means it can be recorded. Stock is checked when
// recording, against the catalogue.
func Validate(draft Draft) []Problem {
	var problems []Problem
	if len(draft.Lines) == 0 {
		problems = append(problems, Problem{Code: CodeNoLines})
	}
	if draft.Channel == AtEvent {
		problems = append(problems, Problem{Code: CodeEventChannel})
	}
	if draft.ShippingCharged < 0 || draft.ShippingCost < 0 {
		problems = append(problems, Problem{Code: CodeNegativeShipping})
	}
	for _, line := range draft.Lines {
		if line.Quantity <= 0 {
			problems = append(problems, Problem{Code: CodeInvalidQuantity, LineID: line.ID})
		}
		if line.UnitPrice < 0 {
			problems = append(problems, Problem{Code: CodeNegativePrice, LineID: line.ID})
		}
	}
	return problems
}

// Period is the stretch of time the sales list shows.
type Period string

// The periods.
const (
	ThisMonth Period = "THIS_MONTH"
	LastMonth Period = "LAST_MONTH"
	ThisYear  Period = "THIS_YEAR"
	AllTime   Period = "ALL"
)

// Periods lists every period, in the order the list offers them.
var Periods = []Period{ThisMonth, LastMonth, ThisYear, AllTime}

// Query is the filters of the sales list; an empty Channel means every channel.
type Query struct {
	Period  Period
	Channel Channel
}

// Filter keeps the sales of the query's period, counted from today, and channel, newest first.
func Filter(sales []Sale, query Query, today calendar.Date, zone *time.Location) []Sale {
	first, last, bounded := rangeOf(query.Period, today)
	var kept []Sale
	for _, sale := range sales {
		if query.Channel != "" && sale.Channel != query.Channel {
			continue
		}
		day := calendar.Of(sale.SoldAt.In(zone))
		if bounded && (day.Before(first) || !day.Before(last)) {
			continue
		}
		kept = append(kept, sale)
	}
	slices.SortStableFunc(kept, func(first, second Sale) int { return second.SoldAt.Compare(first.SoldAt) })
	return kept
}

// rangeOf is the first day a period includes and the first it leaves out; bounded is false for
// all time.
func rangeOf(period Period, today calendar.Date) (first, last calendar.Date, bounded bool) {
	monthStart := calendar.Date{Year: today.Year, Month: today.Month, Day: 1}
	yearStart := calendar.Date{Year: today.Year, Month: time.January, Day: 1}
	switch period {
	case ThisMonth:
		return monthStart, monthStart.AddMonths(1), true
	case LastMonth:
		return monthStart.AddMonths(-1), monthStart, true
	case ThisYear:
		return yearStart, calendar.Date{Year: today.Year + 1, Month: time.January, Day: 1}, true
	default:
		return calendar.Date{}, calendar.Date{}, false
	}
}

// Summary is the totals of a set of sales.
type Summary struct {
	Count   int
	Units   int
	Revenue money.Money
	Cost    money.Money
}

// Margin is the revenue minus the cost.
func (summary Summary) Margin() money.Money { return summary.Revenue - summary.Cost }

// Summarise adds up sales.
func Summarise(sales []Sale) Summary {
	summary := Summary{Count: len(sales)}
	for _, sale := range sales {
		summary.Units += sale.Units()
		summary.Revenue += sale.Total()
		summary.Cost += sale.Cost()
	}
	return summary
}

// ItemLabel is "Lámina gato · A4", or only the product's name when the variant has none.
func ItemLabel(productName, variantName string) string {
	if strings.TrimSpace(variantName) == "" {
		return productName
	}
	return productName + " · " + variantName
}

// OnTheWay keeps the orders still to be delivered, oldest first: the next to prepare leads.
func OnTheWay(sales []Sale) []Sale {
	var kept []Sale
	for _, sale := range sales {
		if sale.Status != Delivered {
			kept = append(kept, sale)
		}
	}
	slices.SortStableFunc(kept, func(first, second Sale) int { return first.SoldAt.Compare(second.SoldAt) })
	return kept
}

// Done keeps the delivered orders, local sales among them: the sales history.
func Done(sales []Sale) []Sale {
	var kept []Sale
	for _, sale := range sales {
		if sale.Status == Delivered {
			kept = append(kept, sale)
		}
	}
	return kept
}

// UnitsSince counts the units of each version, by its ID, that the sales made at or after since
// sold.
func UnitsSince(sales []Sale, since time.Time) map[string]int {
	units := map[string]int{}
	for _, sale := range sales {
		if sale.SoldAt.Before(since) {
			continue
		}
		for _, line := range sale.Lines {
			units[line.VariantID] += line.Quantity
		}
	}
	return units
}
