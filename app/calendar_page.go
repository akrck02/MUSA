package app

import (
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
	look.Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(theme.Padding * 1.25), Gap: 12}, func() {
		program.monthHeader(context, shown)
		grid := shown.MonthGrid()
		context.Grid(len(today.Week()), func() {
			for _, day := range grid[:len(today.Week())] {
				context.Text(calendar.ShortWeekday(day), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
			}
			for _, day := range grid {
				dayCell(context, day, today, "", program.today.plan.on(day), monthEntries, day.Month != shown.Month, monthDayHeight)
			}
		})
	})
}

// monthHeader is the month's name between the arrows to the months around it, and «Nueva feria».
func (program *App) monthHeader(context *ui.Context, shown calendar.Date) {
	theme := context.Theme
	// Months are counted from their first day: the 31st of one month has no twin in the next.
	shown = calendar.Date{Year: shown.Year, Month: shown.Month, Day: 1}
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.Center}, func() {
		if context.IconButton("chevron-left", "##Mes anterior") {
			program.home.month = shown.AddMonths(-1)
		}
		context.Text(calendar.MonthWords(shown), ui.TextStyle{Size: theme.HeadingSize, Weight: draw2d.Bold})
		if context.IconButton("chevron-right", "##Mes siguiente") {
			program.home.month = shown.AddMonths(1)
		}
		context.Spacer()
		if context.IconButton("tent", "Todas las ferias") {
			program.Open(Fairs)
		}
		if context.PrimaryIconButton("plus", "Nueva feria") {
			program.Open(Fairs)
			program.fairs.StartFair(program.today.date)
		}
	})
}
