package look

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// How much deeper a button's border turns towards ink, as Go UI's own buttons do: a step under
// the pointer, two while pressed.
const (
	borderHoverShade = 0.15
	borderPressShade = 0.3
)

// Answering are a button's looks from its look at rest and under the pointer: its border, or its
// fill when it has none, turns a shade deeper when hovered and deeper still when pressed, so it
// answers the finger in its own colour.
func Answering(theme ui.Theme, normal, hovered ui.Look) ui.Looks {
	edge := normal.Border
	if !edge.IsVisible() {
		edge = hovered.Fill
	}
	if !edge.IsVisible() {
		return Answering(theme, normal, hovered)
	}
	pressed := hovered
	hovered.Border, pressed.Border = shade(edge, theme.Ink, borderHoverShade), shade(edge, theme.Ink, borderPressShade)
	hovered.BorderWidth = max(hovered.BorderWidth, 1)
	pressed.BorderWidth = hovered.BorderWidth
	return ui.Looks{Normal: normal, Hovered: hovered, Pressed: pressed}
}

// shade is a colour taken towards ink by an amount from 0 to 1.
func shade(colour, ink draw2d.Color, amount float32) draw2d.Color { return colour.Mix(ink, amount) }
