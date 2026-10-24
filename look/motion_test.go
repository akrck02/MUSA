package look

import (
	"testing"

	ui "github.com/akrck02/go-ui"
)

func TestSomethingArrivingRisesIntoPlaceAsItShows(t *testing.T) {
	cases := []struct {
		shown, fade, shift float32
	}{
		{0, 1, 20},     // not shown yet: gone, and lower
		{0.5, 0.5, 10}, // half way
		{1, 0, 0},      // there, as it is
	}
	for _, each := range cases {
		arriving := Arriving(ui.Look{Radius: 4}, each.shown, 20)
		if arriving.Fade != each.fade || arriving.ShiftY != each.shift || arriving.Radius != 4 {
			t.Errorf("shown %v: fade %v and shift %v (radius %v), not %v and %v", each.shown, arriving.Fade, arriving.ShiftY, arriving.Radius, each.fade, each.shift)
		}
	}
}
