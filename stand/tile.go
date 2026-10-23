package stand

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
)

// Measures of a product tile, in pixels.
const (
	pictureHeight = 96
	letterSize    = 56
	countSize     = 30
	lowStock      = 3 // from this many left down, the tile warns
)

// tile is one product: a picture, its name, price and what is left. Tapping it adds one to the
// basket; once chosen, its lower half turns slate with «en la cesta» and «−», and the picture
// carries how many. A sold-out tile is greyed and does nothing.
func (screens *Screens) tile(context *ui.Context, item Item) {
	theme := context.Theme
	chosen := screens.basket.Quantity(item.VariantID)
	label := fmt.Sprintf("%s · %s", item.Label, money.Format(item.Price))
	normal := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: theme.PanelRadius, Clip: true}
	pressed := normal
	pressed.Fill = theme.Sunken
	pressed.Border = theme.ControlLine
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow()}
	if context.Clickable(label, layout, ui.Looks{Normal: normal, Hovered: normal, Pressed: pressed}, func() {
		// The frame inside cuts the picture and the slate band to the card's rounded corners.
		context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(1)}, ui.Look{Radius: theme.PanelRadius, Clip: true}, func() {
			picture(context, item, chosen)
			screens.tileWords(context, item, chosen)
		})
	}) && !screens.basket.Add(item) {
		screens.tell(true, "No quedan más")
	}
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: label})
}

// picture is the coloured top of a tile with the product's first letter, and the count chosen.
func picture(context *ui.Context, item Item, chosen int) {
	theme := context.Theme
	fill := catalogue.PictureFill(theme, item.Name)
	if item.Left == 0 {
		fill = theme.Sunken
	}
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Fixed(pictureHeight), Padding: ui.Pad(8), AlignX: ui.Center}
	context.Frame(layout, ui.Look{Fill: fill}, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Fixed(countSize), AlignX: ui.End}, func() {
			if chosen > 0 {
				count := ui.Layout{Width: ui.Fixed(countSize), Height: ui.Fixed(countSize), AlignX: ui.Center, AlignY: ui.Center}
				context.Frame(count, ui.Look{Fill: theme.Surface, Radius: countSize / 2}, func() {
					context.Text(fmt.Sprint(chosen), ui.TextStyle{Weight: draw2d.Bold})
				})
			}
		})
		context.Text(catalogue.Initial(item.Name), ui.TextStyle{Size: letterSize, Weight: draw2d.Bold, Color: theme.Surface.WithAlpha(0.6)})
	})
}

// tileWords is the lower half of a tile: the name, the price and what is left, or, once chosen,
// the same on slate with «−».
func (screens *Screens) tileWords(context *ui.Context, item Item, chosen int) {
	theme := context.Theme
	fill, ink, soft := theme.Surface, theme.Ink, theme.InkSoft
	if chosen > 0 {
		fill, ink, soft = theme.Primary, theme.OnPrimary, theme.OnPrimary.WithAlpha(0.8)
	}
	if item.Left == 0 {
		ink, soft = theme.InkQuiet, theme.InkQuiet
	}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(10), Gap: 4}, ui.Look{Fill: fill}, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, AlignY: ui.Center}, func() {
			context.Text(item.Name, ui.TextStyle{Weight: draw2d.Bold, Color: ink})
			if item.Version != "" {
				context.Text(item.Version, ui.TextStyle{Size: theme.SmallTextSize, Color: soft})
			}
		})
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, AlignY: ui.Center}, func() {
			context.Text(money.Format(item.Price), ui.TextStyle{Size: theme.SmallTextSize, Color: soft})
			context.Spacer()
			if chosen > 0 {
				context.Text("en la cesta", ui.TextStyle{Size: theme.SmallTextSize, Color: soft})
				if context.IconButton("minus", "##Quitar uno") {
					screens.basket.Remove(item.VariantID)
				}
				return
			}
			look.Pill(context, leftWords(item.Left-chosen), leftTone(theme, item.Left))
		})
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

func leftTone(theme ui.Theme, left int) look.Tone {
	switch {
	case left == 0:
		return look.Danger(theme)
	case left <= lowStock:
		return look.Warning(theme)
	}
	return look.Neutral(theme)
}
