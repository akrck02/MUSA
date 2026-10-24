package app

import (
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/projects"
	"github.com/Itros97/MUSA/sales"
)

// agendaEntry is one thing on a day of the calendar: a fair, or an order to send.
type agendaEntry struct {
	icon, words string
	fair        bool
	project     bool
}

// agenda is what the calendar shows: the fairs, the orders still on their way and the projects
// still to deliver.
type agenda struct {
	fairs    []events.Event
	onTheWay []sales.Sale
	projects []projects.Project
	zone     *time.Location
}

// on is what happens on one day: the fairs it falls in, the orders made that day still to send,
// and the projects due that day.
func (plan agenda) on(day calendar.Date) []agendaEntry {
	var entries []agendaEntry
	for _, fair := range plan.fairs {
		if day.Within(fair.StartDate, fair.EndDate) {
			entries = append(entries, agendaEntry{icon: "tent", words: fair.Name, fair: true})
		}
	}
	for _, order := range plan.onTheWay {
		if calendar.Of(order.SoldAt.In(plan.zone)) == day {
			entries = append(entries, agendaEntry{icon: "truck", words: "Enviar a " + whoTo(order)})
		}
	}
	for _, project := range plan.projects {
		if project.Stage != projects.Delivered && project.Deadline == day {
			entries = append(entries, agendaEntry{icon: "pencil", words: project.Title, project: true})
		}
	}
	return entries
}

// whoTo is the customer of an order, or its channel when it has none.
func whoTo(order sales.Sale) string {
	if order.Customer != "" {
		return order.Customer
	}
	return sales.ChannelWords(order.Channel)
}

// Measures of an entry in a day of the week, in pixels, as in the author's design.
const (
	chipTop        = 8
	chipRadius     = 9
	chipPaddingX   = 8
	chipPaddingY   = 6
	chipGap        = 6
	chipTextSize   = 13
	chipIconSize   = 14
	chipLineHeight = 1.2
)

// The soft colours of the entries, as in the author's design: a fair in the accent's tint, a
// project in a warm one and an order in a green one.
var (
	projectFill = draw2d.MustHex("#E0B39244")
	orderFill   = draw2d.MustHex("#8FB8A444")
)

// fairAlpha is how much of the accent a fair's entry takes.
const fairAlpha = 0x26 / 255.0

// entryChip is one entry inside a day: its icon and words on a soft patch of its kind's colour.
func entryChip(context *ui.Context, entry agendaEntry) {
	theme := context.Theme
	look.Space(context, chipTop)
	layout := ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(chipPaddingX, chipPaddingY), Gap: chipGap, AlignY: ui.Center}
	context.Frame(layout, ui.Look{Fill: entryFill(theme, entry), Radius: chipRadius, Clip: true}, func() {
		context.IconSized(entry.icon, chipIconSize, theme.Ink)
		context.Text(entry.words, ui.TextStyle{Size: chipTextSize, LineHeight: chipLineHeight})
	})
}

// entryFill is the colour of an entry: a fair in the accent's tint, a project in a warm one and an
// order in a green one.
func entryFill(theme ui.Theme, entry agendaEntry) draw2d.Color {
	switch {
	case entry.fair:
		return theme.Primary.WithAlpha(fairAlpha)
	case entry.project:
		return projectFill
	}
	return orderFill
}

// entryBar is an entry in the month's table: its words on a small plain bar.
func entryBar(context *ui.Context, entry agendaEntry) {
	theme := context.Theme
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(barChipHeight), Padding: ui.PadXY(6, 0), AlignY: ui.Center}
	context.Frame(layout, ui.Look{Fill: entryFill(theme, entry), Radius: theme.ControlRadius / 2, Clip: true}, func() {
		context.Text(entry.words, ui.TextStyle{Size: theme.SmallTextSize})
	})
}

// barChipHeight is the height of an entry in the month's table, in pixels.
const barChipHeight = 22
