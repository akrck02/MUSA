package look

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// Measures of the tabs, in pixels.
const (
	segmentPadding   = 4
	segmentItemWidth = 160
	tabPaddingX      = 18
	tabHeight        = 42
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
	track := ui.Look{Fill: theme.Sunken, Radius: theme.PanelRadius}
	context.Push(name)
	context.Frame(ui.Layout{Padding: ui.Pad(segmentPadding), Gap: segmentPadding}, track, func() {
		for index, tab := range tabs {
			chosen := index == *selected
			normal := ui.Look{Radius: theme.ControlRadius}
			if chosen {
				normal = ui.Look{Fill: theme.Surface, Radius: theme.ControlRadius, Shadow: true}
			}
			hovered := normal
			if !chosen {
				hovered.Fill = theme.Surface.WithAlpha(0.5)
			}
			layout := ui.Layout{Width: ui.Fixed(segmentItemWidth), Height: ui.Fixed(tabHeight), Gap: 8, AlignX: ui.Center, AlignY: ui.Center}
			if context.Clickable(tab.Label, layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
				tabWords(context, tab, chosen)
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
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, Wrap: true}, func() {
		for index, tab := range tabs {
			chosen := index == *selected
			normal := ui.Look{Radius: tabHeight / 2}
			if chosen {
				normal = ui.Look{Fill: theme.Surface, Border: theme.ControlLine, BorderWidth: cardBorder, Radius: tabHeight / 2}
			}
			hovered := normal
			if !chosen {
				hovered.Fill = theme.Sunken
			}
			layout := ui.Layout{Height: ui.Fixed(tabHeight), Padding: ui.PadXY(tabPaddingX, 0), Gap: 8, AlignY: ui.Center}
			if context.Clickable(tab.Label, layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
				tabWords(context, tab, chosen)
			}) && !chosen {
				*selected, changed = index, true
			}
		}
	})
	context.Pop()
	return changed
}

func tabWords(context *ui.Context, tab Tab, chosen bool) {
	theme := context.Theme
	ink := theme.InkSoft
	if chosen {
		ink = theme.Ink
	}
	if tab.Icon != "" {
		context.IconSized(tab.Icon, theme.IconSize, ink)
	}
	context.Text(tab.Label, ui.TextStyle{Color: ink, Weight: map[bool]draw2d.Weight{true: draw2d.Bold, false: draw2d.Regular}[chosen]})
	if tab.HasCount {
		context.Text(fmt.Sprint(tab.Count), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
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
	style := ui.TextStyle{Size: theme.SmallTextSize * linkScale, Weight: draw2d.Bold}
	return context.Clickable(words, ui.Layout{Direction: ui.Vertical, Gap: 1}, ui.Looks{}, func() {
		context.Text(words, style)
		context.Frame(ui.Layout{Width: ui.Grow(), Height: ui.Fixed(1)}, ui.Look{Fill: theme.Ink}, nil)
	})
}

// linkScale is how much bigger than small text a link is.
const linkScale = 1.1
