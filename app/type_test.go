package app

import (
	"testing"

	ui "github.com/akrck02/go-ui"
)

// The author's design draws every border, of cards, fields and buttons alike, with one soft line:
// a darker border on a button stands out as a mistake.
func TestButtonsAndFieldsAreBorderedLikeCards(t *testing.T) {
	theme := pageTheme(ui.LightTheme())
	if theme.ControlLine != theme.Line || theme.FieldLine != theme.Line {
		t.Errorf("buttons are bordered with %v and fields with %v, cards with %v", theme.ControlLine, theme.FieldLine, theme.Line)
	}
}
