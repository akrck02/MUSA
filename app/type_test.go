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

// The field being typed in is lit softly, in a shade between the line and the accent: a dark ring
// would be the only hard line on the page.
func TestTheFocusedFieldIsLitBetweenTheLineAndTheAccent(t *testing.T) {
	theme := pageTheme(ui.LightTheme())
	if theme.FocusRing != theme.Line.Mix(theme.Primary, focusShade) {
		t.Errorf("the focused field is lit with %v", theme.FocusRing)
	}
}
