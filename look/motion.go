package look

import (
	"time"

	ui "github.com/akrck02/go-ui"
)

// How long things take to move, as Bubble moves them: quick for what answers a touch, a little
// longer for what arrives, and quicker to leave than to come.
const (
	PageMotion   = 240 * time.Millisecond
	DialogMotion = 220 * time.Millisecond
	LeaveMotion  = 160 * time.Millisecond
)

// How far things travel as they come, in pixels: a page rises a little, a dialog a little more.
const (
	PageRise   = 10
	DialogRise = 18
)

// Arriving is a look on its way in: shown goes from 0 (not yet) to 1 (there), as Context.Appear
// gives it; until then the box is see-through and lower by part of rise.
func Arriving(look ui.Look, shown, rise float32) ui.Look {
	look.Fade = 1 - shown
	look.ShiftY = rise * (1 - shown)
	return look
}

// Leaving is how a dialog or a floating bar goes when it is no longer shown: it fades as it sinks.
func Leaving(rise float32) ui.Leave {
	return ui.Leave{Duration: LeaveMotion, Y: rise}
}
