package stand

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/cash"
	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
	"github.com/Itros97/MUSA/sales"
)

// dialogWidth is how wide the payment dialogs are, in pixels.
const dialogWidth = 560

// coinColumns is how many notes and coins a row of the cash dialog holds.
const coinColumns = 5

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
	dialogTitle(context, "Cobrar "+money.Format(screens.basket.Total()))
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

func (screens *Screens) methodCard(context *ui.Context, icon, title string) bool {
	return look.ClickableCard(context, title, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(20), Gap: 12, AlignX: ui.Center}, func() {
		look.Badge(context, icon, 72)
		context.Text(title, ui.TextStyle{Weight: draw2d.Bold})
	})
}

// cashDialog lets the artist tap the notes and coins handed over, and shows the change.
func (screens *Screens) cashDialog(context *ui.Context) {
	theme := context.Theme
	total := screens.basket.Total()
	dialogTitle(context, "Pagar en efectivo · "+money.Format(total))
	context.Text("¿Qué te ha dado?", ui.TextStyle{Color: theme.InkSoft})
	context.Grid(coinColumns, func() {
		for _, piece := range cash.Denominations {
			if context.Button(money.Format(piece)) {
				screens.received.Add(piece)
			}
		}
	})
	change, enough := cash.Change(total, screens.received)
	look.Figures(context, 2, func() {
		look.Figure(context, "Te ha dado", money.Format(screens.received.Total()), false)
		if enough {
			look.Figure(context, "Cambio", money.Format(change), false)
		} else {
			look.Figure(context, "Falta", money.Format(change), true)
		}
	})
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.Center}, func() {
		if context.IconButton("undo-2", "Deshacer") {
			screens.received.Undo()
		}
		context.Spacer()
		if context.FlatButton("Atrás") {
			screens.step = choosingMethod
		}
		if enough && context.PrimaryButton("Confirmar") {
			screens.sell(sales.Cash, change)
		}
	})
}

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
