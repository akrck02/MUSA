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
	if !(first < moving && moving < there) || there+slidingPills["sliding"].row.X != target.Rect.X {
		test.Errorf("the pill went %v → %v → %v, not from the first tab to «Tres» at %v", first, moving, there, target.Rect.X)
	}
}

func TestTheChosenPillMovesWithItsRowWhenThePageScrolls(test *testing.T) {
	selected, top := 1, float32(0)
	tabs := []Tab{{Label: "Uno"}, {Label: "Dos"}}
	window, err := uitest.New(func(context *ui.Context) {
		context.Box(ui.Layout{Direction: ui.Vertical, Padding: ui.Padding{Top: top}}, func() {
			Segmented(context, "scrolled", &selected, tabs, Views)
		})
	}, uitest.Size(800, 400), uitest.Test(test))
	if err != nil {
		test.Fatal(err)
	}
	window.Frames(3)
	top = 120      // as if the page scrolled: the row is somewhere else, the tab is still chosen
	window.Frame() // the very frame it moves: the pill must not lag behind by one
	pill := slidingPills["scrolled"]
	target, _ := window.ElementIn("scrolled", "Dos")
	if pill.row.Y+pill.area.Y != target.Rect.Y || pill.row.X+pill.area.X != target.Rect.X {
		test.Errorf("the pill is at %v,%v, not under «Dos» at %v,%v", pill.row.X+pill.area.X, pill.row.Y+pill.area.Y, target.Rect.X, target.Rect.Y)
	}
}
