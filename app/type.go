package app

import ui "github.com/akrck02/go-ui"

// The type of the author's design, in pixels: its body text, its small text and how tall a line
// is, in times its size, as the page's CSS sets them.
const (
	bodySize   = 14
	smallSize  = 13
	lineHeight = 1.5
	focusShade = 0.55 // how far a focused field's line goes towards the accent
)

// pageTheme is Bubble's theme with the text as the author's design sets it: a body of 14 pixels
// on lines one and a half times as tall, big or small, and fields and buttons bordered by the soft
// line of cards. Titles set their own line height.
func pageTheme(theme ui.Theme) ui.Theme {
	theme.TextSize, theme.SmallTextSize = bodySize, smallSize
	theme.LineHeight, theme.HeadingLineHeight = lineHeight, lineHeight
	// The author's design borders fields and buttons with the same soft line as cards.
	theme.FieldLine, theme.ControlLine = theme.Line, theme.Line
	// The field being typed in is lit by its line deepened towards the accent, not by a dark ring.
	theme.FocusRing = theme.Line.Mix(theme.Primary, focusShade)
	return theme
}
