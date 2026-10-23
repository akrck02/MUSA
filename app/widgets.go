package app

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
	"github.com/Itros97/MUSA/stand"
)

// Measures of the home page's widgets, in pixels.
const (
	widgetHeight = 280
	widgetBadge  = 64
	bigFigure    = 2.6 // the month's figure, in times the text size
	restockRows  = 3
)

// widget is one card of the home page: its content on top and its links at the bottom, all the
// same height so a row of them reads as one.
func widget(context *ui.Context, content, links func()) {
	theme := context.Theme
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Fixed(widgetHeight), Padding: ui.Pad(theme.Padding * 1.25), Gap: 10}
	look.Card(context, layout, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Gap: 6}, content)
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.Center}, links)
	})
}

// link is a quiet "see more" button at the bottom of a widget.
func link(context *ui.Context, words string) bool {
	return context.FlatIconButton("arrow-right", words)
}

// fairWidget is the fair to look at: today's, with a button to the stand, or the next one.
func (program *App) fairWidget(context *ui.Context) {
	theme := context.Theme
	day := program.today
	if !day.hasFair {
		widget(context, func() {
			look.Badge(context, "tent", widgetBadge)
			context.Text("Sin ferias a la vista", ui.TextStyle{Size: theme.HeadingSize, Weight: draw2d.Bold})
		}, func() {
			if link(context, "Nueva feria") {
				program.Open(Fairs)
				program.fairs.StartFair(day.date)
			}
		})
		return
	}
	happening := events.TimingOf(day.fair, day.date) == events.Happening
	label := "PRÓXIMA FERIA"
	if happening {
		label = "HOY ESTÁS EN"
	}
	widget(context, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 14, AlignY: ui.Center}, func() {
			look.Badge(context, "tent", widgetBadge)
			context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, func() {
				context.Text(label, ui.TextStyle{Size: theme.SmallTextSize, Weight: draw2d.Bold, Color: theme.InkSoft})
				look.Paragraph(context, day.fair.Name, ui.TextStyle{Size: theme.HeadingSize, Weight: draw2d.Bold})
				look.Paragraph(context, events.WhenWords(day.fair, day.date)+" · "+day.fair.Location, ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
			})
		})
	}, func() {
		program.fairLinks(context, happening)
	})
}

func (program *App) fairLinks(context *ui.Context, happening bool) {
	fair := program.today.fair
	if link(context, "Ver la feria") {
		program.Open(Fairs)
		program.fairs.OpenFair(fair.ID)
	}
	context.Spacer()
	if happening && context.PrimaryIconButton("store", "Ir al puesto") {
		program.Open(Stand)
		program.stand.Choose(stand.Fair{Event: fair, Events: program.stores.Events, Products: program.stores.Products})
	}
	if !happening && context.PrimaryButton("Preparar el puesto") {
		program.Open(Fairs)
		program.fairs.OpenFair(fair.ID)
	}
}

// monthWidget is this month's sales in one big figure.
func (program *App) monthWidget(context *ui.Context) {
	theme := context.Theme
	month := program.today.month
	widget(context, func() {
		context.Text("Este mes llevas", ui.TextStyle{Color: theme.InkSoft})
		context.Text(money.Format(month.Revenue), ui.TextStyle{Size: theme.TextSize * bigFigure, Weight: draw2d.Bold})
		context.Text(salesWords(month.Count), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
	}, func() {
		if link(context, "Ver ventas") {
			program.Open(Sales)
			program.sales.ShowHistory()
		}
	})
}

func salesWords(count int) string {
	if count == 1 {
		return "en 1 venta"
	}
	return fmt.Sprintf("en %d ventas", count)
}

// restockWidget is the products running low, the emptiest in red.
func (program *App) restockWidget(context *ui.Context) {
	theme := context.Theme
	restock := program.today.restock
	widget(context, func() {
		context.Text("Te quedan pocos", ui.TextStyle{Color: theme.InkSoft})
		if len(restock) == 0 {
			look.Paragraph(context, "Tienes de todo: nada se está acabando.", ui.TextStyle{})
		}
		for index, product := range restock {
			if index == restockRows {
				break
			}
			if index > 0 {
				context.Separator()
			}
			context.Box(ui.Layout{Width: ui.Grow(), Gap: 10, AlignY: ui.Center}, func() {
				context.Frame(ui.Layout{Width: ui.Grow()}, ui.Look{Clip: true}, func() { context.Text(product.Name, ui.TextStyle{}) })
				colour := theme.Warning
				if product.TotalStock() == 0 {
					colour = theme.Danger
				}
				context.Text(fmt.Sprint(product.TotalStock()), ui.TextStyle{Weight: draw2d.Bold, Color: colour})
			})
		}
	}, func() {
		if len(restock) > 0 && link(context, fmt.Sprintf("Ver los %d", len(restock))) {
			program.Open(Catalogue)
		}
	})
}
