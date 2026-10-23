package app

import ui "github.com/akrck02/go-ui"

// The type of the author's design, in pixels: its body text, its small text and how tall a line
// is, in times its size, as the page's CSS sets them.
const (
	bodySize   = 14
	smallSize  = 13
	lineHeight = 1.5
)

// pageTheme is Bubble's theme with the text as the author's design sets it: a body of 14 pixels
// on lines one and a half times as tall, big or small. Titles set their own line height.
func pageTheme(theme ui.Theme) ui.Theme {
	theme.TextSize, theme.SmallTextSize = bodySize, smallSize
	theme.LineHeight, theme.HeadingLineHeight = lineHeight, lineHeight
	return theme
}
