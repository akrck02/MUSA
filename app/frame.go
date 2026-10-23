package app

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/look"
)

// Measures of the frame, in pixels, as in Bubble's showcase.
const (
	indexWidth  = 240
	barHeight   = 64
	brandLetter = 0.14 // the brand's letter spacing, in ems
)

// destination is a place as the index shows it.
type destination struct {
	place      Place
	icon, name string
}

var destinations = []destination{
	{Home, "house", "Inicio"},
	{Catalogue, "palette", "Catálogo"},
	{Fairs, "calendar-days", "Eventos"},
	{Sales, "receipt", "Ventas"},
}

// frame is Bubble's: the index on the left, the whole height, and the page beside it; on a narrow
// window the page fills it and the places sit in a bar at the bottom.
func (program *App) frame(context *ui.Context) {
	if look.Narrow(context) {
		program.phoneFrame(context)
		return
	}
	context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow()}, func() {
		program.index(context)
		context.Separator()
		program.scrollingPage(context)
	})
}

func (program *App) phoneFrame(context *ui.Context) {
	theme := context.Theme
	insets := context.Insets()
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Padding{Top: insets.Top}}, func() {
		program.scrollingPage(context)
		context.Separator()
		bar := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(barHeight + insets.Bottom), Padding: ui.Padding{Bottom: insets.Bottom}}
		context.Frame(bar, ui.Look{Fill: theme.Surface}, func() {
			for _, each := range destinations {
				program.barItem(context, each)
			}
		})
	})
}

// barItem is one place in the phone's bar: its icon over its name, brighter and bold when chosen.
func (program *App) barItem(context *ui.Context, each destination) {
	theme := context.Theme
	chosen := each.place == program.place
	colour, weight := theme.InkSoft, draw2d.Regular
	if chosen {
		colour, weight = theme.Ink, draw2d.Bold
	}
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Gap: 2, AlignX: ui.Center, AlignY: ui.Center}
	looks := ui.Looks{Pressed: ui.Look{Fill: theme.Sunken}}
	if context.Clickable(each.name+"##bar", layout, looks, func() {
		context.IconSized(each.icon, theme.IconSize, colour)
		context.Text(each.name, ui.TextStyle{Size: theme.SmallTextSize, Weight: weight, Color: colour})
	}) {
		program.Open(each.place)
	}
}

func (program *App) index(context *ui.Context) {
	theme := context.Theme
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(indexWidth - 1), Height: ui.Grow(), Padding: ui.PadXY(12, 14), Gap: theme.Gap}, func() {
		context.Box(ui.Layout{Height: ui.Fixed(theme.ControlHeight), Padding: ui.PadXY(8, 0), Gap: 10, AlignY: ui.Center}, func() {
			context.IconSized("palette", theme.IconSize, theme.Primary)
			spaced(context, "MUSA", theme.SmallTextSize)
		})
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, func() {
			for _, each := range destinations {
				if context.Selectable(each.icon, each.name, each.place == program.place) {
					program.Open(each.place)
				}
			}
		})
	})
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
}
