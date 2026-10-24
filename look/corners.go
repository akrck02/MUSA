package look

import "github.com/akrck02/go-renderer/draw2d"

// Corner names one corner of a box, for Corners.
type Corner int

// The corners of a box, as flags that can be added: TopLeft | BottomRight.
const (
	TopLeft Corner = 1 << iota
	TopRight
	BottomLeft
	BottomRight
)

// Corners paints a box's fill with only some of its corners rounded: the cell in a corner of a
// rounded table, which CSS would cut with the table's overflow. Keep one per box from frame to
// frame, so drawing it allocates nothing.
type Corners struct {
	Fill    draw2d.Color
	Radius  float32
	Rounded Corner
}

// Paint draws the fill: a rounded rectangle, and a square over each corner that stays square.
func (shape *Corners) Paint(area draw2d.Rectangle, drawing *draw2d.List) {
	drawing.Rectangle(area, shape.Radius, shape.Fill, draw2d.Color{}, 0)
	side := min(shape.Radius, area.Width/2, area.Height/2)
	right, bottom := area.X+area.Width-side, area.Y+area.Height-side
	squares := []struct {
		corner Corner
		x, y   float32
	}{{TopLeft, area.X, area.Y}, {TopRight, right, area.Y}, {BottomLeft, area.X, bottom}, {BottomRight, right, bottom}}
	for _, square := range squares {
		if shape.Rounded&square.corner == 0 {
			drawing.Rectangle(draw2d.Rectangle{X: square.x, Y: square.y, Width: side, Height: side}, 0, shape.Fill, draw2d.Color{}, 0)
		}
	}
}
