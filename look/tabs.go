package look

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// Measures of the tabs, in pixels, as in the author's design.
const (
	tabCountSize        = 14
	pillRadius          = 999
)

// Tab is one choice of a row of tabs: an optional icon, its words and an optional count.
type Tab struct {
	Icon  string
	Label string
	Count int // shown small after the label when HasCount
	// HasCount shows Count, even when it is zero.
	HasCount bool
}

// SegmentSize is how big a Segmented track and its choices are, in pixels.
type SegmentSize struct {
	TrackPadding, TrackRadius float32
	// Radius, PaddingX, PaddingY and Gap shape each choice; Gap is between its icon and words.
	Radius, PaddingX, PaddingY, Gap float32
	TextSize, IconSize              float32
}

// Halves are the segments that split a page in two big halves («Físico», «Digital»); Views are
// the smaller ones that switch a page's view («Resumen», «Calendario»).
var (
	Halves = SegmentSize{TrackPadding: 5, TrackRadius: 16, Radius: 12, PaddingX: 26, PaddingY: 12, Gap: 10, TextSize: 19, IconSize: 20}
	Views  = SegmentSize{TrackPadding: 4, TrackRadius: 14, Radius: 10, PaddingX: 20, PaddingY: 10, Gap: 8, TextSize: 17, IconSize: 18}
)

// Segmented is a track with a raised, lit choice: the halves of a page, or its views. It returns
// true on the frame the choice changes.
func Segmented(context *ui.Context, name string, selected *int, tabs []Tab, size SegmentSize) bool {
	theme := context.Theme
	changed := false
	track := ui.Look{Fill: theme.Sunken, Radius: size.TrackRadius}
	context.Push(name)
	context.Frame(ui.Layout{Padding: ui.Pad(size.TrackPadding)}, track, func() {
		for index, tab := range tabs {
			chosen := index == *selected
			normal := ui.Look{Radius: size.Radius}
			if chosen {
				normal = ui.Look{Fill: theme.Surface, Radius: size.Radius, Shadow: true}
			}
			hovered := normal
			if !chosen {
				hovered.Fill = theme.Surface.WithAlpha(0.5)
			}
			layout := ui.Layout{Padding: ui.PadXY(size.PaddingX, size.PaddingY), Gap: size.Gap, AlignY: ui.Center}
			if context.Clickable(tab.Label, layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
				tabWords(context, tab, chosen, size.TextSize, size.IconSize)
			}) && !chosen {
				*selected, changed = index, true
			}
		}
	})
	context.Pop()
	return changed
}

// PillSize is how big the tabs of a row of PillTabs are.
type PillSize struct {
	TextSize, IconSize float32
	// PaddingX and PaddingY are the room around a tab's words; Gap is between its icon, words
	// and count.
	PaddingX, PaddingY, Gap float32
	// Between is the room between two tabs.
	Between float32
}

// GroupTabs are the tabs of the groups of a list («Todo», «Láminas»…).
var GroupTabs = PillSize{TextSize: 18, IconSize: 20, PaddingX: 18, PaddingY: 10, Gap: 8, Between: 8}

// LargeTabs are the tabs a page is browsed by, big for a quick finger: the stand's.
var LargeTabs = PillSize{TextSize: 20, IconSize: 22, PaddingX: 22, PaddingY: 12, Gap: 10, Between: 8}

// PillTabs is a row of quiet tabs where the chosen one stands in a pill with a fine line: the
// groups of a list («Todo», «Láminas»…). It wraps when they do not fit, and returns true on the
// frame the choice changes.
func PillTabs(context *ui.Context, name string, selected *int, tabs []Tab, size PillSize) bool {
	theme := context.Theme
	changed := false
	context.Push(name)
	context.Box(ui.Layout{Width: ui.Grow(), Gap: size.Between, Wrap: true}, func() {
		for index, tab := range tabs {
			chosen := index == *selected
			// Only the chosen tab has a line, and it is that much wider; the row is as tall as it.
			normal, side := ui.Look{Radius: pillRadius}, size.PaddingX
			if chosen {
				normal = ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: cardBorder, Radius: pillRadius}
				side += cardBorder
			}
			hovered := normal
			if !chosen {
				hovered.Fill = theme.Sunken
			}
			layout := ui.Layout{Padding: ui.PadXY(side, size.PaddingY+cardBorder), Gap: size.Gap, AlignY: ui.Center}
			if context.Clickable(tab.Label, layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
				tabWords(context, tab, chosen, size.TextSize, size.IconSize)
			}) && !chosen {
				*selected, changed = index, true
			}
		}
	})
	context.Pop()
	return changed
}

func tabWords(context *ui.Context, tab Tab, chosen bool, size, iconSize float32) {
	theme := context.Theme
	ink := theme.InkSoft
	if chosen {
		ink = theme.Ink
	}
	if tab.Icon != "" {
		context.IconSized(tab.Icon, iconSize, ink)
	}
	context.Text(tab.Label, ui.TextStyle{Size: size, Weight: draw2d.Medium, Color: ink, LineHeight: ui.LettersLineHeight})
	if tab.HasCount {
		context.Text(fmt.Sprint(tab.Count), ui.TextStyle{Size: tabCountSize, Color: theme.InkQuiet, LineHeight: ui.LettersLineHeight})
	}
}

// TopRounded paints a box's fill with only its upper corners rounded: a picture at the top of a
// card, whose lower edge meets the card's words. Keep one per colour from frame to frame, so
// drawing it allocates nothing.
type TopRounded struct {
	Fill   draw2d.Color
	Radius float32
}

// Paint draws the fill: a rounded rectangle, and a square one over its lower half.
func (shape *TopRounded) Paint(area draw2d.Rectangle, drawing *draw2d.List) {
	drawing.Rectangle(area, shape.Radius, shape.Fill, draw2d.Color{}, 0)
	lower := area
	lower.Y += area.Height / 2
	lower.Height = area.Height / 2
	drawing.Rectangle(lower, 0, shape.Fill, draw2d.Color{}, 0)
}

// Link is words that act, small, bold and underlined, for an action inside a sentence or a
// note («Ver cuáles»). It returns true on the frame it is clicked.
func Link(context *ui.Context, words string) bool {
	theme := context.Theme
	style := ui.TextStyle{Weight: draw2d.Bold}
	return context.Clickable(words, ui.Layout{Direction: ui.Vertical, Gap: 1}, ui.Looks{}, func() {
		context.Text(words, style)
		context.Frame(ui.Layout{Width: ui.Grow(), Height: ui.Fixed(1)}, ui.Look{Fill: theme.Ink}, nil)
	})
}

// QuietLink is soft, underlined words that act, of a size and a line height, under a name
// («Cambiar»). It returns true on the frame it is clicked.
func QuietLink(context *ui.Context, words string, size, lineHeight float32) bool {
	theme := context.Theme
	style := ui.TextStyle{Size: size, Color: theme.InkSoft, LineHeight: lineHeight}
	return context.Clickable(words, ui.Layout{Direction: ui.Stacked, AlignY: ui.End}, ui.Looks{}, func() {
		context.Text(words, style)
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Padding{Bottom: size * underlineDrop}}, func() {
			context.Frame(ui.Layout{Width: ui.Grow(), Height: ui.Fixed(1)}, ui.Look{Fill: theme.InkSoft}, nil)
		})
	})
}

// underlineDrop is how far above the bottom of a line its underline runs, in ems.
const underlineDrop = 0.25
