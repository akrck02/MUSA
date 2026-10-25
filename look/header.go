package look

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// Measures of a page's header and its buttons, in pixels, as in the author's design.
const (
	titleLineHeight    = 1.2
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

// HeaderSize is how big a page header's title and subtitle are, in pixels.
type HeaderSize struct {
	Title, Subtitle float32
}

// PageTitle is the header of most pages; Greeting is the home page's, a little bigger.
var (
	PageTitle = HeaderSize{Title: 34, Subtitle: 17}
	Greeting  = HeaderSize{Title: 36, Subtitle: 18}
)

// PageHeader is the top of a page: its title and a soft line under it on the left, and the page's
// actions on the right, standing on the same line as the bottom of the words (actions may be nil).
func PageHeader(context *ui.Context, title, subtitle string, actions func()) {
	SizedPageHeader(context, title, subtitle, PageTitle, actions)
}

// SizedPageHeader is PageHeader with words of another size.
func SizedPageHeader(context *ui.Context, title, subtitle string, size HeaderSize, actions func()) {
	theme := context.Theme
	context.Box(ui.Layout{Direction: Across(context), Width: ui.Grow(), Gap: headerGap, AlignY: ui.End}, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow()}, func() {
			context.Text(title, ui.TextStyle{Size: size.Title, Weight: draw2d.Bold, LineHeight: titleLineHeight})
			if subtitle != "" {
				Space(context, subtitleSpace)
				Paragraph(context, subtitle, ui.TextStyle{Size: size.Subtitle, Color: theme.InkSoft})
			}
		})
		if actions == nil {
			return
		}
		row := ui.Layout{Gap: headerGap, AlignY: ui.Center}
		if Narrow(context) {
			// under the words on a phone, and onto a second line when they do not fit one
			row.Width, row.Wrap = ui.Grow(), true
		}
		context.Box(row, actions)
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
	clicked := context.Clickable(label, layout, Answering(theme, normal, hovered), func() {
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
	hovered.Fill = theme.Primary.Mix(theme.Ink, HoverDarkening)
	layout := ui.Layout{Padding: ui.PadXY(mainButtonPaddingX, mainButtonPaddingY), Gap: mainButtonGap, AlignY: ui.Center}
	clicked := context.Clickable(words, layout, Answering(theme, normal, hovered), func() {
		context.IconSized(icon, headerIconSize, theme.OnPrimary)
		context.Text(words, ui.TextStyle{Size: mainButtonTextSize, Weight: draw2d.Bold, Color: theme.OnPrimary, LineHeight: ui.LettersLineHeight})
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: words})
	return clicked
}

// HoverDarkening is how much ink a button on a colour takes under the pointer.
const HoverDarkening = 0.15
