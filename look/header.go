package look

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// Measures of a page's header and its buttons, in pixels, as in the author's design.
const (
	titleSize          = 34
	titleLineHeight    = 1.2
	subtitleSize       = 17
	subtitleSpace      = 4
	headerGap          = 16
	toolButtonSide     = 52
	toolButtonRadius   = 12
	mainButtonRadius   = 14
	mainButtonPaddingX = 22
	mainButtonPaddingY = 14
	mainButtonGap      = 10
	mainButtonTextSize = 18
	headerIconSize     = 22
)

// PageHeader is the top of a page: its title and a soft line under it on the left, and the page's
// actions on the right, standing on the same line as the bottom of the words (actions may be nil).
func PageHeader(context *ui.Context, title, subtitle string, actions func()) {
	theme := context.Theme
	context.Box(ui.Layout{Direction: Across(context), Width: ui.Grow(), Gap: headerGap, AlignY: ui.End}, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow()}, func() {
			context.Text(title, ui.TextStyle{Size: titleSize, Weight: draw2d.Bold, LineHeight: titleLineHeight})
			if subtitle != "" {
				Space(context, subtitleSpace)
				Paragraph(context, subtitle, ui.TextStyle{Size: subtitleSize, Color: theme.InkSoft})
			}
		})
		if actions != nil {
			context.Box(ui.Layout{Gap: headerGap, AlignY: ui.Center}, actions)
		}
	})
}

// Space is an empty gap of pixels between the things above and below it, or before and after.
func Space(context *ui.Context, pixels float32) {
	context.Box(ui.Layout{Width: ui.Fixed(pixels), Height: ui.Fixed(pixels)}, nil)
}

// ToolButton is a square button with only an icon, on the surface with a fine line, for a page's
// tools (searching); a screen reader calls it label. It returns true on the frame it is clicked.
func ToolButton(context *ui.Context, icon, label string) bool {
	theme := context.Theme
	normal := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: cardBorder, Radius: toolButtonRadius}
	hovered := normal
	hovered.Border = theme.ControlLine
	side := ui.Fixed(toolButtonSide)
	layout := ui.Layout{Width: side, Height: side, AlignX: ui.Center, AlignY: ui.Center}
	clicked := context.Clickable(label, layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
		context.IconSized(icon, headerIconSize, theme.Ink)
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: label})
	return clicked
}

// MainButton is a page's main action: its icon and words on the accent. It returns true on the
// frame it is clicked.
func MainButton(context *ui.Context, icon, words string) bool {
	theme := context.Theme
	normal := ui.Look{Fill: theme.Primary, Radius: mainButtonRadius}
	hovered := normal
	hovered.Fill = theme.Primary.Mix(theme.Ink, hoverDarkening)
	layout := ui.Layout{Padding: ui.PadXY(mainButtonPaddingX, mainButtonPaddingY), Gap: mainButtonGap, AlignY: ui.Center}
	clicked := context.Clickable(words, layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
		context.IconSized(icon, headerIconSize, theme.OnPrimary)
		context.Text(words, ui.TextStyle{Size: mainButtonTextSize, Weight: draw2d.Bold, Color: theme.OnPrimary, LineHeight: ui.LettersLineHeight})
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: words})
	return clicked
}

// hoverDarkening is how much ink a main button takes under the pointer.
const hoverDarkening = 0.15
