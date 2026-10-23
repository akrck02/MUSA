package stand

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/cash"
	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
	"github.com/Itros97/MUSA/sales"
)

// dialogWidth is how wide the payment dialogs are, in pixels.
const dialogWidth = 760

// drawCheckout draws the payment dialog of the current step, if any.
func (screens *Screens) drawCheckout(context *ui.Context) {
	steps := map[step]func(*ui.Context){choosingMethod: screens.methodDialog, payingCash: screens.cashDialog,
		payingCard: screens.cardDialog, paid: screens.paidDialog}
	draw, found := steps[screens.step]
	if !found {
		return
	}
	options := ui.LayerOptions{AlignX: ui.Center, AlignY: ui.Center, Width: ui.Fixed(dialogWidth)}
	if context.Modal("checkout", options, func() {
		look.Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(24), Gap: 16}, func() { draw(context) })
	}) {
		screens.step = browsing
	}
}

// methodDialog asks how the customer pays: two big cards, cash or card.
func (screens *Screens) methodDialog(context *ui.Context) {
	theme := context.Theme
	screens.dialogHeader(context, "Cobrar "+money.Format(screens.basket.Total()), false)
	context.Text(piecesWords(screens.basket.Count())+": "+screens.basket.Words(), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
	context.Grid(2, func() {
		if screens.methodCard(context, "coins", "Pagar en efectivo") {
			screens.received.Clear()
			screens.step = payingCash
		}
		if screens.methodCard(context, "credit-card", "Pagar con tarjeta") {
			screens.step = payingCard
		}
	})
}

// dialogHeader is a dialog's title, with ‹ to go back to the choice of method when back, and ×.
func (screens *Screens) dialogHeader(context *ui.Context, title string, back bool) {
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 10, AlignY: ui.Center}, func() {
		if back && roundButton(context, "chevron-left", "Atrás") {
			screens.step = choosingMethod
		}
		dialogTitle(context, title)
		context.Spacer()
		if roundButton(context, "x", "Cerrar") {
			screens.step = browsing
		}
	})
}

// roundButton is a small round icon button on a soft fill.
func roundButton(context *ui.Context, icon, label string) bool {
	theme := context.Theme
	round := ui.Look{Fill: theme.Sunken, Radius: roundSide / 2}
	layout := ui.Layout{Width: ui.Fixed(roundSide), Height: ui.Fixed(roundSide), AlignX: ui.Center, AlignY: ui.Center}
	return context.Clickable(label, layout, ui.Looks{Normal: round, Hovered: round, Pressed: round}, func() {
		context.IconSized(icon, theme.IconSize, theme.Ink)
	})
}

func (screens *Screens) methodCard(context *ui.Context, icon, title string) bool {
	return look.ClickableCard(context, title, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(24), Gap: 16, AlignX: ui.Center}, func() {
		look.Badge(context, icon, 72)
		context.Text(title, ui.TextStyle{Weight: draw2d.Bold})
	})
}

// cashPieces are the notes and coins the cash dialog offers, in the order of the author's design;
// «Justo» follows them.
var cashPieces = []money.Money{500, 1000, 2000, 5000, 50, 100, 200}

// cashDialog lets the artist tap the notes and coins handed over, and shows the change.
func (screens *Screens) cashDialog(context *ui.Context) {
	theme := context.Theme
	total := screens.basket.Total()
	screens.dialogHeader(context, "Pagar en efectivo", true)
	context.Text("¿Qué te han dado? Toca cada billete o moneda.", ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
	context.Grid(cashColumns, func() {
		for _, piece := range cashPieces {
			context.Push(money.Format(piece))
			if pieceButton(context, money.Format(piece), screens.received.Count(piece), !cash.IsNote(piece)) {
				screens.received.Add(piece)
			}
			context.Pop()
		}
		if pieceButton(context, "Justo", 0, false) {
			screens.received.Clear()
			screens.received.Add(total)
		}
	})
	change, enough := cash.Change(total, screens.received)
	context.Grid(3, func() {
		cashFigure(context, "A cobrar", money.Format(total), "", false)
		cashFigure(context, "Te han dado", money.Format(screens.received.Total()), "", false)
		if enough {
			cashFigure(context, "Devuelve", money.Format(change), changeWords(change), true)
		} else {
			cashFigure(context, "Falta", money.Format(change), "", true)
		}
	})
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.Center}, func() {
		if context.FlatButton("Borrar") {
			screens.received.Clear()
		}
		context.Spacer()
		if enough && screens.received.Total() > 0 && context.PrimaryButton("Confirmar") {
			screens.sell(sales.Cash, change)
		}
	})
}

// pieceButton is a big note (square-ish) or coin (round) with how many were tapped.
func pieceButton(context *ui.Context, words string, count int, coin bool) bool {
	theme := context.Theme
	radius := theme.ControlRadius
	if coin {
		radius = pieceHeight / 2
	}
	normal := ui.Look{Fill: theme.Background, Border: theme.Line, BorderWidth: 1, Radius: radius}
	if count > 0 {
		normal.Fill = theme.Sunken
	}
	pressed := normal
	pressed.Fill = theme.Sunken
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Fixed(pieceHeight), AlignX: ui.Center, AlignY: ui.Center}
	return context.Clickable(words, layout, ui.Looks{Normal: normal, Hovered: pressed, Pressed: pressed}, func() {
		context.Box(ui.Layout{Height: ui.Fixed(countRowHeight)}, func() {
			if count > 0 {
				look.Pill(context, fmt.Sprintf("×%d", count), look.Tone{Fill: theme.Primary, Ink: theme.OnPrimary})
			}
		})
		context.Text(words, ui.TextStyle{Size: theme.TextSize * pieceTextScale, Weight: draw2d.Bold})
	})
}

// cashFigure is one of the three boxes under the money: a label, an amount and, for the change,
// the notes and coins to give.
func cashFigure(context *ui.Context, label, amount, detail string, stands bool) {
	theme := context.Theme
	fill := theme.Background
	if stands {
		fill = theme.Sunken
	}
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Fixed(figureHeight), Padding: ui.Pad(12), Gap: 6}
	context.Frame(layout, ui.Look{Fill: fill, Radius: theme.ControlRadius}, func() {
		context.Text(label, ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
		size := theme.TextSize * pieceTextScale
		if stands {
			size *= 1.4
		}
		context.Text(amount, ui.TextStyle{Size: size, Weight: draw2d.Bold})
		if detail != "" {
			look.Paragraph(context, detail, ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
		}
	})
}

// Measures of the cash dialog, in pixels.
const (
	cashColumns    = 4
	pieceHeight    = 84
	countRowHeight = 26
	pieceTextScale = 1.2
	figureHeight   = 160
	roundSide      = 40
)

// cardDialog is paying by card on the artist's own reader, for now: MUSA records the sale once she
// says the reader took it.
// TODO(akrck02): charge through payments.Terminal (SumUp), proposal 4 R7.
func (screens *Screens) cardDialog(context *ui.Context) {
	theme := context.Theme
	dialogTitle(context, "Pagar con tarjeta · "+money.Format(screens.basket.Total()))
	context.Text("Cobra en el lector y confirma cuando haya pasado.", ui.TextStyle{Color: theme.InkSoft})
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.Center}, func() {
		context.Spacer()
		if context.FlatButton("Atrás") {
			screens.step = choosingMethod
		}
		if context.PrimaryButton("Ha pasado") {
			screens.sell(sales.Card, 0)
		}
	})
}

// paidDialog closes the sale: what was taken, and the change to hand back.
func (screens *Screens) paidDialog(context *ui.Context) {
	theme := context.Theme
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 12, AlignX: ui.Center}, func() {
		look.Badge(context, "check", 72)
		dialogTitle(context, "Cobrado "+money.Format(screens.receipt.total))
		if screens.receipt.method == sales.Cash {
			context.Text("Devuelve "+money.Format(screens.receipt.change), ui.TextStyle{Size: theme.TextSize * 1.3, Weight: draw2d.Bold})
		}
		if context.PrimaryButton("Siguiente cliente") {
			screens.step = browsing
		}
	})
}

// sell records the basket, paid with method; on success the stand shows what is left.
func (screens *Screens) sell(method sales.Method, change money.Money) {
	total := screens.basket.Total()
	refusal, err := screens.source.Sell(screens.basket.Lines(screens.newID), method)
	switch {
	case err != nil:
		screens.step = browsing
		screens.tell(true, err.Error())
	case refusal != "":
		screens.step = browsing
		screens.tell(true, refusal)
	default:
		screens.receipt = receipt{total: total, change: change, method: method}
		screens.basket.Clear()
		screens.step = paid
		screens.Reload()
	}
}

func dialogTitle(context *ui.Context, title string) {
	context.Text(title, ui.TextStyle{Size: context.Theme.TextSize * 1.3, Weight: draw2d.Bold})
}
