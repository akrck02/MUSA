package look

import "github.com/akrck02/go-renderer/draw2d"

// Fade paints a box from clear at its top to Colour further down, as CSS's
// linear-gradient(transparent, colour Solid%) does: what scrolls under it melts into the page
// instead of meeting a hard edge. Keep one from frame to frame, so drawing it allocates nothing.
type Fade struct {
	Colour draw2d.Color
	// Solid is the share of the height, from the top, at which Colour is reached; below it the box
	// is solid.
	Solid float32
}

// Paint draws the fade a pixel-tall strip at a time.
func (fade *Fade) Paint(area draw2d.Rectangle, drawing *draw2d.List) {
	ramp := area.Height * fade.Solid
	for offset := float32(0); offset < ramp; offset++ {
		strip := draw2d.Rectangle{X: area.X, Y: area.Y + offset, Width: area.Width, Height: 1}
		colour := fade.Colour
		colour[3] *= (offset + 0.5) / ramp
		drawing.Rectangle(strip, 0, colour, draw2d.Color{}, 0)
	}
	solid := draw2d.Rectangle{X: area.X, Y: area.Y + ramp, Width: area.Width, Height: area.Height - ramp}
	drawing.Rectangle(solid, 0, fade.Colour, draw2d.Color{}, 0)
}
