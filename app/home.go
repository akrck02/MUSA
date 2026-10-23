package app

import (
	"errors"
	"strings"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

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

// section is a part of MUSA as the home page offers it.
type section struct {
	place       Place
	icon, title string
	description string
	ready       bool
}

var sections = []section{
	{Catalogue, "palette", "Catálogo", "Tus productos digitales y físicos, con variantes, precios y stock.", true},
	{Fairs, "calendar-days", "Eventos", "Ferias y mercados: qué llevas, cuánto vendes y cuánto te cuesta ir.", true},
	{Sales, "receipt", "Ventas", "Cada venta registrada, con su canal, su margen y su envío.", true},
	{Home, "trending-up", "Análisis de mercado", "Qué se vende mejor, dónde y a qué precio. Pronto.", false},
}

func (program *App) drawHome(context *ui.Context) {
	theme := context.Theme
	look.PageHeader(context, "MUSA", "Tu estudio, tu stock y tus ventas, en un sitio", nil)
	context.Box(ui.Layout{Width: ui.Grow(), Columns: look.Columns(context, 2), Gap: theme.Gap}, func() {
		for _, each := range sections {
			program.sectionCard(context, each)
		}
	})
	if program.options.DevelopmentTools {
		program.developmentTools(context)
	}
	studioFooter(context)
	if time.Now().Before(program.home.noteUntil) {
		context.Toast(program.home.note, program.home.noteWarning)
		context.RequestFrame(time.Until(program.home.noteUntil))
	}
}

func (program *App) sectionCard(context *ui.Context, each section) {
	theme := context.Theme
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(theme.Padding), Gap: 6}
	content := func() {
		colour := theme.Primary
		if !each.ready {
			colour = theme.InkQuiet
		}
		context.IconSized(each.icon, theme.IconSize*1.5, colour)
		context.Text(each.title, ui.TextStyle{Size: theme.HeadingSize, Weight: draw2d.Bold})
		look.Paragraph(context, each.description, ui.TextStyle{Color: theme.InkSoft})
	}
	if !each.ready {
		look.Card(context, layout, content)
		return
	}
	if look.ClickableCard(context, each.title, layout, content) {
		program.Open(each.place)
	}
}

func (program *App) developmentTools(context *ui.Context) {
	look.Section(context, "Herramientas de prueba", func() {
		look.Paragraph(context, "Solo en builds de desarrollo. Rellena la base de datos con datos de juguete o vacíala para empezar de cero.",
			ui.TextStyle{Color: context.Theme.InkSoft})
		context.Box(ui.Layout{Gap: 8, Wrap: true}, func() {
			if context.PrimaryIconButton("sparkles", "Rellenar con datos de prueba") {
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
