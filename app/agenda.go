package app

import (
	"strconv"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/events"
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

// chipHeight is the height of an entry in a day of the calendar, in pixels.
const chipHeight = 26

// entryChip is one entry inside a day: its icon and words on a soft patch, a fair in the accent's
// tint, a project in the warning colour's and an order in the positive colour's.
func entryChip(context *ui.Context, entry agendaEntry) {
	theme := context.Theme
	fill := entryFill(theme, entry)
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(chipHeight), Padding: ui.PadXY(8, 0), Gap: 6, AlignY: ui.Center}
	context.Frame(layout, ui.Look{Fill: fill, Radius: chipHeight / 2, Clip: true}, func() {
		context.IconSized(entry.icon, theme.SmallTextSize, theme.Ink)
		context.Text(entry.words, ui.TextStyle{Size: theme.SmallTextSize, Weight: draw2d.Medium})
	})
}

// entryFill is the colour of an entry: a fair in the accent's tint, a project in the warning
// colour's and an order in the positive colour's.
func entryFill(theme ui.Theme, entry agendaEntry) draw2d.Color {
	switch {
	case entry.fair:
		return theme.Surface.Mix(theme.Primary, 0.14)
	case entry.project:
		return theme.Warning.WithAlpha(0.18)
	}
	return theme.Positive.WithAlpha(0.16)
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

// dayCell is one day of the calendar: its name when given, its number, and its entries, at most
// shown of them and "+N" for the rest. Today stands on the accent's tint; days of another month
// are quieter.
func dayCell(context *ui.Context, day, today calendar.Date, name string, entries []agendaEntry, shown int, quiet bool, height float32) {
	theme := context.Theme
	fill := theme.Sunken.WithAlpha(0.5)
	if day == today {
		fill = theme.Surface.Mix(theme.Primary, 0.12)
	}
	ink := theme.Ink
	if quiet {
		ink = theme.InkQuiet
	}
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Fixed(height), Padding: ui.Pad(10), Gap: 6}
	context.Frame(layout, ui.Look{Fill: fill, Radius: theme.ControlRadius, Clip: true}, func() {
		if name != "" {
			context.Text(name, ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
		}
		context.Text(strconv.Itoa(day.Day), ui.TextStyle{Size: theme.TextSize * 1.2, Weight: draw2d.Bold, Color: ink})
		for index, entry := range entries {
			if index == shown {
				context.Text("+"+strconv.Itoa(len(entries)-shown), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
				break
			}
			entryChip(context, entry)
		}
	})
}
