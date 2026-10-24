package app

import (
	"time"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/demo"
	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/sales"
)

// today is what the home page shows of the day, read when the page opens, not on every frame.
type today struct {
	date    calendar.Date
	fair    events.Event
	hasFair bool
	month   sales.Summary
	// projectsDelivered counts the month's sales that are delivered projects.
	projectsDelivered int
	restock           []catalogue.Product
	plan              agenda
	loadErr           error
}

// readToday gathers the fair to look at (the one happening, or else the next one), this month's
// sales and the physical products running low.
func readToday(stores demo.Stores, now time.Time) today {
	day := today{date: calendar.Of(now)}
	fairs, err := stores.Events.All()
	if err != nil {
		day.loadErr = err
		return day
	}
	day.fair, day.hasFair = fairToWatch(fairs, day.date)
	sold, err := stores.Sales.All()
	if err != nil {
		day.loadErr = err
		return day
	}
	commissions, err := stores.Projects.All()
	if err != nil {
		day.loadErr = err
		return day
	}
	day.plan = agenda{fairs: fairs, onTheWay: sales.OnTheWay(sold), projects: commissions, zone: now.Location()}
	thisMonth := sales.Filter(sold, sales.Query{Period: sales.ThisMonth}, day.date, time.Local)
	day.month = sales.Summarise(thisMonth)
	day.projectsDelivered = len(sales.OfKind(thisMonth, func(channel sales.Channel) bool { return channel == sales.Commission }))
	products, err := stores.Products.All()
	if err != nil {
		day.loadErr = err
		return day
	}
	for _, product := range products {
		if product.Status != catalogue.Archived && product.LowStock() {
			day.restock = append(day.restock, product)
		}
	}
	return day
}

// fairToWatch is the fair happening today or, if none is, the soonest one to come.
func fairToWatch(fairs []events.Event, date calendar.Date) (events.Event, bool) {
	var next events.Event
	found := false
	for _, fair := range fairs {
		switch events.TimingOf(fair, date) {
		case events.Happening:
			return fair, true
		case events.Upcoming:
			if !found || fair.StartDate.Before(next.StartDate) {
				next, found = fair, true
			}
		}
	}
	return next, found
}
