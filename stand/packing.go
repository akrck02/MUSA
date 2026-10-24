package stand

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/look"
)

// Measures of the packing page's header, in pixels.
const (
	packingTitleSize = 26
	packingGap       = 14
	packingBelow     = 12
)

// packing is the stand turned round: the catalogue's stock, tapped into a case to take to an
// event instead of sold from it.
type packing struct {
	event events.Event
	done  func() // goes back to the event's page, whether anything was packed or not
}

// Pack opens the catalogue's stock on the stand to choose what goes to event; done is called when
// the artist leaves, after «Llevar al evento» or without it.
func (screens *Screens) Pack(event events.Event, done func()) {
	screens.Choose(WholeStock{Sales: screens.sales, Products: screens.products, Now: screens.now})
	screens.packing = &packing{event: event, done: done}
}

// Packing tells whether the stand is choosing what to take to an event.
func (screens *Screens) Packing() bool { return screens.packing != nil }

// leavePacking closes the packing page, moving nothing more.
func (screens *Screens) leavePacking() {
	done := screens.packing.done
	screens.Choose(nil)
	done()
}

// packingHeader is «Preparar» and the event's name, with the way back and what to do.
func (screens *Screens) packingHeader(context *ui.Context) {
	theme := context.Theme
	name := screens.packing.event.Name // leaving clears packing in the middle of this frame
	context.Box(ui.Layout{Width: ui.Grow(), Gap: packingGap, AlignY: ui.Center, Padding: ui.Padding{Bottom: packingBelow}}, func() {
		if look.ToolButton(context, "arrow-left", "Volver") {
			screens.leavePacking()
		}
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow()}, func() {
			context.Text("Preparar "+name, ui.TextStyle{Size: packingTitleSize, Weight: draw2d.Bold})
			look.Paragraph(context, "Toca lo que te llevas, como en el puesto. Lo que no vendas vuelve al catálogo al cerrar.",
				ui.TextStyle{Color: theme.InkSoft})
		})
	})
}

// packAll moves what is in the case to the event, and goes back to its page. A refusal stops it
// there and says why; what moved before stays moved.
func (screens *Screens) packAll() {
	moved := 0
	for _, line := range screens.basket.Lines(func() string { return "" }) {
		refused, err := screens.events.MoveStock(screens.packing.event.ID, line.VariantID, line.Quantity)
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
	screens.leavePacking()
	screens.tell(false, fmt.Sprintf("%s llevados al evento", piecesWords(moved)))
}

// inBasketWords is what a chosen tile says: in the case when packing, in the basket when selling.
func (screens *Screens) inBasketWords() string {
	if screens.packing != nil {
		return "en la maleta"
	}
	return "en la cesta"
}
