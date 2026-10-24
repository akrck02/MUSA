package look

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// A button answers the finger with its border, a shade deeper hovered and deeper still pressed.
func TestAButtonsBorderDeepensWhenHoveredAndMoreWhenPressed(t *testing.T) {
	theme := ui.LightTheme()
	normal := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1}
	hovered := normal
	hovered.Fill = theme.Sunken
	looks := Answering(theme, normal, hovered)
	darkness := func(colour draw2d.Color) float32 { return 3 - colour[0] - colour[1] - colour[2] }
	if looks.Normal.Border != theme.Line {
		t.Errorf("at rest the border changed to %v", looks.Normal.Border)
	}
	if !(darkness(looks.Hovered.Border) > darkness(theme.Line) && darkness(looks.Pressed.Border) > darkness(looks.Hovered.Border)) {
		t.Errorf("the border does not deepen: %v hovered, %v pressed", looks.Hovered.Border, looks.Pressed.Border)
	}
}
