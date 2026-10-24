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

// Measures of the cash dialog, in pixels, as in the author's design.
const (
	cashColumns      = 4
	cashGap          = 12
	cashLabelSize    = 16
	cashLabelAbove   = 24
	cashLabelBelow   = 12
	pieceHeight      = 84
	pieceRadius      = 14
	pieceTextSize    = 24
	pieceCountSide   = 30
	pieceCountSize   = 15
	pieceCountPadX   = 8
	pieceCountTop    = -10
	pieceCountRight  = -8
	sumsAbove        = 24
	sumRadius        = 14
	sumPaddingX      = 18
	sumPaddingY      = 16
	sumLabelSize     = 15
	sumAmountSize    = 30
	changeAmountSize = 40
	sumAmountSpace   = 4
	sumHintSpace     = 6
	changeWeight     = 1.3
	actionsAbove     = 24
	clearSize        = 17
	clearPaddingX    = 8
	clearPaddingY    = 14
	shortTint        = 0.094 // the warning colour's share of the box that says what is missing
	offOpacity       = 0.4
)

// cashPieces are the notes and coins the cash dialog offers, in the order of the author's design;
// «Justo» follows them.
var cashPieces = []money.Money{500, 1000, 2000, 5000, 50, 100, 200}

// cashDialog lets the artist tap the notes and coins handed over, and shows the change.
func (screens *Screens) cashDialog(context *ui.Context) {
	theme := context.Theme
	total := screens.basket.Total()
	screens.dialogHead(context, "Pagar en efectivo")
	look.Space(context, cashLabelAbove)
	context.Text("¿Qué te han dado? Toca cada billete o moneda.", ui.TextStyle{Size: cashLabelSize, Color: theme.InkSoft})
	look.Space(context, cashLabelBelow)
	context.Box(ui.Layout{Width: ui.Grow(), Columns: cashColumns, Gap: cashGap}, func() {
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
	look.Space(context, sumsAbove)
	screens.sums(context, total, change, enough)
	look.Space(context, actionsAbove)
	screens.cashActions(context, change, enough && screens.received.Total() > 0)
}

// sums are what to charge, what was handed over, and the change or what is still missing.
func (screens *Screens) sums(context *ui.Context, total, change money.Money, enough bool) {
	theme := context.Theme
	context.Box(ui.Layout{Width: ui.Grow(), Gap: cashGap}, func() {
		sumBox(context, ui.Share(1), theme.Background, "A cobrar", money.Format(total), sumAmountSize, theme.Ink, "")
		sumBox(context, ui.Share(1), theme.Background, "Te han dado", money.Format(screens.received.Total()), sumAmountSize, theme.Ink, "")
		if enough {
			sumBox(context, ui.Share(changeWeight), theme.Sunken, "Devuelve", money.Format(change), changeAmountSize, theme.Ink, changeWords(change))
			return
		}
		short := theme.Surface.Mix(theme.Warning, shortTint)
		sumBox(context, ui.Share(changeWeight), short, "Falta", money.Format(change), changeAmountSize, theme.Warning,
			"Toca lo que te den hasta llegar a "+money.Format(total))
	})
}

// cashActions are «Borrar», which forgets what was tapped, and «Confirmar», faded until enough was
// handed over.
func (screens *Screens) cashActions(context *ui.Context, change money.Money, ready bool) {
	theme := context.Theme
	context.Box(ui.Layout{Width: ui.Grow(), AlignY: ui.Center}, func() {
		layout := ui.Layout{Padding: ui.PadXY(clearPaddingX, clearPaddingY)}
		if context.Clickable("Borrar", layout, ui.Looks{}, func() {
			context.Text("Borrar", ui.TextStyle{Size: clearSize, Color: theme.InkSoft})
		}) {
			screens.received.Clear()
		}
		context.Spacer()
		if confirmButton(context, "Confirmar", ready) && ready {
			screens.sell(sales.Cash, change)
		}
	})
}

// Measures of a dialog's main buttons, in pixels.
const (
	confirmRadius   = 14
	confirmPaddingX = 40
	confirmPaddingY = 18
	confirmTextSize = 20
)

// confirmButton is a dialog's main action on the accent; faded while it cannot act yet.
func confirmButton(context *ui.Context, words string, ready bool) bool {
	theme := context.Theme
	fill, ink := theme.Primary, theme.OnPrimary
	if !ready {
		fill, ink = theme.Surface.Mix(fill, offOpacity), theme.Surface.Mix(ink, offOpacity)
	}
	normal := ui.Look{Fill: fill, Radius: confirmRadius}
	layout := ui.Layout{Padding: ui.PadXY(confirmPaddingX, confirmPaddingY)}
	return context.Clickable(words, layout, ui.Looks{Normal: normal, Hovered: normal, Pressed: normal}, func() {
		context.Text(words, ui.TextStyle{Size: confirmTextSize, Weight: draw2d.Bold, Color: ink, LineHeight: ui.LettersLineHeight})
	})
}

// pieceButton is a big note or, round, a coin, with how many were tapped in a badge over its corner.
func pieceButton(context *ui.Context, words string, count int, coin bool) bool {
	theme := context.Theme
	radius := float32(pieceRadius)
	if coin {
		radius = roundRadius
	}
	normal := ui.Look{Fill: theme.Background, Border: theme.Line, BorderWidth: 1, Radius: radius}
	if count > 0 {
		normal.Fill, normal.Border = theme.Sunken, theme.Sunken
	}
	pressed := normal
	pressed.Fill = theme.Sunken
	clicked := false
	context.Box(ui.Layout{Direction: ui.Stacked, Width: ui.Grow(), Height: ui.Fixed(pieceHeight)}, func() {
		layout := ui.Layout{Width: ui.Grow(), Height: ui.Grow(), AlignX: ui.Center, AlignY: ui.Center}
		clicked = context.Clickable(words, layout, ui.Looks{Normal: normal, Hovered: pressed, Pressed: pressed}, func() {
			context.Text(words, ui.TextStyle{Size: pieceTextSize, Weight: draw2d.Bold})
		})
		if count > 0 {
			pieceCount(context, count)
		}
	})
	return clicked
}

// pieceCount is the badge over a note's corner: "×2".
func pieceCount(context *ui.Context, count int) {
	theme := context.Theme
	corner := ui.Padding{Top: pieceCountTop, Right: pieceCountRight}
	context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow(), Padding: corner, AlignX: ui.End}, func() {
		layout := ui.Layout{Height: ui.Fixed(pieceCountSide), Padding: ui.PadXY(pieceCountPadX, 0), AlignX: ui.Center, AlignY: ui.Center}
		context.Frame(layout, ui.Look{Fill: theme.Primary, Radius: pieceCountSide / 2}, func() {
			context.Text(fmt.Sprintf("×%d", count), ui.TextStyle{Size: pieceCountSize, Color: theme.OnPrimary, LineHeight: ui.LettersLineHeight})
		})
	})
}

// sumBox is one of the three boxes under the money: a label, an amount and, for the change, a
// line under it.
func sumBox(context *ui.Context, width ui.Sizing, fill draw2d.Color, label, amount string, size float32, ink draw2d.Color, detail string) {
	theme := context.Theme
	layout := ui.Layout{Direction: ui.Vertical, Width: width, Height: ui.Grow(), Padding: ui.PadXY(sumPaddingX, sumPaddingY)}
	context.Frame(layout, ui.Look{Fill: fill, Radius: sumRadius}, func() {
		context.Text(label, ui.TextStyle{Size: sumLabelSize, Color: theme.InkSoft})
		look.Space(context, sumAmountSpace)
		context.Text(amount, ui.TextStyle{Size: size, Weight: draw2d.Bold, Color: ink})
		if detail != "" {
			look.Space(context, sumHintSpace)
			look.Paragraph(context, detail, ui.TextStyle{Size: sumLabelSize, Color: theme.InkSoft})
		}
	})
}

// Measures of the card and paid dialogs, in pixels.
const (
	bigCircleSide    = 120
	cardIconSize     = 56
	paidIconSize     = 60
	cardCircleAbove  = 24
	paidCircleAbove  = 12
	circleBelow      = 16
	amountSize       = 56
	messageSize      = 18
	messageAbove     = 8
	twoAbove         = 28
	twoGap           = 12
	twoMainWeight    = 1.4
	twoRadius        = 14
	twoPadding       = 18
	twoTextSize      = 19
	linesAbove       = 20
	linesRadius      = 14
	linesPaddingX    = 18
	linesPaddingY    = 6
	lineSize         = 17
	linePaddingY     = 10
	undoAbove        = 20
	undoGap          = 10
	positiveTintPart = 0.133 // #1E7D4F22: the positive colour at 34 of 255
)

// cardDialog is paying by card on the artist's own reader, for now: MUSA records the sale once she
// says the reader took it.
// TODO(akrck02): charge through payments.Terminal (SumUp), proposal 4 R7.
func (screens *Screens) cardDialog(context *ui.Context) {
	theme := context.Theme
	total := money.Format(screens.basket.Total())
	screens.dialogHead(context, "Pagar con tarjeta")
	bigResult(context, cardCircleAbove, "credit-card", cardIconSize, theme.Sunken, theme.Primary, total,
		"Cobra "+total+" en tu datáfono y, cuando salga aprobado, confírmalo aquí.")
	look.Space(context, twoAbove)
	context.Box(ui.Layout{Width: ui.Grow(), Gap: twoGap}, func() {
		if twoButton(context, ui.Share(1), "No ha salido", false) {
			screens.step = choosingMethod
		}
		if twoButton(context, ui.Share(twoMainWeight), "Cobrado con tarjeta", true) {
			screens.sell(sales.Card, 0)
		}
	})
}

// paidDialog closes the sale: what was taken, how, the change to hand back, and a way to undo it.
func (screens *Screens) paidDialog(context *ui.Context) {
	theme := context.Theme
	receipt := screens.receipt
	how := "Con tarjeta"
	if receipt.method == sales.Cash {
		how = "En efectivo · devuelve " + money.Format(receipt.change)
	}
	tint := theme.Surface.Mix(theme.Positive, positiveTintPart)
	bigResult(context, paidCircleAbove, "check", paidIconSize, tint, theme.Positive, "Cobrado "+money.Format(receipt.total), how)
	look.Space(context, linesAbove)
	receiptLines(context, receipt.taken)
	look.Space(context, twoAbove)
	if twoButton(context, ui.Grow(), "Seguir vendiendo", true) {
		screens.step = browsing
	}
	look.Space(context, undoAbove)
	context.Box(ui.Layout{Width: ui.Grow(), Gap: undoGap, AlignX: ui.Center, AlignY: ui.Center}, func() {
		context.Text("¿Te has equivocado?", ui.TextStyle{Color: theme.InkSoft})
		if look.Link(context, "Deshacer el cobro") {
			screens.undo()
		}
	})
}

// bigResult is a big round icon, an amount in large type and a line under it, all centred.
func bigResult(context *ui.Context, above float32, icon string, iconSize float32, fill, ink draw2d.Color, amount, message string) {
	theme := context.Theme
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), AlignX: ui.Center}, func() {
		look.Space(context, above)
		circle(context, icon, bigCircleSide, iconSize, fill, ink)
		look.Space(context, circleBelow)
		context.Text(amount, ui.TextStyle{Size: amountSize, Weight: draw2d.Bold})
		look.Space(context, messageAbove)
		look.CentredParagraph(context, message, ui.TextStyle{Size: messageSize, Color: theme.InkSoft})
	})
}

// twoButton is one of a dialog's closing buttons: plain with a line, or main on the accent.
func twoButton(context *ui.Context, width ui.Sizing, words string, main bool) bool {
	theme := context.Theme
	look := ui.Look{Fill: theme.Background, Border: theme.Line, BorderWidth: 1, Radius: twoRadius}
	ink := theme.Ink
	if main {
		look.Fill, look.Border, ink = theme.Primary, theme.Primary, theme.OnPrimary
	}
	layout := ui.Layout{Width: width, Padding: ui.Pad(twoPadding + 1), AlignX: ui.Center}
	return context.Clickable(words, layout, ui.Looks{Normal: look, Hovered: look, Pressed: look}, func() {
		context.Text(words, ui.TextStyle{Size: twoTextSize, Weight: draw2d.Bold, Color: ink, LineHeight: ui.LettersLineHeight})
	})
}

// receiptLines lists what was sold, each line with its amount, on a soft panel.
func receiptLines(context *ui.Context, taken []Taken) {
	theme := context.Theme
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.PadXY(linesPaddingX, linesPaddingY)}
	context.Frame(layout, ui.Look{Fill: theme.Background, Radius: linesRadius}, func() {
		for index, line := range taken {
			if index > 0 {
				context.Frame(ui.Layout{Width: ui.Grow(), Height: ui.Fixed(1)}, ui.Look{Fill: theme.Line}, nil)
			}
			context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(0, linePaddingY), AlignX: ui.SpaceBetween}, func() {
				context.Text(line.Words, ui.TextStyle{Size: lineSize})
				context.Text(money.Format(line.Amount), ui.TextStyle{Size: lineSize, Weight: draw2d.Bold})
			})
		}
	})
}
