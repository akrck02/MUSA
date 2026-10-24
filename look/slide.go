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
	// area is where the pill is inside its row, from the row's corner, so scrolling the row moves
	// the pill with it instead of making it chase the tab.
	area       draw2d.Rectangle
	row        draw2d.Rectangle // where the row was painted last, to measure the chosen tab from
	painted    bool             // the row has been painted once, so row is known
	ready      bool             // area is known: the pill draws itself, and the chosen tab draws none
	look       ui.Look
	blur, drop float32
	shadow     draw2d.Color
}

// slidingPills are the rows' pills by the name of their row.
var slidingPills = map[string]*slidingPill{}

// Paint draws the pill, with its shadow when it has one.
func (pill *slidingPill) Paint(row draw2d.Rectangle, drawing *draw2d.List) {
	pill.row, pill.painted = row, true
	area := pill.area
	area.X += row.X
	area.Y += row.Y
	if !pill.ready {
		return
	}
	if pill.look.Shadow {
		drawing.Shadow(area, pill.look.Radius, pill.blur, pill.drop, pill.shadow)
	}
	drawing.Rectangle(area, pill.look.Radius, pill.look.Fill, pill.look.Border, pill.look.BorderWidth)
}

// slideTo is the pill of a row on its way to the chosen tab, to paint under the row; true when
// it draws itself, false while the row or the tab has not been laid out yet and the chosen tab
// must draw its own pill.
func slideTo(context *ui.Context, name, chosen string, look ui.Look) (*slidingPill, bool) {
	pill, kept := slidingPills[name]
	if !kept {
		pill = &slidingPill{}
		slidingPills[name] = pill
	}
	target, found := context.Area(name, chosen)
	pill.ready = found && pill.painted
	if !pill.ready {
		return pill, false
	}
	theme := context.Theme
	pill.look, pill.blur, pill.drop, pill.shadow = look, theme.ShadowBlur, theme.ShadowDrop, theme.Shadow
	pill.area = draw2d.Rectangle{
		X:      context.Animate(name+" pill x", target.X-pill.row.X, tabMotion),
		Y:      target.Y - pill.row.Y,
		Width:  context.Animate(name+" pill width", target.Width, tabMotion),
		Height: target.Height,
	}
	return pill, true
}
