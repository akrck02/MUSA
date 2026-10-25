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

// Measures of the history, in pixels, as in the author's design.
const (
	periodWidth       = 200
	periodChevronSize = 14
	periodChevronGap  = 4
	periodListPadding = 6
	periodRowPaddingX = 12
	periodRowPaddingY = 10
	filtersBottom     = 14
	figuresGap        = 28
	figuresBottom     = 14
	figureLabelSize   = 15
	figureAmountSize  = 20
	figureSpace       = 6
	historyPadX       = 20 + 1
	historyPadY       = 6 + 1
	historyRadius     = 18
	rowPaddingY       = 14
	rowGap            = 14
	dateWidth         = 120
	spareWidth        = 160 // an empty column of the design, kept so amounts line up as there
	amountWidth       = 110
	daySize           = 15
	amountSize        = 17
)

// filterTabs are the tabs of the history's kinds.
var filterTabs = look.PillSize{TextSize: 16, PaddingX: 14, PaddingY: 8, Between: 8}

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
	look.Space(context, filtersBottom)
	context.Box(ui.Layout{Gap: figuresGap, AlignY: ui.End}, func() {
		screens.periodFigure(context, money.Format(summary.Revenue))
		inlineFigure(context, "Coste", money.Format(summary.Cost))
		inlineFigure(context, "Te queda", money.Format(summary.Margin()))
	})
	look.Space(context, figuresBottom)
	if len(shown) == 0 {
		look.Empty(context, "receipt", "Sin ventas aquí", "Mira otro periodo, u otro filtro.", "")
		return
	}
	theme := context.Theme
	card := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: historyRadius}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.PadXY(historyPadX, historyPadY)}, card, func() {
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
	context.Box(ui.Layout{Gap: figureSpace, AlignY: ui.End}, func() {
		context.Text(label, ui.TextStyle{Size: figureLabelSize, Color: theme.InkSoft})
		context.Text(amount, ui.TextStyle{Size: figureAmountSize, Weight: draw2d.Bold})
	})
}

// historyFilters are the kinds as tabs on the left, and the period on the right.
func (screens *Screens) historyFilters(context *ui.Context) {
	tabs := make([]look.Tab, len(historyKinds))
	for index, kind := range historyKinds {
		tabs[index] = look.Tab{Label: kind.words}
	}
	look.PillTabs(context, "kind", &screens.kind, tabs, filterTabs)
}

// periodFigure is the period's revenue, named by the period: tapping its name lists the other
// periods to choose from.
func (screens *Screens) periodFigure(context *ui.Context, amount string) {
	theme := context.Theme
	label := periodWords[Periods[screens.period]]
	context.Box(ui.Layout{Gap: figureSpace, AlignY: ui.End}, func() {
		if context.Clickable("Periodo", ui.Layout{Gap: periodChevronGap, AlignY: ui.Center}, ui.Looks{}, func() {
			context.Text(label, ui.TextStyle{Size: figureLabelSize, Color: theme.InkSoft})
			context.IconSized("chevron-down", periodChevronSize, theme.InkSoft)
		}) {
			screens.choosingPeriod = !screens.choosingPeriod
		}
		context.Describe(ui.Description{Role: ui.ButtonRole, Label: "Periodo"})
		context.Popover("periods", &screens.choosingPeriod, ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1,
			Radius: theme.ControlRadius, Shadow: true}, func() {
			screens.periodChoices(context)
		})
		context.Text(amount, ui.TextStyle{Size: figureAmountSize, Weight: draw2d.Bold})
	})
}

// periodChoices are the periods the history can show, one per line.
func (screens *Screens) periodChoices(context *ui.Context) {
	theme := context.Theme
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(periodWidth), Padding: ui.Pad(periodListPadding)}, func() {
		for index, period := range Periods {
			hovered := ui.Look{Fill: theme.Sunken, Radius: theme.ControlRadius}
			row := ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(periodRowPaddingX, periodRowPaddingY)}
			if context.Clickable(periodWords[period], row, ui.Looks{Hovered: hovered, Pressed: hovered}, func() {
				context.Text(periodWords[period], ui.TextStyle{Weight: chosenWeight(index == screens.period)})
			}) {
				screens.period, screens.choosingPeriod = index, false
			}
			context.Describe(ui.Description{Role: ui.ButtonRole, Label: periodWords[period]})
		}
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
	layout := ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(0, rowPaddingY), Gap: rowGap, AlignY: ui.Center}
	// No pressed fill: the sale's dialog answers the tap, and a fill would linger under it.
	clicked := context.Clickable(label, layout, ui.Looks{}, func() {
		context.Box(ui.Layout{Width: ui.Fixed(dateWidth)}, func() {
			context.Text(day, ui.TextStyle{Size: daySize, Color: theme.InkSoft})
		})
		nameAndDetail(context, sale, sourceWords(sale), screens.who(sale), SummaryWords(sale))
		context.Box(ui.Layout{Width: ui.Fixed(spareWidth)}, nil)
		context.Box(ui.Layout{Width: ui.Fixed(amountWidth), AlignX: ui.End}, func() {
			context.Text(money.Format(sale.Total()), ui.TextStyle{Size: amountSize, Weight: draw2d.Bold})
		})
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

// chosenWeight is bold for the option chosen, and the normal weight for the rest.
func chosenWeight(chosen bool) draw2d.Weight {
	if chosen {
		return draw2d.Bold
	}
	return draw2d.Regular
}
