package app

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/look"
)

// Measures of the frame, in pixels, as in Bubble's showcase.
const (
	barHeight   = 88
	barWidth    = 900 // the places' width on a wide window
	pillWidth   = 64  // the soft pill behind the chosen place's icon
	pillHeight  = 32
	brandLetter = 0.14 // the brand's letter spacing, in ems
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
	{Fairs, "tent", "Ferias"},
	{Sales, "package", "Pedidos"},
}

// frame is the page above and the places in a bar along the bottom, on every size, as on an
// Android phone: one way to move around, always in the same place, within reach of a thumb.
func (program *App) frame(context *ui.Context) {
	theme := context.Theme
	insets := context.Insets()
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Padding{Top: insets.Top}}, func() {
		program.scrollingPage(context)
		context.Separator()
		bar := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(barHeight + insets.Bottom), Padding: ui.Padding{Bottom: insets.Bottom}, AlignX: ui.Center}
		context.Frame(bar, ui.Look{Fill: theme.Surface}, func() {
			// On a wide window the places stay together in the middle, not spread to the edges.
			width := ui.Grow()
			if !look.Narrow(context) {
				width = ui.Fixed(barWidth)
			}
			context.Box(ui.Layout{Width: width, Height: ui.Grow()}, func() {
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
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Gap: 4, AlignX: ui.Center, AlignY: ui.Center}
	looks := ui.Looks{Pressed: ui.Look{Fill: theme.Sunken}}
	if context.Clickable(each.name+"##bar", layout, looks, func() {
		context.Frame(ui.Layout{Width: ui.Fixed(pillWidth), Height: ui.Fixed(pillHeight), AlignX: ui.Center, AlignY: ui.Center},
			ui.Look{Fill: pill, Radius: pillHeight / 2}, func() {
				context.IconSized(each.icon, theme.IconSize, colour)
			})
		context.Text(each.name, ui.TextStyle{Size: theme.SmallTextSize, Weight: weight, Color: colour})
	}) {
		program.Open(each.place)
	}
}

func (program *App) scrollingPage(context *ui.Context) {
	theme := context.Theme
	padding := theme.Padding * 1.6
	if look.Narrow(context) {
		padding = theme.Padding
	}
	context.ScrollArea("page", ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(padding), Gap: theme.Gap}, func() {
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
