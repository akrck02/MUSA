package app

import (
	"strconv"

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
	dayDetailWide    = 560 // the width of a detail listing every entry
	dayDetailList    = 420 // the tallest the whole list grows before it scrolls
	dayDetailEntries = 4   // entries shown before «Ver todos»
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
	width := float32(dayDetailWidth)
	if program.home.dayWhole {
		width = dayDetailWide
	}
	options := ui.LayerOptions{AlignX: ui.Center, AlignY: ui.Center, Width: ui.Fixed(context.Animate("day detail width", width, look.DialogMotion)),
		Hidden: 1 - shown, Leave: look.Leaving(look.DialogRise)}
	card := look.Arriving(ui.Look{Fill: theme.Surface, Radius: dayDetailRadius, Shadow: true}, shown, look.DialogRise)
	if context.Modal("day detail", options, func() {
		context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(dayDetailPadding)}, card, func() {
			program.dayDetailContent(context, day)
		})
	}) {
		program.closeDay()
	}
}

// closeDay closes the day's detail, which opens short next time.
func (program *App) closeDay() {
	program.home.dayShown, program.home.dayWhole = calendar.Date{}, false
}

// dayDetailContent is the day's name, its entries, «Ver todos» when they are many, and «Nuevo evento» on it.
func (program *App) dayDetailContent(context *ui.Context, day calendar.Date) {
	theme := context.Theme
	context.Text(calendar.LongWords(day), ui.TextStyle{Size: dayDetailTitle, Weight: draw2d.Bold})
	entries := program.today.plan.on(day)
	if len(entries) == 0 {
		look.Space(context, chipTop)
		context.Text("Nada este día.", ui.TextStyle{Size: dayDetailNothing, Color: theme.InkSoft})
	}
	if program.home.dayWhole {
		list := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Fit().Between(0, dayDetailList)}
		context.ScrollArea("whole day", list, func() { program.dayEntries(context, entries) })
	} else {
		program.dayEntries(context, entries[:min(len(entries), dayDetailEntries)])
	}
	if !program.home.dayWhole && len(entries) > dayDetailEntries {
		look.Space(context, chipTop)
		if look.Link(context, "Ver todos ("+strconv.Itoa(len(entries))+")") {
			program.home.dayWhole = true
		}
	}
	look.Space(context, dayDetailBottom)
	if newFairButton(context) {
		program.closeDay()
		program.Open(Fairs)
		program.fairs.StartFair(day)
	}
}

// dayEntries lists entries of the day, each opening what it is about.
func (program *App) dayEntries(context *ui.Context, entries []agendaEntry) {
	for _, entry := range entries {
		if entryChip(context, entry) {
			program.closeDay()
			program.openEntry(entry)
		}
	}
}
