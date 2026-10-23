package sales

import (
	"fmt"
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
)

// Widths of the history, in pixels: its date column and its period's choice.
const (
	dateWidth   = 90
	periodWidth = 200
)

// historyKind is a group of channels the history filters by.
type historyKind struct {
	words string
	keeps func(Channel) bool
}

// historyKinds are the filters of the history: everything, what was handed over on the spot, and
// what was sent.
var historyKinds = []historyKind{
	{"Todo", func(Channel) bool { return true }},
	{"Ventas locales", IsLocal},
	{"Envíos", func(channel Channel) bool { return !IsLocal(channel) && channel != Commission }},
	{"Proyectos", func(channel Channel) bool { return channel == Commission }},
}

// OfKind keeps the sales whose channel keeps answers yes to.
func OfKind(sales []Sale, keeps func(Channel) bool) []Sale {
	var kept []Sale
	for _, sale := range sales {
		if keeps(sale.Channel) {
			kept = append(kept, sale)
		}
	}
	return kept
}

// drawHistory is the delivered orders, local sales among them, of a period and a kind, in one card.
func (screens *Screens) drawHistory(context *ui.Context) {
	now := screens.now()
	period := Query{Period: Periods[screens.period]}
	shown := OfKind(Filter(Done(screens.sales), period, calendar.Of(now), now.Location()), historyKinds[screens.kind].keeps)
	screens.historyFilters(context)
	summary := Summarise(shown)
	context.Box(ui.Layout{Gap: 24, AlignY: ui.End}, func() {
		inlineFigure(context, periodWords[Periods[screens.period]], money.Format(summary.Revenue))
		inlineFigure(context, "Coste", money.Format(summary.Cost))
		inlineFigure(context, "Te queda", money.Format(summary.Margin()))
	})
	if len(shown) == 0 {
		look.Empty(context, "receipt", "Sin ventas aquí", "Mira otro periodo, u otro filtro.", "")
		return
	}
	look.Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.PadXY(16, 4)}, func() {
		for index, sale := range shown {
			if index > 0 {
				context.Separator()
			}
			context.Push(sale.ID)
			if screens.historyRow(context, sale) {
				screens.opened = sale.ID
			}
			context.Pop()
		}
	})
}

// inlineFigure is a small label and its amount on one line: "Coste 52 €".
func inlineFigure(context *ui.Context, label, amount string) {
	theme := context.Theme
	context.Box(ui.Layout{Gap: 8, AlignY: ui.Center}, func() {
		context.Text(label, ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
		context.Text(amount, ui.TextStyle{Size: theme.TextSize * 1.3, Weight: draw2d.Bold})
	})
}

// historyFilters are the kinds as tabs on the left, and the period on the right.
func (screens *Screens) historyFilters(context *ui.Context) {
	var kinds, periods []string
	for _, kind := range historyKinds {
		kinds = append(kinds, kind.words)
	}
	for _, period := range Periods {
		periods = append(periods, periodWords[period])
	}
	context.Box(ui.Layout{Width: ui.Grow(), Gap: context.Theme.Gap, AlignY: ui.Center}, func() {
		tabs := make([]look.Tab, len(kinds))
		for index, words := range kinds {
			tabs[index] = look.Tab{Label: words}
		}
		context.Box(ui.Layout{Width: ui.Grow()}, func() { look.PillTabs(context, "kind", &screens.kind, tabs) })
		context.Box(ui.Layout{Width: ui.Fixed(periodWidth)}, func() { context.Select("period", &screens.period, periods) })
	})
}

// historyRow is one sale: the day, who and where from, what, and the amount; it returns true on
// the frame it is clicked.
func (screens *Screens) historyRow(context *ui.Context, sale Sale) bool {
	theme := context.Theme
	day := calendar.DayWords(calendar.Of(sale.SoldAt.In(screens.now().Location())))
	if calendar.Of(sale.SoldAt.In(screens.now().Location())) == calendar.Of(screens.now()) {
		day = "Hoy"
	}
	label := fmt.Sprintf("%s · %s · %s", day, screens.who(sale), money.Format(sale.Total()))
	layout := ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(0, 12), Gap: 16, AlignY: ui.Center}
	clicked := context.Clickable(label, layout, ui.Looks{Pressed: ui.Look{Fill: theme.Sunken}}, func() {
		context.Box(ui.Layout{Width: ui.Fixed(dateWidth)}, func() {
			context.Text(day, ui.TextStyle{Color: theme.InkSoft})
		})
		context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 4}, ui.Look{Clip: true}, func() {
			context.Box(ui.Layout{Gap: 8, AlignY: ui.Center}, func() {
				context.Text(screens.who(sale), ui.TextStyle{Weight: draw2d.Bold})
				look.Pill(context, sourceWords(sale), sourceTone(theme, sale.Channel))
			})
			context.Text(screens.headline(sale), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
		})
		context.Text(money.Format(sale.Total()), ui.TextStyle{Weight: draw2d.Bold})
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: label})
	return clicked
}

// sourceWords is the pill of a sale: local sales say so, the rest say their channel, or the shop
// written in the note («Etsy»).
func sourceWords(sale Sale) string {
	switch {
	case IsLocal(sale.Channel):
		return "Venta local"
	case sale.Channel == Marketplace && sale.Note != "" && !strings.Contains(sale.Note, " "):
		return sale.Note
	}
	return channelWords[sale.Channel]
}

// who is who a sale went to: the fair's name for a fair's sales, the customer or the channel otherwise.
func (screens *Screens) who(sale Sale) string {
	if sale.EventID != "" && screens.eventName != nil {
		if name := screens.eventName(sale.EventID); name != "" {
			return "Venta en el " + name
		}
	}
	return whoOrdered(sale)
}
