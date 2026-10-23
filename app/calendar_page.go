package app

import (
	"strconv"
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/look"
)

// Heights of the calendar's days, in pixels, and how many entries each shows before "+N".
const (
	weekDayHeight  = 150
	monthDayHeight = 112
	weekEntries    = 3
	monthEntries   = 2
	monthLine      = 0.5
	todaySide      = 24
	legendDot      = 10
)

// drawWeek is this week, Monday to Sunday, with what happens each day.
func (program *App) drawWeek(context *ui.Context) {
	theme := context.Theme
	today := program.today.date
	look.Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(theme.Padding * 1.25), Gap: 12}, func() {
		context.Box(ui.Layout{Width: ui.Grow(), AlignY: ui.Center}, func() {
			context.Box(ui.Layout{Width: ui.Grow()}, func() {
				context.Text("Esta semana", ui.TextStyle{Size: theme.HeadingSize, Weight: draw2d.Bold})
			})
			if link(context, "Ver el mes entero") {
				program.home.tab = 1
			}
		})
		context.Grid(len(today.Week()), func() {
			for _, day := range today.Week() {
				name := calendar.ShortWeekday(day)
				if day == today {
					name = "hoy"
				}
				dayCell(context, day, today, name, program.today.plan.on(day), weekEntries, false, weekDayHeight)
			}
		})
	})
}

// drawMonthCalendar is a whole month, as a wall calendar, with buttons to the months around it.
func (program *App) drawMonthCalendar(context *ui.Context) {
	theme := context.Theme
	today := program.today.date
	shown := program.home.month
	if shown.Zero() {
		shown = today
	}
	program.monthHeader(context, shown)
	grid := shown.MonthGrid()
	look.Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(1)}, func() {
		context.Grid(len(today.Week()), func() {
			for _, day := range grid[:len(today.Week())] {
				context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(10, 8)}, func() {
					context.Text(calendar.ShortWeekday(day), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
				})
			}
		})
		context.Separator()
		context.Box(ui.Layout{Width: ui.Grow(), Columns: len(today.Week())}, func() {
			for _, day := range grid {
				monthCell(context, day, today, program.today.plan.on(day), day.Month != shown.Month)
			}
		})
	})
}

// monthCell is one day of the month's table: a fine line around it, its number (in a dark circle
// today) and its entries as plain bars.
func monthCell(context *ui.Context, day, today calendar.Date, entries []agendaEntry, quiet bool) {
	theme := context.Theme
	fill := theme.Surface
	ink := theme.Ink
	if quiet {
		fill, ink = theme.Sunken.WithAlpha(0.5), theme.InkQuiet
	}
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Fixed(monthDayHeight), Padding: ui.Pad(8), Gap: 4}
	context.Frame(layout, ui.Look{Fill: fill, Border: theme.Line, BorderWidth: monthLine, Clip: true}, func() {
		number := strconv.Itoa(day.Day)
		if day == today {
			side := ui.Fixed(todaySide)
			context.Frame(ui.Layout{Width: side, Height: side, AlignX: ui.Center, AlignY: ui.Center}, ui.Look{Fill: theme.Primary, Radius: todaySide / 2}, func() {
				context.Text(number, ui.TextStyle{Size: theme.SmallTextSize, Weight: draw2d.Bold, Color: theme.OnPrimary})
			})
		} else {
			context.Text(number, ui.TextStyle{Size: theme.SmallTextSize, Weight: draw2d.Bold, Color: ink})
		}
		for index, entry := range entries {
			if index == monthEntries {
				context.Text("+"+strconv.Itoa(len(entries)-monthEntries), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
				break
			}
			entryBar(context, entry)
		}
	})
}

// monthHeader is the month's name between the arrows to the months around it, and «Nueva feria».
func (program *App) monthHeader(context *ui.Context, shown calendar.Date) {
	theme := context.Theme
	// Months are counted from their first day: the 31st of one month has no twin in the next.
	shown = calendar.Date{Year: shown.Year, Month: shown.Month, Day: 1}
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 12, AlignY: ui.Center}, func() {
		if context.IconButton("chevron-left", "##Mes anterior") {
			program.home.month = shown.AddMonths(-1)
		}
		context.Text(capitalised(calendar.MonthWords(shown)), ui.TextStyle{Size: theme.HeadingSize, Weight: draw2d.Bold})
		if context.IconButton("chevron-right", "##Mes siguiente") {
			program.home.month = shown.AddMonths(1)
		}
		if context.PrimaryIconButton("plus", "Nueva feria") {
			program.Open(Fairs)
			program.fairs.StartFair(program.today.date)
		}
		context.Spacer()
		legend(context)
		if look.Link(context, "Todas las ferias") {
			program.Open(Fairs)
		}
	})
}

// legend says what each colour of the calendar stands for.
func legend(context *ui.Context) {
	theme := context.Theme
	for _, item := range []agendaEntry{{words: "Ferias", fair: true}, {words: "Proyectos", project: true}, {words: "Envíos"}} {
		context.Box(ui.Layout{Gap: 6, AlignY: ui.Center}, func() {
			side := ui.Fixed(legendDot)
			context.Frame(ui.Layout{Width: side, Height: side}, ui.Look{Fill: legendColor(theme, item), Radius: legendDot / 2}, nil)
			context.Text(item.words, ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
		})
	}
}

// capitalised is words with a capital first letter: "Noviembre de 2026".
func capitalised(words string) string {
	for index, letter := range words {
		return strings.ToUpper(string(letter)) + words[index+len(string(letter)):]
	}
	return words
}

// legendColor is the deep colour of an entry's kind, for the legend's dots.
func legendColor(theme ui.Theme, entry agendaEntry) draw2d.Color {
	switch {
	case entry.fair:
		return theme.Primary
	case entry.project:
		return theme.Warning
	}
	return theme.Positive
}
