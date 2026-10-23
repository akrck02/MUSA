// Package sales is every order the artist makes, whatever the channel: its lines, what shipping
// was charged and cost, and the margin left. Recording a sale outside a fair takes its units from
// the catalogue; deleting it gives them back.
package sales

import (
	"time"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/money"
)

// Channel is where a sale was made.
type Channel string

// The channels. The words are the ones stored in the database.
const (
	// AtEvent sales are made at a fair, and only recorded from its event.
	AtEvent     Channel = "EVENT"
	OnlineShop  Channel = "ONLINE_SHOP"
	Social      Channel = "SOCIAL"
	Marketplace Channel = "MARKETPLACE"
	InPerson    Channel = "IN_PERSON"
	Other       Channel = "OTHER"
)

// SelectableChannels are the channels the order editor offers: every one but AtEvent.
var SelectableChannels = []Channel{OnlineShop, Social, Marketplace, InPerson, Other}

// Status is how far an order has travelled. It only moves forward.
type Status string

// The statuses, in order. The words are the ones stored in the database.
const (
	ToPrepare Status = "TO_PREPARE"
	Sent      Status = "SENT"
	Delivered Status = "DELIVERED"
)

// IsLocal tells whether a channel hands the goods over on the spot, at a fair or in person: its
// sales never travel, so they are delivered as they are made.
func IsLocal(channel Channel) bool { return channel == AtEvent || channel == InPerson }

// StatusOfNew is the status an order starts with: delivered when it is local, to prepare otherwise.
func StatusOfNew(channel Channel) Status {
	if IsLocal(channel) {
		return Delivered
	}
	return ToPrepare
}

// Next is the status after this one; found is false for a delivered order, which has none.
func (status Status) Next() (next Status, found bool) {
	switch status {
	case ToPrepare:
		return Sent, true
	case Sent:
		return Delivered, true
	}
	return "", false
}

// Method is how a sale was paid.
type Method string

// The methods. The words are the ones stored in the database; empty means not said.
const (
	Cash Method = "CASH"
	Card Method = "CARD"
)

// Sale is an order: one or more lines sold together through a channel.
type Sale struct {
	ID       string
	Channel  Channel
	Status   Status
	PaidWith Method // empty when not said
	EventID  string // empty outside fairs
	Customer string
	Note     string
	Lines    []Line
	// ShippingCharged is what the customer paid for shipping; ShippingCost what it cost the artist.
	ShippingCharged money.Money
	ShippingCost    money.Money
	SoldAt          time.Time
}

// Line is the units of one variant in a sale. Names, price and cost are copies, so editing the
// product later does not change a sale already made.
type Line struct {
	ID          string
	VariantID   string
	ProductName string
	VariantName string
	Quantity    int
	UnitPrice   money.Money
	UnitCost    money.Money
}

// Total is what the line was sold for.
func (line Line) Total() money.Money { return line.UnitPrice.Times(line.Quantity) }

// TotalCost is what the line's goods cost.
func (line Line) TotalCost() money.Money { return line.UnitCost.Times(line.Quantity) }

// Units is the units sold across every line.
func (sale Sale) Units() int {
	units := 0
	for _, line := range sale.Lines {
		units += line.Quantity
	}
	return units
}

// Total is what the customer paid: the lines and the shipping charged.
func (sale Sale) Total() money.Money { return money.Sum(sale.Lines, Line.Total) + sale.ShippingCharged }

// Cost is what the sale cost the artist: the goods and the shipping paid.
func (sale Sale) Cost() money.Money { return money.Sum(sale.Lines, Line.TotalCost) + sale.ShippingCost }

// Margin is the total minus the cost; negative when the sale lost money.
func (sale Sale) Margin() money.Money { return sale.Total() - sale.Cost() }

// Draft is what the order editor saves; names and costs are copied from the catalogue on saving.
type Draft struct {
	Channel         Channel
	PaidWith        Method
	Customer        string
	Note            string
	Lines           []LineDraft
	ShippingCharged money.Money
	ShippingCost    money.Money
	SoldAt          time.Time
}

// LineDraft is one line of the order editor.
type LineDraft struct {
	ID        string
	VariantID string
	Quantity  int
	UnitPrice money.Money
}

// noon is the hour a sale made on another day is stored at, so no time zone moves it to the next
// or the previous day.
const noon = 12

// SoldAt is the moment a sale picked for day is stored at: now when day is today, its noon otherwise.
func SoldAt(day calendar.Date, now time.Time) time.Time {
	if day == calendar.Of(now) {
		return now
	}
	return day.At(noon, now.Location())
}
