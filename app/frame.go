package app

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/look"
)

// Measures of the frame, in pixels, as in Bubble's showcase.
const (
	barHeight      = 87  // under a line of 1
	barWidth       = 900 // the places' width on a wide window
	pillWidth      = 64  // the soft pill behind the chosen place's icon
	pillHeight     = 32
	barIconSize    = 22
	barLabelSize   = 13
	brandLetter    = 0.14 // the brand's letter spacing, in ems
	pageTop        = 24
	registerTop    = 20 // the stand while selling: its top bar sits higher
	pageSide       = 32
	pageBottom     = 32
	narrowPageSide = 16
)

// destination is a place as the bar shows it.
type destination struct {
	place      Place
	icon, name string
}

var destinations = []destination{
	{Home, "house", "Inicio"},
	{Stand, "store", "Puesto"},
	{Catalogue, "palette", "Mis productos"},
	{Projects, "square-kanban", "Proyectos"},
	{Sales, "package", "Pedidos"},
}

// frame is the page above and the places in a bar along the bottom, on every size, as on an
// Android phone: one way to move around, always in the same place, within reach of a thumb.
func (program *App) frame(context *ui.Context) {
	theme := context.Theme
	insets := context.Insets()
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Padding{Top: insets.Top}}, func() {
		program.scrollingPage(context)
		if program.place == Stand {
			program.stand.DrawBasketBar(context, barHeight+1+insets.Bottom) // the bar and the line over it
		}
		context.Separator()
		bar := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(barHeight + insets.Bottom), Padding: ui.Padding{Bottom: insets.Bottom}, AlignX: ui.Center}
		context.Frame(bar, ui.Look{Fill: theme.Surface}, func() {
			// On a wide window the places stay together in the middle, not spread to the edges.
			width := ui.Grow()
			if !look.Narrow(context) {
				width = ui.Fixed(barWidth)
			}
			context.Box(ui.Layout{Width: width, Height: ui.Grow(), AlignX: ui.SpaceAround, AlignY: ui.Center}, func() {
				for _, each := range destinations {
					program.barItem(context, each)
				}
			})
		})
	})
}

// barItem is one place in the bar: its icon over its name; the chosen one has its icon on a soft
// pill and its name in bold, as Android's navigation bar marks it.
func (program *App) barItem(context *ui.Context, each destination) {
	theme := context.Theme
	chosen := each.place == program.place
	colour, weight, pill := theme.InkSoft, draw2d.Regular, draw2d.Transparent
	if chosen {
		colour, weight, pill = theme.Ink, draw2d.Bold, theme.Sunken
	}
	layout := ui.Layout{Direction: ui.Vertical, Gap: 4, AlignX: ui.Center, AlignY: ui.Center}
	looks := ui.Looks{}
	if context.Clickable(each.name+"##bar", layout, looks, func() {
		context.Frame(ui.Layout{Width: ui.Fixed(pillWidth), Height: ui.Fixed(pillHeight), AlignX: ui.Center, AlignY: ui.Center},
			ui.Look{Fill: pill, Radius: pillHeight / 2}, func() {
				context.IconSized(each.icon, barIconSize, colour)
			})
		context.Text(each.name, ui.TextStyle{Size: barLabelSize, Weight: weight, Color: colour})
	}) {
		program.Open(each.place)
	}
}

func (program *App) scrollingPage(context *ui.Context) {
	side := float32(pageSide)
	if look.Narrow(context) {
		side = narrowPageSide
	}
	top := float32(pageTop)
	if program.place == Stand && program.stand.Selling() {
		top = registerTop
	}
	padding := ui.Padding{Top: top, Left: side, Right: side, Bottom: pageBottom}
	context.ScrollArea("page", ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: padding}, func() {
			program.page(context)
		})
	})
}

// spaced writes the brand in capitals with Bubble's letter spacing, letter by letter.
func spaced(context *ui.Context, text string, size float32) {
	context.Box(ui.Layout{Gap: size * brandLetter, AlignY: ui.Center}, func() {
		for _, letter := range text {
			context.Text(string(letter), ui.TextStyle{Size: size, Weight: draw2d.Bold})
		}
	})
	context.Describe(ui.Description{Role: ui.TextRole, Label: text}) // a word, not letters, for a screen reader
}
