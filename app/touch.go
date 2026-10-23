package app

import ui "github.com/akrck02/go-ui"

// touchScale is how much bigger than Bubble's comfortable size MUSA draws on a tablet:
// it is used with a finger at arm's length, by someone who should never have to squint or aim.
const touchScale = 1.2

// forTouch scales a theme's text, icons and controls for a finger on a tablet.
//
// TODO(akrck02): ask Bubble for a touch density, and use it here instead.
func forTouch(theme ui.Theme) ui.Theme {
	theme.TextSize *= touchScale
	theme.SmallTextSize *= touchScale
	theme.HeadingSize *= touchScale
	theme.IconSize *= touchScale
	theme.ControlHeight *= touchScale
	theme.FlatHeight *= touchScale
	theme.Gap *= touchScale
	theme.Padding *= touchScale
	theme.ButtonPadding *= touchScale
	return theme
}
