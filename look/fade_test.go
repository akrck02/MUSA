package look

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
)

func TestAFadeIsClearAtTheTopAndSolidBelowItsRamp(t *testing.T) {
	fade := &Fade{Colour: draw2d.Color{1, 1, 1, 1}, Solid: 0.5}
	var drawing draw2d.List
	fade.Paint(draw2d.Rectangle{Width: 10, Height: 100}, &drawing)
	first, last := drawing.Commands[0], drawing.Commands[len(drawing.Commands)-1]
	if first.Fill[3] > 0.05 {
		t.Errorf("the top strip is %v opaque, not nearly clear", first.Fill[3])
	}
	if last.Fill[3] != 1 || last.Area.Y != 50 || last.Area.Height != 50 {
		t.Errorf("the lower half is not solid: %+v", last)
	}
}
