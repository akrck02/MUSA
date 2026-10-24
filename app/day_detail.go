package app

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/look"
)

// Measures of a day's detail, in pixels.
const (
	dayDetailWidth   = 420
	dayDetailPadding = 24
	dayDetailRadius  = 20
	dayDetailTitle   = 22
	dayDetailNothing = 15
	dayDetailBottom  = 16
)

// drawDayDetail is the day tapped on the calendar, over it: everything that happens that day, each
// to be opened, and a button to put a fair on it.
func (program *App) drawDayDetail(context *ui.Context) {
	day := program.home.dayShown
	if day.Zero() {
		return
	}
	theme := context.Theme
	shown := context.Appear("day "+calendar.Words(day), look.DialogMotion)
	options := ui.LayerOptions{AlignX: ui.Center, AlignY: ui.Center, Width: ui.Fixed(dayDetailWidth),
		Hidden: 1 - shown, Leave: look.Leaving(look.DialogRise)}
	card := look.Arriving(ui.Look{Fill: theme.Surface, Radius: dayDetailRadius, Shadow: true}, shown, look.DialogRise)
	if context.Modal("day detail", options, func() {
		context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(dayDetailPadding)}, card, func() {
			program.dayDetailContent(context, day)
		})
	}) {
		program.home.dayShown = calendar.Date{}
	}
}

// dayDetailContent is the day's name, its entries, and «Nueva feria» on it.
func (program *App) dayDetailContent(context *ui.Context, day calendar.Date) {
	theme := context.Theme
	context.Text(calendar.LongWords(day), ui.TextStyle{Size: dayDetailTitle, Weight: draw2d.Bold})
	entries := program.today.plan.on(day)
	if len(entries) == 0 {
		look.Space(context, chipTop)
		context.Text("Nada este día.", ui.TextStyle{Size: dayDetailNothing, Color: theme.InkSoft})
	}
	for _, entry := range entries {
		if entryChip(context, entry) {
			program.home.dayShown = calendar.Date{}
			program.openEntry(entry)
		}
	}
	look.Space(context, dayDetailBottom)
	if newFairButton(context) {
		program.home.dayShown = calendar.Date{}
		program.Open(Fairs)
		program.fairs.StartFair(day)
	}
}
