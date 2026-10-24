package app

import (
	"fmt"
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
	"github.com/Itros97/MUSA/stand"
)

// Measures of the home page's widgets, in pixels, as in the author's design. A border takes no
// room in go-ui, so paddings beside one are a pixel larger than the design's CSS.
const (
	greetingTop     = 4 // the home page starts a little lower than the others
	widgetsTop      = 22
	widgetsGap      = 18
	fairShare       = 1.4 // the fair is the widget the artist acts on, so it is wider
	widgetRadius    = 22
	widgetPadding   = 22 + 1
	linksTop        = 14
	fairLinksTop    = 16
	linkSize        = 15
	linkIconSize    = 16
	linkGap         = 6
	fairCircle      = 72
	fairIconSize    = 34
	fairGap         = 18
	fairLabelSize   = 13
	fairLabelSpace  = 0.05 // the label's letter spacing, in ems
	fairNameSize    = 24
	fairNameSpace   = 4
	bigFigure       = 40
	bigFigureSpace  = 6
	restockRows     = 3
	restockTitleGap = 6
	restockRowY     = 9
	restockGap      = 12
	swatchSide      = 32
	swatchRadius    = 9
	goRadius        = 14
	goPaddingX      = 20
	goPaddingY      = 14
	goSize          = 17
	goIconSize      = 18
	goGap           = 8
)

// widget is one card of the home page: its content on top and its links at the bottom, as tall as
// the tallest widget of its row.
func widget(context *ui.Context, linksSpace float32, content, links func()) {
	theme := context.Theme
	card := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: widgetRadius}
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(widgetPadding)}
	context.Frame(layout, card, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow()}, content)
		context.Box(ui.Layout{Height: ui.Grow()}, nil)
		look.Space(context, linksSpace)
		// A link stays on the bottom line, as CSS's margin-top: auto leaves it beside a taller button.
		context.Box(ui.Layout{Width: ui.Grow(), AlignY: ui.End}, links)
	})
}

// link is a quiet "see more" at the bottom of a widget: bold words and an arrow after them.
func link(context *ui.Context, words string) bool {
	theme := context.Theme
	clicked := context.Clickable(words, ui.Layout{Gap: linkGap, AlignY: ui.Center}, ui.Looks{}, func() {
		context.Text(words, ui.TextStyle{Size: linkSize, Weight: draw2d.Bold})
		context.IconSized("arrow-right", linkIconSize, theme.Ink)
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: words})
	return clicked
}

// goButton is a widget's main action, on the accent; icon may be empty.
func goButton(context *ui.Context, icon, words string) bool {
	theme := context.Theme
	normal := ui.Look{Fill: theme.Primary, Radius: goRadius}
	hovered := normal
	hovered.Fill = theme.Primary.Mix(theme.Ink, hoverDarkening)
	layout := ui.Layout{Padding: ui.PadXY(goPaddingX, goPaddingY), Gap: goGap, AlignY: ui.Center}
	clicked := context.Clickable(words, layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
		if icon != "" {
			context.IconSized(icon, goIconSize, theme.OnPrimary)
		}
		context.Text(words, ui.TextStyle{Size: goSize, Weight: draw2d.Bold, Color: theme.OnPrimary, LineHeight: ui.LettersLineHeight})
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: words})
	return clicked
}

// hoverDarkening is how much ink a button on the accent takes under the pointer.
const hoverDarkening = 0.15

// fairWidget is the fair to look at: today's, with a button to the stand, or the next one.
func (program *App) fairWidget(context *ui.Context) {
	day := program.today
	if !day.hasFair {
		widget(context, fairLinksTop, func() {
			fairHeading(context, "", "Sin ferias a la vista", "")
		}, func() {
			if link(context, "Nueva feria") {
				program.Open(Fairs)
				program.fairs.StartFair(day.date)
			}
		})
		return
	}
	happening := events.TimingOf(day.fair, day.date) == events.Happening
	label := "PRÓXIMA FERIA"
	if happening {
		label = "HOY ESTÁS EN"
	}
	widget(context, fairLinksTop, func() {
		fairHeading(context, label, day.fair.Name, events.WhenWords(day.fair, day.date)+" · "+day.fair.Location)
	}, func() {
		program.fairLinks(context, happening)
	})
}

// fairHeading is the fair widget's top: a tent on a circle, a spaced label, the fair's name and
// when and where it is; label and words may be empty.
func fairHeading(context *ui.Context, label, name, words string) {
	theme := context.Theme
	context.Box(ui.Layout{Width: ui.Grow(), Gap: fairGap, AlignY: ui.Center}, func() {
		side := ui.Fixed(fairCircle)
		context.Frame(ui.Layout{Width: side, Height: side, AlignX: ui.Center, AlignY: ui.Center},
			ui.Look{Fill: theme.Sunken, Radius: fairCircle / 2}, func() {
				context.IconSized("tent", fairIconSize, theme.Primary)
			})
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow()}, func() {
			if label != "" {
				spacedBy(context, label, ui.TextStyle{Size: fairLabelSize, Weight: draw2d.Bold, Color: theme.InkSoft}, fairLabelSpace)
			}
			look.Space(context, fairNameSpace)
			look.Paragraph(context, name, ui.TextStyle{Size: fairNameSize, Weight: draw2d.Bold})
			look.Space(context, fairNameSpace)
			if words != "" {
				look.Paragraph(context, words, ui.TextStyle{Color: theme.InkSoft})
			}
		})
	})
}

func (program *App) fairLinks(context *ui.Context, happening bool) {
	fair := program.today.fair
	if link(context, "Ver la feria") {
		program.Open(Fairs)
		program.fairs.OpenFair(fair.ID)
	}
	context.Spacer()
	if happening && goButton(context, "store", "Ir al puesto") {
		program.Open(Stand)
		program.stand.Choose(stand.Fair{Event: fair, Events: program.stores.Events, Products: program.stores.Products})
	}
	if !happening && goButton(context, "", "Preparar el puesto") {
		program.Open(Fairs)
		program.fairs.OpenFair(fair.ID)
	}
}

// monthWidget is this month's sales in one big figure.
func (program *App) monthWidget(context *ui.Context) {
	theme := context.Theme
	month := program.today
	widget(context, linksTop, func() {
		look.Paragraph(context, "Este mes llevas", ui.TextStyle{Color: theme.InkSoft})
		look.Space(context, bigFigureSpace)
		context.Text(money.Format(month.month.Revenue), ui.TextStyle{Size: bigFigure, Weight: draw2d.Bold})
		look.Space(context, bigFigureSpace)
		look.Paragraph(context, monthWords(month.month.Count-month.projectsDelivered, month.projectsDelivered), ui.TextStyle{Color: theme.InkSoft})
	}, func() {
		if link(context, "Ver ventas") {
			program.Open(Sales)
			program.sales.ShowHistory()
		}
	})
}

// monthWords counts the month's sales and the projects delivered in it: "en 23 ventas · 2
// proyectos entregados"; no projects, no second part.
func monthWords(sales, projects int) string {
	words := "en 1 venta"
	if sales != 1 {
		words = fmt.Sprintf("en %d ventas", sales)
	}
	switch {
	case projects == 1:
		words += " · 1 proyecto entregado"
	case projects > 1:
		words += fmt.Sprintf(" · %d proyectos entregados", projects)
	}
	return words
}

// restockWidget is the versions running low, each with its colour and what is left, the sold out
// in red.
func (program *App) restockWidget(context *ui.Context) {
	theme := context.Theme
	rows := restockRowsOf(program.today.restock)
	widget(context, linksTop, func() {
		look.Paragraph(context, "Te quedan pocos", ui.TextStyle{Color: theme.InkSoft})
		look.Space(context, restockTitleGap)
		if len(rows) == 0 {
			look.Paragraph(context, "Tienes de todo: nada se está acabando.", ui.TextStyle{})
		}
		for index, row := range rows {
			if index == restockRows {
				break
			}
			if index > 0 {
				context.Separator()
			}
			restockLine(context, index, row)
		}
	}, func() {
		if len(rows) > 0 && link(context, fmt.Sprintf("Ver los %d", len(rows))) {
			program.Open(Catalogue)
			program.catalogue.ShowRestock()
		}
	})
}

// restockLine is one version running low: its colour, its name and what is left.
func restockLine(context *ui.Context, index int, row restockRow) {
	theme := context.Theme
	context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(0, restockRowY), Gap: restockGap, AlignY: ui.Center}, func() {
		side := ui.Fixed(swatchSide)
		context.Frame(ui.Layout{Width: side, Height: side}, ui.Look{Fill: catalogue.ListFill(theme, index), Radius: swatchRadius}, nil)
		context.Frame(ui.Layout{Width: ui.Grow()}, ui.Look{Clip: true}, func() {
			context.Text(row.words, ui.TextStyle{Weight: draw2d.Medium})
		})
		colour := theme.Warning
		if row.left == 0 {
			colour = theme.Danger
		}
		context.Text(fmt.Sprint(row.left), ui.TextStyle{Weight: draw2d.Bold, Color: colour})
	})
}

// restockRow is one version running low: its product, how it is called and what is left.
type restockRow struct {
	product, words string
	left           int
}

// restockRowsOf lists the versions running low of products: "Gato astronauta A3", "Luna Mate".
func restockRowsOf(products []catalogue.Product) []restockRow {
	var rows []restockRow
	for _, product := range products {
		for _, variant := range product.Variants {
			if variant.LowStock() {
				rows = append(rows, restockRow{product: product.Name, words: strings.TrimSpace(product.Name + " " + variant.Name), left: variant.Stock})
			}
		}
	}
	return rows
}
