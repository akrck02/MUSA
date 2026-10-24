package app

import (
	"strconv"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/look"
)

// Measures of the week on the home page, in pixels, as in the author's design.
const (
	weekTop         = 18
	weekTitleSize   = 22
	weekTitleHeight = 1.2
	weekLinkGap     = 8
	weekDaysTop     = 16
	weekDaysGap     = 10
	weekDayHeight   = 150
	weekDayRadius   = 14
	weekDayPadding  = 10
	weekDayNameSize = 13
	weekDayNumber   = 22
)

// drawWeek is this week, Monday to Sunday, with what happens each day.
func (program *App) drawWeek(context *ui.Context) {
	theme := context.Theme
	today := program.today.date
	look.Space(context, weekTop)
	card := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: widgetRadius}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(widgetPadding)}, card, func() {
		context.Box(ui.Layout{Width: ui.Grow(), AlignY: ui.Center}, func() {
			context.Box(ui.Layout{Width: ui.Grow()}, func() {
				context.Text("Esta semana", ui.TextStyle{Size: weekTitleSize, Weight: draw2d.Bold, LineHeight: weekTitleHeight})
			})
			if link(context, "Ver el mes entero") {
				program.home.tab = 1
			}
		})
		look.Space(context, weekDaysTop)
		context.Box(ui.Layout{Width: ui.Grow(), Gap: weekDaysGap, Columns: len(today.Week())}, func() {
			for _, day := range today.Week() {
				name := calendar.ShortWeekday(day)
				if day == today {
					name = "hoy"
				}
				weekDay(context, day, today, name, program.today.plan.on(day))
			}
		})
	})
}

// weekDay is one day of the week: its name, its number, and its entries; today stands on the
// sunken colour with its number in the accent.
func weekDay(context *ui.Context, day, today calendar.Date, name string, entries []agendaEntry) {
	theme := context.Theme
	fill, number := theme.Background, theme.Ink
	if day == today {
		fill, number = theme.Sunken, theme.Primary
	}
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow().Between(weekDayHeight, 0), Padding: ui.Pad(weekDayPadding)}
	context.Frame(layout, ui.Look{Fill: fill, Radius: weekDayRadius, Clip: true}, func() {
		context.Text(name, ui.TextStyle{Size: weekDayNameSize, Color: theme.InkSoft})
		context.Text(strconv.Itoa(day.Day), ui.TextStyle{Size: weekDayNumber, Weight: draw2d.Bold, Color: number})
		for _, entry := range entries {
			entryChip(context, entry)
		}
	})
}

// Measures of the month's calendar, in pixels, as in the author's design.
const (
	monthHeaderTop    = 18
	monthHeaderBottom = 12
	monthHeaderGap    = 12
	monthTitleSize    = 24
	monthTitleWidth   = 200
	roundSide         = 44
	roundIconSize     = 20
	newFairSpace      = 16
	newFairPaddingX   = 18
	newFairPaddingY   = 10
	newFairSize       = 16
	legendGap         = 16
	legendSize        = 15
	monthRadius       = 18
	weekdayPaddingX   = 12
	weekdayPaddingY   = 10
	weekdaySize       = 14
	monthDayHeight    = 96
	monthDayPadding   = 8
	monthDaySize      = 14
	todayPaddingX     = 8
	todayPaddingY     = 2
	monthEntries      = 2
	legendDot         = 10
)

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
	days := len(today.Week())
	card := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: monthRadius, Clip: true}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(1)}, card, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Columns: days}, func() {
			for _, day := range grid[:days] {
				context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(weekdayPaddingX, weekdayPaddingY)}, func() {
					context.Text(calendar.ShortWeekday(day), ui.TextStyle{Size: weekdaySize, Color: theme.InkSoft})
				})
			}
		})
		context.Separator()
		context.Box(ui.Layout{Width: ui.Grow(), Columns: days}, func() {
			for index, day := range grid {
				cell := monthCellOf(day, today, shown, index%days == 0)
				cell.corner = program.cornerOf(theme, cell, index, len(grid), days)
				monthCell(context, cell, program.today.plan.on(day))
			}
		})
	})
}

// cellOfMonth is how one day of the month's table is drawn.
type cellOfMonth struct {
	day          calendar.Date
	today, quiet bool // quiet: a day of another month
	first        bool // the first of its row, with no line on its left
	// corner, when not nil, paints a cell in a bottom corner of the table, rounded as the table.
	corner *look.Corners
}

func monthCellOf(day, today, shown calendar.Date, first bool) cellOfMonth {
	return cellOfMonth{day: day, today: day == today, quiet: day.Month != shown.Month, first: first}
}

// cornerOf is the rounded fill of a cell in a bottom corner of the table, or nil for the others.
func (program *App) cornerOf(theme ui.Theme, cell cellOfMonth, index, cells, days int) *look.Corners {
	lastRow := index >= cells-days
	if !lastRow || (!cell.first && index != cells-1) {
		return nil
	}
	corner, rounded := &program.home.corners[0], look.BottomLeft
	if !cell.first {
		corner, rounded = &program.home.corners[1], look.BottomRight
	}
	*corner = look.Corners{Fill: cellFill(theme, cell), Radius: monthRadius - 1, Rounded: rounded}
	return corner
}

// cellFill is a day's background: the page's colour for a day of another month.
func cellFill(theme ui.Theme, cell cellOfMonth) draw2d.Color {
	if cell.quiet {
		return theme.Background
	}
	return theme.Surface
}

// monthCell is one day of the month's table: a fine line above it and on its left (the first of a
// row has none there), its number (on an accent pill today) and its entries as bars.
func monthCell(context *ui.Context, cell cellOfMonth, entries []agendaEntry) {
	theme := context.Theme
	ink := theme.Ink
	if cell.quiet {
		ink = theme.InkQuiet
	}
	background := ui.Look{Fill: cellFill(theme, cell), Clip: true}
	if cell.corner != nil {
		background = ui.Look{Backdrop: cell.corner, Clip: true}
	}
	context.Frame(ui.Layout{Width: ui.Grow(), Height: ui.Fixed(monthDayHeight)}, background, func() {
		if !cell.first {
			context.Frame(ui.Layout{Width: ui.Fixed(1), Height: ui.Grow()}, ui.Look{Fill: theme.Line}, nil)
		}
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow()}, func() {
			context.Separator()
			context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(monthDayPadding)}, func() {
				dayNumber(context, cell.day.Day, cell.today, ink)
				for index, entry := range entries {
					if index == monthEntries {
						context.Text("+"+strconv.Itoa(len(entries)-monthEntries), ui.TextStyle{Size: chipTextSize, Color: theme.InkSoft})
						break
					}
					entryBar(context, entry)
				}
			})
		})
	})
}

// dayNumber is a day's number, on an accent pill when it is today.
func dayNumber(context *ui.Context, number int, today bool, ink draw2d.Color) {
	theme := context.Theme
	style := ui.TextStyle{Size: monthDaySize, Weight: draw2d.Bold, Color: ink}
	if !today {
		context.Text(strconv.Itoa(number), style)
		return
	}
	style.Color = theme.OnPrimary
	context.Box(ui.Layout{}, func() {
		context.Frame(ui.Layout{Padding: ui.PadXY(todayPaddingX, todayPaddingY)}, ui.Look{Fill: theme.Primary, Radius: monthDaySize}, func() {
			context.Text(strconv.Itoa(number), style)
		})
	})
}

// monthHeader is the month's name between the arrows to the months around it, «Nueva feria», the
// legend and a link to every fair.
func (program *App) monthHeader(context *ui.Context, shown calendar.Date) {
	// Months are counted from their first day: the 31st of one month has no twin in the next.
	shown = calendar.Date{Year: shown.Year, Month: shown.Month, Day: 1}
	look.Space(context, monthHeaderTop)
	context.Box(ui.Layout{Width: ui.Grow(), Gap: monthHeaderGap, AlignY: ui.Center, Padding: ui.Padding{Bottom: monthHeaderBottom}}, func() {
		if roundButton(context, "chevron-left", "Mes anterior") {
			program.home.month = shown.AddMonths(-1)
		}
		context.Box(ui.Layout{Width: ui.Fit().Between(monthTitleWidth, 0)}, func() {
			context.Text(calendar.MonthTitle(shown), ui.TextStyle{Size: monthTitleSize, Weight: draw2d.Bold, LineHeight: weekTitleHeight})
		})
		if roundButton(context, "chevron-right", "Mes siguiente") {
			program.home.month = shown.AddMonths(1)
		}
		look.Space(context, newFairSpace-monthHeaderGap)
		if newFairButton(context) {
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

// roundButton is a round button with only an icon, on the surface with a fine line.
func roundButton(context *ui.Context, icon, label string) bool {
	theme := context.Theme
	normal := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: roundSide / 2}
	hovered := normal
	hovered.Fill = theme.Sunken
	side := ui.Fixed(roundSide)
	clicked := context.Clickable(label, ui.Layout{Width: side, Height: side, AlignX: ui.Center, AlignY: ui.Center},
		ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
			context.IconSized(icon, roundIconSize, theme.Ink)
		})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: label})
	return clicked
}

// newFairButton is «Nueva feria» on the accent, a little smaller than a page's main button.
func newFairButton(context *ui.Context) bool {
	theme := context.Theme
	normal := ui.Look{Fill: theme.Primary, Radius: goRadius}
	hovered := normal
	hovered.Fill = theme.Primary.Mix(theme.Ink, hoverDarkening)
	layout := ui.Layout{Padding: ui.PadXY(newFairPaddingX, newFairPaddingY), Gap: mainGap, AlignY: ui.Center}
	clicked := context.Clickable("Nueva feria", layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
		context.IconSized("plus", goIconSize, theme.OnPrimary)
		context.Text("Nueva feria", ui.TextStyle{Size: newFairSize, Weight: draw2d.Bold, Color: theme.OnPrimary})
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: "Nueva feria"})
	return clicked
}

// mainGap is the space between a main button's icon and its words.
const mainGap = 10

// legend says what each colour of the calendar stands for.
func legend(context *ui.Context) {
	theme := context.Theme
	context.Box(ui.Layout{Gap: legendGap, AlignY: ui.Center}, func() {
		for _, item := range []agendaEntry{{words: "Ferias", fair: true}, {words: "Proyectos", project: true}, {words: "Envíos"}} {
			context.Box(ui.Layout{Gap: chipGap, AlignY: ui.Center}, func() {
				side := ui.Fixed(legendDot)
				context.Frame(ui.Layout{Width: side, Height: side}, ui.Look{Fill: legendColor(theme, item), Radius: legendDot / 2}, nil)
				context.Text(item.words, ui.TextStyle{Size: legendSize, Color: theme.InkSoft})
			})
		}
	})
}

// The deep colours of the legend's dots, as in the author's design.
var (
	projectDot = draw2d.MustHex("#E8C9B0")
	orderDot   = draw2d.MustHex("#B9CFC4")
)

// legendColor is the colour of an entry's kind, for the legend's dots.
func legendColor(theme ui.Theme, entry agendaEntry) draw2d.Color {
	switch {
	case entry.fair:
		return theme.Primary
	case entry.project:
		return projectDot
	}
	return orderDot
}
