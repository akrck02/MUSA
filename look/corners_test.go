package look

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
)

func TestOnlyTheChosenCornersOfABoxAreRounded(test *testing.T) {
	shape := &Corners{Fill: draw2d.Color{1, 1, 1, 1}, Radius: 10, Rounded: BottomLeft}
	var drawing draw2d.List
	shape.Paint(draw2d.Rectangle{Width: 100, Height: 50}, &drawing)
	if len(drawing.Commands) != 4 {
		test.Fatalf("got %d shapes, want a rounded box and a square over each of three corners", len(drawing.Commands))
	}
	for _, square := range drawing.Commands[1:] {
		if square.Area.X == 0 && square.Area.Y == 40 {
			test.Errorf("the bottom-left corner was squared: %+v", square.Area)
		}
	}
}
