package stand

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
)

// Measures of a product tile, in pixels, as in the author's design.
const (
	tileRadius        = 18
	tileBorder        = 1
	pictureHeight     = 132
	letterSize        = 48
	letterAlpha       = 0.6
	countSide         = 34
	countInset        = 10
	countSize         = 17
	wordsPaddingX     = 14
	wordsPaddingTop   = 12
	wordsPaddingBelow = 14
	nameSize          = 17
	rowSpace          = 6
	priceSize         = 17
	chosenPriceAlpha  = 0.8
	leftSize          = 13
	leftPaddingX      = 10
	leftPaddingY      = 3
	inBasketAlpha     = 0.85
	minusSide         = 30
	minusSize         = 20
	minusAlpha        = 0.15
	soldOutOpacity    = 0.5
	roundRadius       = 999
	versionGap        = 6
	lowStock          = 3 // from this many left down, the tile warns
)

// tile is one product: a picture, its name, price and what is left. Tapping it adds one to the
// basket; once chosen, the tile turns the accent's colour with «en la cesta» and «−», and the
// picture carries how many. A sold-out tile is faded and does nothing.
func (screens *Screens) tile(context *ui.Context, item Item, fill draw2d.Color) {
	theme := context.Theme
	chosen := screens.basket.Quantity(item.VariantID)
	paint := tilePaint(theme, chosen > 0, item.Left == 0)
	label := fmt.Sprintf("%s · %s", item.Label, money.Format(item.Price))
	normal := ui.Look{Fill: paint.surface, Border: paint.border, BorderWidth: tileBorder, Radius: tileRadius, Clip: true}
	pressed := normal
	pressed.Border = theme.ControlLine
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(tileBorder)}
	if context.Clickable(label, layout, ui.Looks{Normal: normal, Hovered: normal, Pressed: pressed}, func() {
		screens.picture(context, item, chosen, paint.fade(fill))
		screens.tileWords(context, item, chosen, paint)
	}) && !screens.basket.Add(item) {
		screens.tell(true, "No quedan más")
	}
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: label})
}

// paint is the colours of a tile: plain, chosen (on the accent) or sold out (faded halfway into
// the page, as CSS's opacity would).
type paint struct {
	surface, border, ink, soft, quiet, sunken draw2d.Color
	background                                draw2d.Color
	faded                                     bool
}

func tilePaint(theme ui.Theme, chosen, soldOut bool) paint {
	colours := paint{surface: theme.Surface, border: theme.Line, ink: theme.Ink, soft: theme.InkSoft, quiet: theme.InkQuiet,
		sunken: theme.Sunken, background: theme.Background}
	if chosen {
		colours.surface, colours.border, colours.ink = theme.Primary, theme.Primary, theme.OnPrimary
		colours.soft = theme.Primary.Mix(theme.OnPrimary, chosenPriceAlpha)
	}
	if soldOut {
		colours.faded = true
		colours.surface, colours.border = colours.fade(colours.surface), colours.fade(colours.border)
		colours.ink, colours.soft, colours.quiet = colours.fade(colours.ink), colours.fade(colours.soft), colours.fade(colours.quiet)
		colours.sunken = colours.fade(colours.sunken)
	}
	return colours
}

// fade is a colour seen through a faded tile.
func (colours paint) fade(colour draw2d.Color) draw2d.Color {
	if !colours.faded {
		return colour
	}
	return colours.background.Mix(colour, soldOutOpacity)
}

// pictureShapes are the top-rounded fills of the tiles' pictures, kept from frame to frame.
func (screens *Screens) pictureShape(fill draw2d.Color) *look.TopRounded {
	if screens.shapes == nil {
		screens.shapes = map[draw2d.Color]*look.TopRounded{}
	}
	shape, found := screens.shapes[fill]
	if !found {
		shape = &look.TopRounded{Fill: fill, Radius: tileRadius - tileBorder}
		screens.shapes[fill] = shape
	}
	return shape
}

// picture is the coloured top of a tile with the product's first letter, and the count chosen.
func (screens *Screens) picture(context *ui.Context, item Item, chosen int, fill draw2d.Color) {
	theme := context.Theme
	layout := ui.Layout{Direction: ui.Stacked, Width: ui.Grow(), Height: ui.Fixed(pictureHeight)}
	context.Frame(layout, ui.Look{Backdrop: screens.pictureShape(fill)}, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow(), AlignX: ui.Center, AlignY: ui.Center}, func() {
			context.Text(catalogue.Initial(item.Name), ui.TextStyle{Size: letterSize, Weight: draw2d.Bold, Color: draw2d.Color{1, 1, 1, letterAlpha}})
		})
		if chosen == 0 {
			return
		}
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(countInset), AlignX: ui.End}, func() {
			side := ui.Fixed(countSide)
			count := ui.Layout{Width: side, Height: side, AlignX: ui.Center, AlignY: ui.Center}
			context.Frame(count, ui.Look{Fill: theme.Surface, Radius: countSide / 2, Shadow: true}, func() {
				context.Text(fmt.Sprint(chosen), ui.TextStyle{Size: countSize, Weight: draw2d.Bold, Color: theme.Primary})
			})
		})
	})
}

// tileWords is the lower half of a tile: the name, then the price and what is left, or, once
// chosen, the price, «en la cesta» and «−».
func (screens *Screens) tileWords(context *ui.Context, item Item, chosen int, colours paint) {
	padding := ui.Padding{Left: wordsPaddingX, Right: wordsPaddingX, Top: wordsPaddingTop, Bottom: wordsPaddingBelow}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: padding}, ui.Look{Clip: true}, func() {
		context.Box(ui.Layout{Gap: versionGap, AlignY: ui.Center}, func() {
			context.Text(item.Name, ui.TextStyle{Size: nameSize, Weight: draw2d.Bold, Color: colours.ink})
			if item.Version != "" {
				context.Text(item.Version, ui.TextStyle{Size: leftSize, Color: colours.soft})
			}
		})
		look.Space(context, rowSpace)
		context.Box(ui.Layout{Width: ui.Grow(), AlignX: ui.SpaceBetween, AlignY: ui.Center}, func() {
			context.Text(money.Format(item.Price), ui.TextStyle{Size: priceSize, Color: colours.soft})
			if chosen > 0 {
				screens.inBasket(context, item, colours)
				return
			}
			leftPill(context, item.Left, colours)
		})
	})
}

// inBasket is «en la cesta» and the round «−» that takes one out.
func (screens *Screens) inBasket(context *ui.Context, item Item, colours paint) {
	context.Text(screens.inBasketWords(), ui.TextStyle{Size: leftSize, Color: colours.surface.Mix(colours.ink, inBasketAlpha)})
	side := ui.Fixed(minusSide)
	layout := ui.Layout{Width: side, Height: side, AlignX: ui.Center, AlignY: ui.Center}
	round := ui.Look{Fill: draw2d.Color{1, 1, 1, minusAlpha}, Radius: minusSide / 2}
	clicked := context.Clickable("Quitar uno", layout, ui.Looks{Normal: round, Hovered: round, Pressed: round}, func() {
		context.Text("−", ui.TextStyle{Size: minusSize, Color: colours.ink, LineHeight: 1})
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: "Quitar uno"}) // a sign alone says nothing
	if clicked {
		screens.basket.Remove(item.VariantID)
	}
}

// leftPill is what is left in a small sunken pill: warm when few, red when none.
func leftPill(context *ui.Context, left int, colours paint) {
	theme := context.Theme
	ink := colours.quiet
	switch {
	case left == 0:
		ink = colours.fade(theme.Danger)
	case left <= lowStock:
		ink = theme.Warning
	}
	layout := ui.Layout{Padding: ui.PadXY(leftPaddingX, leftPaddingY)}
	context.Frame(layout, ui.Look{Fill: colours.sunken, Radius: roundRadius}, func() {
		context.Text(leftWords(left), ui.TextStyle{Size: leftSize, Color: ink})
	})
}

func leftWords(left int) string {
	switch {
	case left == 0:
		return "Agotado"
	case left <= lowStock:
		return fmt.Sprintf("Quedan %d", left)
	}
	return fmt.Sprint(left)
}
