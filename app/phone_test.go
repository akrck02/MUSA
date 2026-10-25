package app

import (
	"fmt"
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-ui/uitest"
)

// A phone held upright, in layout pixels: an iPhone's or a small Android's width.
const (
	phoneWidth  = 390
	phoneHeight = 844
)

// overflowTolerance is how far past the edge a drawing may reach before it counts as cut off: a
// shadow's blur or a border's half pixel.
const overflowTolerance = 2

// TestEveryPageFitsAPhone: MUSA runs on phones too, and a page wider than the screen hides its
// right side, where the prices and buttons are.
func TestEveryPageFitsAPhone(test *testing.T) {
	for _, page := range explorePages {
		program, window := openApp(test, true)
		if err := window.Resize(phoneWidth, phoneHeight); err != nil {
			test.Fatal(err)
		}
		openPage(program, page)
		window.Frames(3)
		if text, found := firstCutOffText(window); found {
			test.Errorf("%s: %q does not fit the width of a phone", page, text)
		}
	}
}

// firstCutOffText is the first text, icon or box drawn past the window's left or right edge,
// where no inner clip keeps it inside. A text's width is the test measurer's, as the layout saw it.
func firstCutOffText(window *uitest.Window) (string, bool) {
	measurer := uitest.FixedMeasurer{}
	for _, command := range window.DrawList().Commands {
		right := command.Area.X + command.Area.Width
		switch command.Kind {
		case draw2d.TextKind:
			right = command.Area.X + measurer.MeasureText(command.Text, command.Size, command.Weight, 0).Width
		case draw2d.IconKind, draw2d.RectangleKind:
		default:
			continue
		}
		if command.HasClip && command.Clip.Width < phoneWidth-overflowTolerance {
			// an inner clip (a strip that scrolls sideways, a text cut with an ellipsis) keeps it in;
			// the page's own scroll area, as wide as the window, would only hide the cut
			right = min(right, command.Clip.X+command.Clip.Width)
		}
		if right > phoneWidth+overflowTolerance || command.Area.X < -overflowTolerance {
			return describe(command), true
		}
	}
	return "", false
}

// describe names a drawing for a failure: its text, or its kind and where it ends.
func describe(command draw2d.Command) string {
	if command.Text != "" {
		return command.Text
	}
	return fmt.Sprintf("a box %v wide at %v", command.Area.Width, command.Area.X)
}
