package look

import (
	"testing"

	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"
)

func TestTheChosenPillSlidesFromTheOldTabToTheNewOne(test *testing.T) {
	selected := 0
	tabs := []Tab{{Label: "Uno"}, {Label: "Dos"}, {Label: "Tres"}}
	window, err := uitest.New(func(context *ui.Context) {
		Segmented(context, "sliding", &selected, tabs, Views)
	}, uitest.Size(800, 200), uitest.Test(test))
	if err != nil {
		test.Fatal(err)
	}
	window.Frames(2)
	first := slidingPills["sliding"].area.X
	if err := window.Click("Tres"); err != nil {
		test.Fatal(err)
	}
	window.Wait(tabMotion / 2)
	window.Frame()
	moving := slidingPills["sliding"].area.X
	window.Wait(2 * tabMotion)
	window.Frames(2)
	target, _ := window.ElementIn("sliding", "Tres")
	there := slidingPills["sliding"].area.X
	if !(first < moving && moving < there) || there != target.Rect.X {
		test.Errorf("the pill went %v → %v → %v, not from the first tab to «Tres» at %v", first, moving, there, target.Rect.X)
	}
}
