package app

import (
	"errors"
	"strings"
	"time"

	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/demo"
	"github.com/Itros97/MUSA/look"
)

// toastTime is how long a note about what a button did stays on screen.
const toastTime = 3 * time.Second

// homeState is what the home page remembers: the tab shown, the month the calendar shows, and the
// development tools' note and question.
type homeState struct {
	tab         int           // index in homeTabs
	month       calendar.Date // a day of the month the calendar shows; zero for this month
	note        string
	noteWarning bool
	noteUntil   time.Time
	askingWipe  bool
	// corners are the fills of the month's bottom corner cells, kept so drawing them allocates
	// nothing.
	corners [2]look.Corners
}

// homeTabs are the two views of the home page.
var homeTabs = []look.Tab{{Label: "Resumen"}, {Icon: "calendar", Label: "Calendario"}}

func (program *App) drawHome(context *ui.Context) {
	look.Space(context, greetingTop)
	look.SizedPageHeader(context, "Hola", calendar.LongWords(program.today.date), look.Greeting, func() {
		look.Segmented(context, "home", &program.home.tab, homeTabs, look.Views)
	})
	if program.today.loadErr != nil {
		context.Toast("No se pudo leer lo de hoy", true)
	}
	if program.home.tab == 1 {
		program.drawMonthCalendar(context)
	} else {
		program.drawSummary(context)
	}
	if time.Now().Before(program.home.noteUntil) {
		context.Toast(program.home.note, program.home.noteWarning)
		context.RequestFrame(time.Until(program.home.noteUntil))
	}
}

// drawSummary is the home page at a glance: the fair, the month, what runs low, and this week.
func (program *App) drawSummary(context *ui.Context) {
	if look.Narrow(context) {
		context.Grid(1, func() {
			program.fairWidget(context)
			program.monthWidget(context)
			program.restockWidget(context)
		})
	} else {
		look.Space(context, widgetsTop)
		context.Box(ui.Layout{Width: ui.Grow(), Gap: widgetsGap}, func() {
			context.Box(ui.Layout{Width: ui.Share(fairShare), Height: ui.Grow()}, func() { program.fairWidget(context) })
			context.Box(ui.Layout{Width: ui.Share(1), Height: ui.Grow()}, func() { program.monthWidget(context) })
			context.Box(ui.Layout{Width: ui.Share(1), Height: ui.Grow()}, func() { program.restockWidget(context) })
		})
	}
	program.drawWeek(context)
	if program.options.DevelopmentTools {
		program.developmentTools(context)
	}
	program.studioFooter(context)
}

func (program *App) developmentTools(context *ui.Context) {
	look.Section(context, "Para probar (solo en desarrollo)", func() {
		look.Paragraph(context, "Solo en builds de desarrollo. Rellena la base de datos con datos de juguete o vacíala para empezar de cero.",
			ui.TextStyle{Color: context.Theme.InkSoft})
		context.Box(ui.Layout{Direction: look.Across(context), Gap: 8}, func() {
			if context.IconButton("sparkles", "Rellenar con datos de prueba") {
				program.fillWithDemo()
			}
			if context.IconButton("trash-2", "Vaciar base de datos") {
				program.home.askingWipe = true
			}
		})
	})
	if program.home.askingWipe {
		program.askWipe(context)
	}
}

func (program *App) askWipe(context *ui.Context) {
	if context.Sheet("¿Vaciar la base de datos?", ui.SheetOptions{}, func() {
		look.Paragraph(context, "Se borrarán todos los productos, etiquetas, eventos, ventas y gastos. No se puede deshacer.", ui.TextStyle{})
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignX: ui.End}, func() {
			if context.Button("Cancelar") {
				program.home.askingWipe = false
			}
			if context.PrimaryButton("Vaciar") {
				program.home.askingWipe = false
				program.tell(program.stores.Database.Wipe(), "Base de datos vacía")
				program.Open(Home)
			}
		})
	}) {
		program.home.askingWipe = false
	}
}

func (program *App) fillWithDemo() {
	err := demo.Fill(program.stores)
	if err != nil && strings.HasPrefix(err.Error(), demo.CodeNotEmpty) {
		err = errors.New("La base de datos ya tiene datos: vacíala primero")
	}
	program.tell(err, "Datos de prueba cargados")
	program.Open(Home)
}

// tell shows what a development tool did, or why it could not.
func (program *App) tell(err error, done string) {
	program.home.note, program.home.noteWarning = done, false
	if err != nil {
		program.home.note, program.home.noteWarning = err.Error(), true
	}
	program.home.noteUntil = time.Now().Add(toastTime)
}

// studioFooter is the quiet end of the home page: the switch between the light and the dark
// theme, and the signature of the app and who makes it.
func (program *App) studioFooter(context *ui.Context) {
	theme := context.Theme
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.PadXY(0, theme.Padding*1.5), Gap: 6, AlignX: ui.Center}, func() {
		context.Box(ui.Layout{Gap: 8, AlignY: ui.Center}, func() {
			spaced(context, "MUSA", theme.SmallTextSize)
			context.Text("by Software Anvil", ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkQuiet})
		})
		context.Text("© 2026 · Itros97", ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkQuiet})
		program.themeSwitch(context)
	})
}

// themeSwitch is a quiet button that says which theme it changes to: «Modo oscuro» or «Modo claro».
func (program *App) themeSwitch(context *ui.Context) {
	theme := context.Theme
	icon, words := "moon", "Modo oscuro"
	if program.dark {
		icon, words = "sun", "Modo claro"
	}
	hovered := ui.Look{Fill: theme.Sunken, Radius: theme.ControlRadius}
	layout := ui.Layout{Gap: themeSwitchGap, AlignY: ui.Center, Padding: ui.PadXY(themeSwitchPaddingX, themeSwitchPaddingY)}
	if context.Clickable(words, layout, ui.Looks{Hovered: hovered, Pressed: hovered}, func() {
		context.IconSized(icon, theme.SmallTextSize, theme.InkSoft)
		context.Text(words, ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
	}) {
		program.SetDark(!program.dark)
	}
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: words})
}

// Measures of the theme switch, in pixels.
const (
	themeSwitchGap      = 6
	themeSwitchPaddingX = 10
	themeSwitchPaddingY = 6
)
