package stand

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/look"
)

// Measures of the preparing page's header, in pixels.
const (
	preparingTitleSize = 26
	preparingGap       = 14
	preparingBelow     = 12
)

// preparing is the stand turned round: the catalogue's stock, tapped into a case to take to an
// event instead of sold from it.
type preparing struct {
	event events.Event
	done  func() // goes back to the event's page, whether anything was packed or not
}

// Prepare opens the catalogue's stock on the stand to choose what goes to event; done is called when
// the artist leaves, after «Llevar al evento» or without it.
func (screens *Screens) Prepare(event events.Event, done func()) {
	screens.Choose(WholeStock{Sales: screens.sales, Products: screens.products, Now: screens.now})
	screens.preparing = &preparing{event: event, done: done}
}

// Preparing tells whether the stand is choosing what to take to an event.
func (screens *Screens) Preparing() bool { return screens.preparing != nil }

// leavePreparing closes the preparing page, moving nothing more.
func (screens *Screens) leavePreparing() {
	done := screens.preparing.done
	screens.Choose(nil)
	done()
}

// preparingHeader is «Preparar» and the event's name, with the way back and what to do.
func (screens *Screens) preparingHeader(context *ui.Context) {
	theme := context.Theme
	name := screens.preparing.event.Name // leaving clears preparing in the middle of this frame
	context.Box(ui.Layout{Width: ui.Grow(), Gap: preparingGap, AlignY: ui.Center, Padding: ui.Padding{Bottom: preparingBelow}}, func() {
		if look.ToolButton(context, "arrow-left", "Volver") {
			screens.leavePreparing()
		}
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow()}, func() {
			context.Text("Preparar "+name, ui.TextStyle{Size: preparingTitleSize, Weight: draw2d.Bold})
			look.Paragraph(context, "Toca lo que te llevas, como en el puesto. Lo que no vendas vuelve al catálogo al cerrar.",
				ui.TextStyle{Color: theme.InkSoft})
		})
	})
}

// takeAll moves what is in the case to the event, and goes back to its page. A refusal stops it
// there and says why; what moved before stays moved.
func (screens *Screens) takeAll() {
	moved := 0
	for _, line := range screens.basket.Lines(func() string { return "" }) {
		refused, err := screens.events.MoveStock(screens.preparing.event.ID, line.VariantID, line.Quantity)
		if err != nil {
			screens.tell(true, "No se pudo llevar: "+err.Error())
			return
		}
		if len(refused) > 0 {
			screens.tell(true, events.RefusalWords(refused[0]))
			return
		}
		moved += line.Quantity
	}
	screens.leavePreparing()
	screens.tell(false, fmt.Sprintf("%s llevados al evento", piecesWords(moved)))
}

// inBasketWords is what a chosen tile says: in the case when preparing, in the basket when selling.
func (screens *Screens) inBasketWords() string {
	if screens.preparing != nil {
		return "en la maleta"
	}
	return "en la cesta"
}
