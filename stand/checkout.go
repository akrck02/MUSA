package stand

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
	"github.com/Itros97/MUSA/sales"
)

// Widths of the payment dialogs, in pixels, as in the author's design.
const (
	methodWidth = 640
	cashWidth   = 760
	cardWidth   = 600
	paidWidth   = 560
)

// Measures every payment dialog shares, in pixels.
const (
	dialogPadding   = 32
	dialogRadius    = 16
	titleSize       = 28
	titleHeight     = 1.2 // a heading's line, as the design's h2
	roundSide       = 48
	roundInset      = 20
	headGap         = 14
	summarySize     = 18
	summarySpace    = 6
	summarySpaceOut = 24
)

// drawCheckout draws the payment dialog of the current step, if any.
func (screens *Screens) drawCheckout(context *ui.Context) {
	steps := map[step]func(*ui.Context){choosingMethod: screens.methodDialog, payingCash: screens.cashDialog,
		payingCard: screens.cardDialog, paid: screens.paidDialog}
	widths := map[step]float32{choosingMethod: methodWidth, payingCash: cashWidth, payingCard: cardWidth, paid: paidWidth}
	draw, found := steps[screens.step]
	if !found {
		return
	}
	theme := context.Theme
	options := ui.LayerOptions{AlignX: ui.Center, AlignY: ui.Center, Width: ui.Fixed(widths[screens.step])}
	if context.Modal("checkout", options, func() {
		context.Frame(ui.Layout{Direction: ui.Stacked, Width: ui.Grow()}, ui.Look{Fill: theme.Surface, Radius: dialogRadius, Shadow: true}, func() {
			context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(dialogPadding)}, func() { draw(context) })
			if screens.step != paid {
				screens.cornerClose(context)
			}
		})
	}) {
		screens.step = browsing
	}
}

// methodDialog asks how the customer pays: two big cards, cash or card, under the total and what
// the basket holds.
func (screens *Screens) methodDialog(context *ui.Context) {
	theme := context.Theme
	dialogTitle(context, "Cobrar "+money.Format(screens.basket.Total()))
	look.Space(context, summarySpace)
	context.Text(piecesWords(screens.basket.Count())+": "+screens.basket.Words(), ui.TextStyle{Size: summarySize, Color: theme.InkSoft})
	look.Space(context, summarySpaceOut)
	context.Box(ui.Layout{Width: ui.Grow(), Columns: 2, Gap: methodGap}, func() {
		if screens.methodCard(context, "coins", "Pagar en efectivo") {
			screens.received.Clear()
			screens.step = payingCash
		}
		if screens.methodCard(context, "credit-card", "Pagar con tarjeta") {
			screens.step = payingCard
		}
	})
}

// Measures of the choice of method, in pixels.
const (
	methodGap       = 16
	methodRadius    = 18
	methodPaddingX  = 20
	methodPaddingY  = 32
	methodBadge     = 96
	methodIconSize  = 48
	methodTitleSize = 21
)

func (screens *Screens) methodCard(context *ui.Context, icon, title string) bool {
	theme := context.Theme
	normal := ui.Look{Fill: theme.Background, Border: theme.Line, BorderWidth: 1, Radius: methodRadius}
	hovered := normal
	hovered.Border = theme.ControlLine
	padding := ui.PadXY(methodPaddingX+1, methodPaddingY+1)
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: padding, Gap: methodGap, AlignX: ui.Center}
	return context.Clickable(title, layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
		circle(context, icon, methodBadge, methodIconSize, theme.Sunken, theme.Primary)
		context.Text(title, ui.TextStyle{Size: methodTitleSize, Weight: draw2d.Bold})
	})
}

// circle is an icon in the middle of a round patch.
func circle(context *ui.Context, icon string, side, iconSize float32, fill, ink draw2d.Color) {
	size := ui.Fixed(side)
	context.Frame(ui.Layout{Width: size, Height: size, AlignX: ui.Center, AlignY: ui.Center}, ui.Look{Fill: fill, Radius: side / 2}, func() {
		context.IconSized(icon, iconSize, ink)
	})
}

// dialogHead is a dialog's ‹ and its title: ‹ goes back to the choice of method.
func (screens *Screens) dialogHead(context *ui.Context, title string) {
	context.Box(ui.Layout{Width: ui.Grow(), Gap: headGap, AlignY: ui.Center}, func() {
		if roundButton(context, "chevron-left", "Atrás") {
			screens.step = choosingMethod
		}
		context.Box(ui.Layout{Width: ui.Grow()}, func() { dialogTitle(context, title) })
	})
}

// cornerClose is a dialog's ×, in its top right corner, 20 pixels in, over whatever line is there.
func (screens *Screens) cornerClose(context *ui.Context) {
	context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Top: roundInset, Right: roundInset}, AlignX: ui.End}, func() {
		if roundButton(context, "x", "Cerrar") {
			screens.step = browsing
		}
	})
}

// roundButton is a round icon button on a soft fill.
func roundButton(context *ui.Context, icon, label string) bool {
	theme := context.Theme
	round := ui.Look{Fill: theme.Sunken, Radius: roundSide / 2}
	layout := ui.Layout{Width: ui.Fixed(roundSide), Height: ui.Fixed(roundSide), AlignX: ui.Center, AlignY: ui.Center}
	clicked := context.Clickable(label, layout, ui.Looks{Normal: round, Hovered: round, Pressed: round}, func() {
		context.IconSized(icon, roundIconSize, theme.Ink)
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: label})
	return clicked
}

// roundIconSize is the icon of a round button, in pixels.
const roundIconSize = 22

func dialogTitle(context *ui.Context, title string) {
	context.Text(title, ui.TextStyle{Size: titleSize, Weight: draw2d.Bold, LineHeight: titleHeight})
}

// sell records the basket, paid with method; on success the stand shows what is left.
func (screens *Screens) sell(method sales.Method, change money.Money) {
	total, taken := screens.basket.Total(), screens.basket.Taken()
	saleID, refusal, err := screens.source.Sell(screens.basket.Lines(screens.newID), method)
	switch {
	case err != nil:
		screens.step = browsing
		screens.tell(true, err.Error())
	case refusal != "":
		screens.step = browsing
		screens.tell(true, refusal)
	default:
		screens.receipt = receipt{saleID: saleID, total: total, change: change, method: method, taken: taken}
		screens.basket.Clear()
		screens.step = paid
		screens.Reload()
	}
}

// undo takes back the last sale, after the artist says it was a mistake.
func (screens *Screens) undo() {
	refusal, err := screens.source.Undo(screens.receipt.saleID)
	screens.step = browsing
	screens.Reload()
	switch {
	case err != nil:
		screens.tell(true, err.Error())
	case refusal != "":
		screens.tell(true, refusal)
	default:
		screens.tell(false, "Cobro deshecho: las piezas vuelven al puesto")
	}
}
