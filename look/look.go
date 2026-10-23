// Package look holds the few pieces of Bubble 4 that MUSA's screens need and bubble/goui does not
// give yet: a card, a page's header, a figure with its label, a pill and an empty state. Each one
// is drawn the way Bubble's showcase draws it (bubble/goui/cmd/showcase), only from the theme's
// tokens, so swapping a piece for Bubble's own, once it has it, changes no screen.
//
// TODO(akrck02): ask the Bubble session to move these into bubble/goui, then use them from there.
package look

import (
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// Measures of these pieces that are not tokens of the theme, in pixels, as in Bubble's CSS.
const (
	titleSize   = 26
	figureSize  = 22
	pillHeight  = 24
	pillPadding = 10
	emptyIcon   = 40
	cardBorder  = 1
)

// Card is Bubble's card on paper: the surface, a fine line around it and the panels' corners.
func Card(context *ui.Context, layout ui.Layout, content func()) {
	context.Frame(layout, cardLook(context.Theme), content)
}

// ClickableCard is a Card the whole of which is one button; it returns true on the frame it is
// clicked. Hovering brightens it a step, as Bubble's hover does; it never moves.
func ClickableCard(context *ui.Context, name string, layout ui.Layout, content func()) bool {
	theme := context.Theme
	normal := cardLook(theme)
	hovered := normal
	hovered.Border = theme.ControlLine
	pressed := hovered
	pressed.Fill = theme.Sunken
	return context.Clickable(name, layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: pressed}, content)
}

func cardLook(theme ui.Theme) ui.Look {
	return ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: cardBorder, Radius: theme.PanelRadius}
}

// PageHeader is the top of a page: its title and a soft line under it on the left, and the page's
// actions on the right (actions may be nil).
func PageHeader(context *ui.Context, title, subtitle string, actions func()) {
	theme := context.Theme
	context.Box(ui.Layout{Direction: Across(context), Width: ui.Grow(), Gap: theme.Gap, AlignY: ui.Center}, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, func() {
			context.Text(title, ui.TextStyle{Size: titleSize, Weight: draw2d.Bold})
			if subtitle != "" {
				Paragraph(context, subtitle, ui.TextStyle{Color: theme.InkSoft})
			}
		})
		if actions != nil {
			context.Box(ui.Layout{Gap: 8, AlignY: ui.Center}, actions)
		}
	})
}

// Figure is one number of a summary: its label, small and soft, above it in bold. A negative
// figure (an amount lost) is painted with the theme's danger colour.
func Figure(context *ui.Context, label, value string, negative bool) {
	theme := context.Theme
	colour := theme.Ink
	if negative {
		colour = theme.Danger
	}
	Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(theme.Padding * 0.75), Gap: 4}, func() {
		context.Text(label, ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
		context.Text(value, ui.TextStyle{Size: figureSize, Weight: draw2d.Bold, Color: colour})
	})
}

// Tone is the colours of a pill.
type Tone struct{ Fill, Ink draw2d.Color }

// Neutral is the tone of a pill that only informs: the sunken well and the soft ink.
func Neutral(theme ui.Theme) Tone { return Tone{Fill: theme.Sunken, Ink: theme.InkSoft} }

// Warning is the tone of something that needs care soon, such as low stock.
func Warning(theme ui.Theme) Tone { return Tone{Fill: theme.Warning.WithAlpha(0.18), Ink: theme.Ink} }

// Danger is the tone of something wrong now, such as a sold-out product.
func Danger(theme ui.Theme) Tone { return Tone{Fill: theme.Danger.WithAlpha(0.16), Ink: theme.Ink} }

// Positive is the tone of something going well, such as stock to spare.
func Positive(theme ui.Theme) Tone { return Tone{Fill: theme.Positive.WithAlpha(0.16), Ink: theme.Ink} }

// Pill is a short label in a rounded well of its tone.
func Pill(context *ui.Context, text string, tone Tone) {
	layout := ui.Layout{Height: ui.Fixed(pillHeight), Padding: ui.PadXY(pillPadding, 0), AlignY: ui.Center}
	context.Frame(layout, ui.Look{Fill: tone.Fill, Radius: pillHeight / 2}, func() {
		context.Text(text, ui.TextStyle{Size: context.Theme.SmallTextSize, Weight: draw2d.Bold, Color: tone.Ink})
	})
}

// Empty fills a page with nothing to show yet: an icon, a title, a line of help and, when action is
// not empty, a button; it returns true on the frame that button is clicked.
func Empty(context *ui.Context, icon, title, message, action string) bool {
	theme := context.Theme
	clicked := false
	Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(theme.Padding * 2), Gap: theme.Gap, AlignX: ui.Center}, func() {
		context.IconSized(icon, emptyIcon, theme.InkQuiet)
		context.Text(title, ui.TextStyle{Size: theme.HeadingSize, Weight: draw2d.Bold})
		Paragraph(context, message, ui.TextStyle{Color: theme.InkSoft})
		if action != "" {
			clicked = context.PrimaryButton(action)
		}
	})
	return clicked
}

// Section is a card with a heading, for one part of a form or a detail page.
func Section(context *ui.Context, title string, content func()) {
	theme := context.Theme
	Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(theme.Padding), Gap: theme.Gap}, func() {
		context.Heading(title)
		content()
	})
}

// Field is a labelled text field: the label small and bold above it, and, when problem is not
// empty, the problem under it in the danger colour.
func Field(context *ui.Context, label, name string, text *string, placeholder, problem string) bool {
	theme := context.Theme
	changed := false
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 4}, func() {
		context.Text(label, ui.TextStyle{Size: theme.SmallTextSize, Weight: draw2d.Bold})
		changed = context.TextFieldWith(name, text, ui.TextFieldOptions{Placeholder: placeholder})
		if problem != "" {
			context.Text(problem, ui.TextStyle{Size: theme.SmallTextSize, Color: theme.Danger})
		}
	})
	return changed
}

// Dark tells whether the theme is a dark one, from how light its background is.
func Dark(theme ui.Theme) bool {
	background := theme.Background
	return background[0]+background[1]+background[2] < 1.5
}

// wordSpace is the room between two words of a paragraph, in ems.
const wordSpace = 0.28

// Paragraph writes text that breaks into lines to fit its box, word by word: go-ui's Text keeps
// one line.
//
// TODO(akrck02): ask go-ui for text that wraps, and use it here.
func Paragraph(context *ui.Context, text string, style ui.TextStyle) {
	size := style.Size
	if size == 0 {
		size = context.Theme.TextSize
	}
	context.Box(ui.Layout{Width: ui.Grow(), Wrap: true, Gap: size * wordSpace}, func() {
		for _, word := range strings.Fields(text) {
			context.Text(word, style)
		}
	})
}

// narrowWidth is the window width under which MUSA lays itself out for a phone, in pixels.
const narrowWidth = 760

// Narrow tells whether the window is phone-sized.
func Narrow(context *ui.Context) bool {
	width, _ := context.Size()
	return width < narrowWidth
}

// Columns is how many columns a grid of wanted columns gets: on a phone, half of them, at least one.
func Columns(context *ui.Context, wanted int) int {
	if Narrow(context) {
		return max(1, wanted/2)
	}
	return wanted
}

// Across is the direction of a row that holds words and their actions: side by side on a wide
// window, the actions under the words on a phone, where long names would run into them.
func Across(context *ui.Context) ui.Direction {
	if Narrow(context) {
		return ui.Vertical
	}
	return ui.Horizontal
}
