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
	figureSize  = 28
	pillHeight  = 28
	pillPadding = 10
	emptyIcon   = 40
	cardBorder  = 1
)

// Card is Bubble's card on paper: the surface, a fine line around it and the panels' corners.
func Card(context *ui.Context, layout ui.Layout, content func()) {
	context.Frame(layout, cardLook(context.Theme), content)
}

// ClickableCard is a Card the whole of which is one button, which a screen reader calls label; it
// returns true on the frame it is clicked. Hovering brightens its edge, as Bubble's hover does; it
// never moves.
func ClickableCard(context *ui.Context, label string, layout ui.Layout, content func()) bool {
	theme := context.Theme
	normal := cardLook(theme)
	hovered := normal
	hovered.Border = theme.ControlLine
	pressed := hovered
	pressed.Fill = theme.Sunken
	clicked := context.Clickable(label, layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: pressed}, content)
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: label})
	return clicked
}

func cardLook(theme ui.Theme) ui.Look {
	return ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: cardBorder, Radius: theme.PanelRadius}
}

// Figures is one quiet card holding a row of figures, so a summary reads as one thing rather
// than a wall of boxes. On a narrow window the row folds into two columns.
func Figures(context *ui.Context, count int, figures func()) {
	theme := context.Theme
	columns := count
	if Narrow(context) && count > 3 {
		columns = 2
	}
	Card(context, ui.Layout{Width: ui.Grow(), Columns: columns, Padding: ui.Pad(theme.Padding), Gap: theme.Gap}, figures)
}

// Figure is one number of a summary: its label, small and soft, above it in bold. A negative
// figure (an amount lost) is painted with the theme's danger colour.
func Figure(context *ui.Context, label, value string, negative bool) {
	theme := context.Theme
	colour := theme.Ink
	if negative {
		colour = theme.Danger
	}
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 4}, func() {
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
// not empty, a button; it returns true on the frame that button is clicked. When the page cannot
// be used until the artist does something, the line says what and the button does it.
func Empty(context *ui.Context, icon, title, message, action string) bool {
	theme := context.Theme
	clicked := false
	Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(theme.Padding * 2), Gap: theme.Gap, AlignX: ui.Center}, func() {
		context.IconSized(icon, emptyIcon, theme.InkQuiet)
		context.Text(title, ui.TextStyle{Size: theme.HeadingSize, Weight: draw2d.Bold})
		CentredParagraph(context, message, ui.TextStyle{Color: theme.InkSoft})
		if action != "" {
			clicked = context.PrimaryButton(action)
		}
	})
	return clicked
}

// Section is one part of a form or a detail page: a small, soft title above a card, so the
// content leads and the titles stay out of its way.
func Section(context *ui.Context, title string, content func()) {
	theme := context.Theme
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 8}, func() {
		Label(context, title)
		Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(theme.Padding), Gap: theme.Gap}, content)
	})
}

// Label is the small, soft title of a group: a section, or a part of a list.
func Label(context *ui.Context, title string) {
	theme := context.Theme
	context.Text(title, ui.TextStyle{Size: theme.SmallTextSize, Weight: draw2d.Bold, Color: theme.InkSoft})
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

// wordSpace is the room between two words of a paragraph, in ems: a space of the system's font.
const wordSpace = 0.318

// Paragraph writes text that breaks into lines to fit its box, word by word: go-ui's Text keeps
// one line. Its lines are as far apart as the theme's line height says, as CSS's are.
//
// TODO(akrck02): ask go-ui for text that wraps, and use it here.
func Paragraph(context *ui.Context, text string, style ui.TextStyle) {
	paragraph(context, text, style, ui.Start)
}

// CentredParagraph is a Paragraph whose lines stand in the middle of its box.
func CentredParagraph(context *ui.Context, text string, style ui.TextStyle) {
	paragraph(context, text, style, ui.Center)
}

func paragraph(context *ui.Context, text string, style ui.TextStyle, align ui.Align) {
	size := style.Size
	if size == 0 {
		size = context.Theme.TextSize
	}
	context.Box(ui.Layout{Width: ui.Grow(), Wrap: true, Gap: size * wordSpace, LineGap: ui.NoLineGap, AlignX: align}, func() {
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

// wideWidth is the window width from which a page can hold two panes side by side, in pixels: a
// large tablet held across (an iPad of 13 inches is 1366 wide), a laptop, a desktop.
const wideWidth = 1180

// Wide tells whether the window is wide enough for two panes side by side.
func Wide(context *ui.Context) bool {
	width, _ := context.Size()
	return width >= wideWidth
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

// Badge is an icon in a soft round patch of the accent, the picture of a place or an action.
func Badge(context *ui.Context, icon string, size float32) {
	theme := context.Theme
	context.Frame(ui.Layout{Width: ui.Fixed(size), Height: ui.Fixed(size), AlignX: ui.Center, AlignY: ui.Center},
		ui.Look{Fill: theme.Surface.Mix(theme.Primary, badgeTint), Radius: size / 2}, func() {
			context.IconSized(icon, size*badgeIcon, theme.Primary)
		})
}

// badgeTint is how much of the accent a badge's patch takes; badgeIcon is its icon's share of it.
const (
	badgeTint = 0.12
	badgeIcon = 0.5
)

// meterHeight is how thick a meter is, in pixels.
const meterHeight = 6

// Meter is a thin bar of how much of a whole is done: units sold of those taken.
func Meter(context *ui.Context, done, whole int) {
	theme := context.Theme
	share := float32(0)
	if whole > 0 {
		share = float32(min(done, whole)) / float32(whole)
	}
	context.Frame(ui.Layout{Width: ui.Grow(), Height: ui.Fixed(meterHeight)}, ui.Look{Fill: theme.Sunken, Radius: meterHeight / 2}, func() {
		if share > 0 {
			context.Frame(ui.Layout{Width: ui.Percent(share), Height: ui.Grow()}, ui.Look{Fill: theme.Primary, Radius: meterHeight / 2}, nil)
		}
	})
}
