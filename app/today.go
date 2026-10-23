package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/demo"
	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
	"github.com/Itros97/MUSA/sales"
)

// restockShown is how many products the home page lists as running low; the catalogue has the rest.
const restockShown = 4

// today is what the home page shows of the day, read when the page opens, not on every frame.
type today struct {
	date    calendar.Date
	fair    events.Event
	hasFair bool
	month   sales.Summary
	restock []catalogue.Product
	loadErr error
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
	day.month = sales.Summarise(sales.Filter(sold, sales.Query{Period: sales.ThisMonth}, day.date, time.Local))
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

func (program *App) drawToday(context *ui.Context) {
	day := program.today
	if day.loadErr != nil {
		context.Toast("No se pudo leer lo de hoy", true)
	}
	if !day.hasFair {
		return
	}
	theme := context.Theme
	happening := events.TimingOf(day.fair, day.date) == events.Happening
	look.Card(context, ui.Layout{Direction: look.Across(context), Width: ui.Grow(), Padding: ui.Pad(theme.Padding * 1.25), Gap: theme.Gap, AlignY: ui.Center}, func() {
		look.Badge(context, "calendar-days", sectionIcon)
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 4}, func() {
			look.Paragraph(context, fairSentence(day.fair, day.date, happening), ui.TextStyle{Size: theme.HeadingSize, Weight: draw2d.Bold})
			look.Paragraph(context, day.fair.WhenAndWhere(), ui.TextStyle{Color: theme.InkSoft})
		})
		action := "Ver la feria"
		if happening {
			action = "Ir al puesto"
		}
		if context.PrimaryIconButton("store", action) {
			program.Open(Fairs)
			program.fairs.OpenFair(day.fair.ID)
		}
	})
}

// fairSentence says in one sentence which fair is on: today's, or the next one and when.
func fairSentence(fair events.Event, date calendar.Date, happening bool) string {
	if happening {
		return "Hoy estás en " + fair.Name
	}
	return "Próxima feria: " + fair.Name + ", " + strings.ToLower(events.WhenWords(fair, date))
}

// drawMonth tells how the month goes in a sentence, and what is running out.
func (program *App) drawMonth(context *ui.Context) {
	theme := context.Theme
	day := program.today
	look.Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(theme.Padding * 1.25), Gap: theme.Gap}, func() {
		look.Paragraph(context, monthSentence(day.month), ui.TextStyle{Size: theme.HeadingSize, Weight: draw2d.Bold})
		if len(day.restock) == 0 {
			look.Paragraph(context, "Tienes de todo: nada se está acabando.", ui.TextStyle{Color: theme.InkSoft})
			return
		}
		context.Text("Se te está acabando:", ui.TextStyle{Color: theme.InkSoft})
		for index, product := range day.restock {
			if index == restockShown {
				break
			}
			program.restockRow(context, product)
		}
	})
}

// monthSentence is this month's sales in words: how much, in how many sales.
func monthSentence(month sales.Summary) string {
	switch month.Count {
	case 0:
		return "Este mes aún no has vendido nada"
	case 1:
		return "Este mes llevas " + money.Format(month.Revenue) + " en 1 venta"
	}
	return fmt.Sprintf("Este mes llevas %s en %d ventas", money.Format(month.Revenue), month.Count)
}

func (program *App) restockRow(context *ui.Context, product catalogue.Product) {
	theme := context.Theme
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 10, AlignY: ui.Center}, func() {
		context.IconSized("package", theme.IconSize, theme.InkSoft)
		context.Frame(ui.Layout{Width: ui.Grow()}, ui.Look{Clip: true}, func() {
			context.Text(product.Name, ui.TextStyle{})
		})
		colour := theme.InkSoft
		if product.TotalStock() == 0 {
			colour = theme.Danger
		}
		context.Text(restockWords(product.TotalStock()), ui.TextStyle{Color: colour})
	})
}

func restockWords(stock int) string {
	switch stock {
	case 0:
		return "No te queda ninguno"
	case 1:
		return "Te queda 1"
	}
	return fmt.Sprintf("Te quedan %d", stock)
}
