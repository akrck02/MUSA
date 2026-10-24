package look

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// tabMotion is how long the chosen tab's pill takes to slide to another tab.
const tabMotion = DialogMotion

// slidingPill is the chosen tab's pill, painted under a row of tabs where it is on its way: from
// the tab chosen before to the one chosen now. Rows keep theirs from frame to frame, by name.
type slidingPill struct {
	area       draw2d.Rectangle
	look       ui.Look
	blur, drop float32
	shadow     draw2d.Color
}

// slidingPills are the rows' pills by the name of their row.
var slidingPills = map[string]*slidingPill{}

// Paint draws the pill, with its shadow when it has one.
func (pill *slidingPill) Paint(_ draw2d.Rectangle, drawing *draw2d.List) {
	if pill.look.Shadow {
		drawing.Shadow(pill.area, pill.look.Radius, pill.blur, pill.drop, pill.shadow)
	}
	drawing.Rectangle(pill.area, pill.look.Radius, pill.look.Fill, pill.look.Border, pill.look.BorderWidth)
}

// slideTo is the pill of a row on its way to the chosen tab, whose area is known from the last
// frame; false the first time the row is drawn, when the chosen tab must draw its own pill.
func slideTo(context *ui.Context, name, chosen string, look ui.Look) (*slidingPill, bool) {
	target, found := context.Area(name, chosen)
	if !found {
		return nil, false
	}
	pill, kept := slidingPills[name]
	if !kept {
		pill = &slidingPill{}
		slidingPills[name] = pill
	}
	theme := context.Theme
	pill.look, pill.blur, pill.drop, pill.shadow = look, theme.ShadowBlur, theme.ShadowDrop, theme.Shadow
	pill.area = draw2d.Rectangle{
		X:      context.Animate(name+" pill x", target.X, tabMotion),
		Y:      context.Animate(name+" pill y", target.Y, tabMotion),
		Width:  context.Animate(name+" pill width", target.Width, tabMotion),
		Height: target.Height,
	}
	return pill, true
}
