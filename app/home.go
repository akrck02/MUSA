package app

import (
	"errors"
	"strings"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/demo"
	"github.com/Itros97/MUSA/look"
)

// toastTime is how long a note about what a button did stays on screen.
const toastTime = 3 * time.Second

// homeState is what the home page remembers: the development tools' note and question.
type homeState struct {
	note        string
	noteWarning bool
	noteUntil   time.Time
	askingWipe  bool
}

// section is a part of MUSA as the home page offers it: a big picture button with a plain name.
type section struct {
	place       Place
	icon, title string
	description string
}

var sections = []section{
	{Stand, "store", "Abrir el puesto", "Cobra en una feria o en casa"},
	{Catalogue, "palette", "Mis productos", "Tus láminas, pegatinas y demás"},
	{Fairs, "tent", "Ferias", "Lo que llevas y lo que ganas"},
	{Sales, "package", "Pedidos", "Lo que tienes que enviar, y lo vendido"},
}

func (program *App) drawHome(context *ui.Context) {
	look.PageHeader(context, "Hola", calendar.LongWords(program.today.date), nil)
	program.drawToday(context)
	columns := 1
	if look.Wide(context) {
		columns = len(sections)
	}
	context.Box(ui.Layout{Width: ui.Grow(), Columns: columns, Gap: context.Theme.Gap}, func() {
		for _, each := range sections {
			program.sectionCard(context, each)
		}
	})
	program.drawMonth(context)
	if program.options.DevelopmentTools {
		program.developmentTools(context)
	}
	studioFooter(context)
	if time.Now().Before(program.home.noteUntil) {
		context.Toast(program.home.note, program.home.noteWarning)
		context.RequestFrame(time.Until(program.home.noteUntil))
	}
}

// sectionIcon is how big the picture of a section is, in pixels.
const sectionIcon = 72

// sectionCard is a big button with a picture: three side by side on a wide window, and on a
// narrower one a full-width row each, the picture beside the words, so no word is ever cut.
func (program *App) sectionCard(context *ui.Context, each section) {
	theme := context.Theme
	layout := ui.Layout{Width: ui.Grow(), Padding: ui.Pad(theme.Padding * 1.25), Gap: theme.Gap * 1.5, AlignY: ui.Center}
	if look.Wide(context) {
		layout.Direction, layout.AlignX = ui.Vertical, ui.Center
	}
	if look.ClickableCard(context, each.title, layout, func() {
		badge, title := float32(sectionIcon), theme.HeadingSize*1.2
		if look.Narrow(context) {
			badge, title = sectionIcon*0.75, theme.HeadingSize
		}
		look.Badge(context, each.icon, badge)
		words := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 6, AlignX: layout.AlignX}
		if look.Wide(context) {
			words.Width = ui.Layout{}.Width
		}
		context.Box(words, func() {
			context.Text(each.title, ui.TextStyle{Size: title, Weight: draw2d.Bold})
			look.Paragraph(context, each.description, ui.TextStyle{Color: theme.InkSoft})
		})
	}) {
		program.Open(each.place)
	}
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

// studioFooter is the quiet signature at the end of the home page: the app and who makes it.
func studioFooter(context *ui.Context) {
	theme := context.Theme
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.PadXY(0, theme.Padding*1.5), Gap: 6, AlignX: ui.Center}, func() {
		context.Box(ui.Layout{Gap: 8, AlignY: ui.Center}, func() {
			spaced(context, "MUSA", theme.SmallTextSize)
			context.Text("by Software Anvil", ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkQuiet})
		})
		context.Text("© 2026 · Itros97", ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkQuiet})
	})
}
