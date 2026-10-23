package look

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// Measures of the tabs, in pixels, as in the author's design.
const (
	segmentTrackPadding = 5
	segmentTrackRadius  = 16
	segmentRadius       = 12
	segmentPaddingX     = 26
	segmentPaddingY     = 12
	segmentGap          = 10
	segmentTextSize     = 19
	pillTabPaddingX     = 18
	pillTabPaddingY     = 10
	pillTabGap          = 8
	pillTabTextSize     = 18
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

// Segmented is a track with a raised, lit choice: the halves of a page («Físico», «Digital»). It
// returns true on the frame the choice changes.
func Segmented(context *ui.Context, name string, selected *int, tabs []Tab) bool {
	theme := context.Theme
	changed := false
	track := ui.Look{Fill: theme.Sunken, Radius: segmentTrackRadius}
	context.Push(name)
	context.Frame(ui.Layout{Padding: ui.Pad(segmentTrackPadding)}, track, func() {
		for index, tab := range tabs {
			chosen := index == *selected
			normal := ui.Look{Radius: segmentRadius}
			if chosen {
				normal = ui.Look{Fill: theme.Surface, Radius: segmentRadius, Shadow: true}
			}
			hovered := normal
			if !chosen {
				hovered.Fill = theme.Surface.WithAlpha(0.5)
			}
			layout := ui.Layout{Padding: ui.PadXY(segmentPaddingX, segmentPaddingY), Gap: segmentGap, AlignY: ui.Center}
			if context.Clickable(tab.Label, layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
				tabWords(context, tab, chosen, segmentTextSize)
			}) && !chosen {
				*selected, changed = index, true
			}
		}
	})
	context.Pop()
	return changed
}

// PillTabs is a row of quiet tabs where the chosen one stands in a pill with a fine line: the
// groups of a list («Todo», «Láminas»…). It wraps when they do not fit, and returns true on the
// frame the choice changes.
func PillTabs(context *ui.Context, name string, selected *int, tabs []Tab) bool {
	theme := context.Theme
	changed := false
	context.Push(name)
	context.Box(ui.Layout{Width: ui.Grow(), Gap: pillTabGap, Wrap: true}, func() {
		for index, tab := range tabs {
			chosen := index == *selected
			// Only the chosen tab has a line, and it is that much wider; the row is as tall as it.
			normal, side := ui.Look{Radius: pillRadius}, float32(pillTabPaddingX)
			if chosen {
				normal = ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: cardBorder, Radius: pillRadius}
				side += cardBorder
			}
			hovered := normal
			if !chosen {
				hovered.Fill = theme.Sunken
			}
			layout := ui.Layout{Padding: ui.PadXY(side, pillTabPaddingY+cardBorder), Gap: pillTabGap, AlignY: ui.Center}
			if context.Clickable(tab.Label, layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
				tabWords(context, tab, chosen, pillTabTextSize)
			}) && !chosen {
				*selected, changed = index, true
			}
		}
	})
	context.Pop()
	return changed
}

func tabWords(context *ui.Context, tab Tab, chosen bool, size float32) {
	theme := context.Theme
	ink := theme.InkSoft
	if chosen {
		ink = theme.Ink
	}
	if tab.Icon != "" {
		context.IconSized(tab.Icon, theme.IconSize, ink)
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
