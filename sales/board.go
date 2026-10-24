package sales

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
)

// boardColumns are the statuses the board shows, one column each: a delivered order leaves it.
var boardColumns = []Status{ToPrepare, Sent}

// columnIcons is the picture of each column.
var columnIcons = map[Status]string{ToPrepare: "package", Sent: "truck"}

// actionWords is what the button of an order says: the status it moves the order to.
var actionWords = map[Status]string{ToPrepare: "Preparado", Sent: "Entregado"}

// boardWords is the subtitle of the orders page: how many are waiting at each step.
func boardWords(onTheWay []Sale) string {
	counts := map[Status]int{}
	for _, sale := range onTheWay {
		counts[sale.Status]++
	}
	return fmt.Sprintf("%d por preparar · %d en camino", counts[ToPrepare], counts[Sent])
}

// Measures of the orders page, in pixels, as in the author's design. A border takes no room in
// go-ui, so paddings beside one are a pixel larger than the design's CSS.
const (
	viewsTop         = 18
	viewsBottom      = 16
	boardGap         = 16
	panelRadius      = 18
	panelPadding     = 18 + 1
	panelTitleSize   = 17
	panelTitleBottom = 12
	panelIconCircle  = 36
	panelIconSize    = 18
	panelTitleGap    = 10
	orderPaddingY    = 14
	orderGap         = 12
	nameSize         = 16
	detailSize       = 14
	rowLineHeight    = 1.25
	thumbSide        = 40
	thumbRadius      = 10
	thumbBorder      = 2
	thumbOverlap     = -10
	thumbsShown      = 2
	actionRadius     = 12
	actionPaddingX   = 14 + 1
	actionPaddingY   = 10 + 1
	actionSize       = 15
	noteTop          = 120 + 16 // under the page's own bottom padding
	noteSize         = 15
	sourceSize       = 12
	sourcePaddingX   = 8
	sourcePaddingY   = 2
	sourceSpace      = 6
	deliveredNote    = "Al marcar «Entregado», el pedido pasa al historial. Las ventas locales (ferias y en mano) van directas al historial."
)

// A shop's pill is in its own orange: a soft fill and a deep ink.
var (
	marketplaceFill = draw2d.MustHex("#F1641E22")
	marketplaceInk  = draw2d.MustHex("#B4470F")
)

// viewTabs are the tabs of the page's two views, «Por enviar» and «Historial de ventas».
var viewTabs = look.PillSize{TextSize: 18, PaddingX: 18, PaddingY: 10, Between: 6}

// drawBoard is the orders still to deliver, in one panel per status.
func (screens *Screens) drawBoard(context *ui.Context, onTheWay []Sale) {
	context.Box(ui.Layout{Width: ui.Grow(), Gap: boardGap, Columns: look.Columns(context, len(boardColumns))}, func() {
		for _, status := range boardColumns {
			screens.drawColumn(context, status, onTheWay)
		}
	})
	look.Space(context, noteTop)
	look.Paragraph(context, deliveredNote, ui.TextStyle{Size: noteSize, Color: context.Theme.InkSoft})
}

func (screens *Screens) drawColumn(context *ui.Context, status Status, onTheWay []Sale) {
	theme := context.Theme
	var orders []Sale
	for _, sale := range onTheWay {
		if sale.Status == status {
			orders = append(orders, sale)
		}
	}
	panel := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: panelRadius}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(panelPadding)}, panel, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Gap: panelTitleGap, AlignY: ui.Center, Padding: ui.Padding{Bottom: panelTitleBottom}}, func() {
			iconCircle(context, columnIcons[status])
			context.Box(ui.Layout{Width: ui.Grow()}, func() {
				context.Text(StatusWords(status), ui.TextStyle{Size: panelTitleSize, Weight: draw2d.Bold})
			})
			context.Text(fmt.Sprint(len(orders)), ui.TextStyle{Size: panelTitleSize, Weight: draw2d.Medium, Color: theme.InkQuiet})
		})
		for _, sale := range orders {
			context.Separator()
			context.Push(sale.ID)
			screens.boardRow(context, sale)
			context.Pop()
		}
	})
}

// iconCircle is an icon in the accent on a sunken circle.
func iconCircle(context *ui.Context, icon string) {
	theme := context.Theme
	side := ui.Fixed(panelIconCircle)
	context.Frame(ui.Layout{Width: side, Height: side, AlignX: ui.Center, AlignY: ui.Center},
		ui.Look{Fill: theme.Sunken, Radius: panelIconCircle / 2}, func() {
			context.IconSized(icon, panelIconSize, theme.Primary)
		})
}

// boardRow is one order on the board: its products' pictures, who and what on the left, the
// button that moves it on the right.
func (screens *Screens) boardRow(context *ui.Context, sale Sale) {
	context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(0, orderPaddingY), Gap: orderGap, AlignY: ui.Center}, func() {
		screens.thumbs(context, sale)
		nameAndDetail(context, sale, sourceWords(sale), whoOrdered(sale), fmt.Sprintf("%s · %s", piecesWords(sale), money.Format(sale.Total())))
		if actionButton(context, actionWords[sale.Status]) {
			problems, err := screens.store.Advance(sale.ID)
			screens.afterAdvance(sale.Status, problems, err)
		}
	})
}

// nameAndDetail is who a sale is for with its source's pill, over a soft line about it.
func nameAndDetail(context *ui.Context, sale Sale, source, name, detail string) {
	theme := context.Theme
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow()}, ui.Look{Clip: true}, func() {
		context.Box(ui.Layout{Gap: sourceSpace, AlignY: ui.Center}, func() {
			context.Text(name, ui.TextStyle{Size: nameSize, Weight: draw2d.Bold})
			sourcePill(context, source, sale.Channel)
		})
		context.Text(detail, ui.TextStyle{Size: detailSize, Color: theme.InkSoft})
	})
}

// actionButton is an outlined button on the page's colour, for a row's one action.
func actionButton(context *ui.Context, words string) bool {
	theme := context.Theme
	normal := ui.Look{Fill: theme.Background, Border: theme.Line, BorderWidth: 1, Radius: actionRadius}
	hovered := normal
	hovered.Fill = theme.Sunken
	layout := ui.Layout{Padding: ui.PadXY(actionPaddingX, actionPaddingY)}
	clicked := context.Clickable(words, layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
		context.Text(words, ui.TextStyle{Size: actionSize, Weight: draw2d.Medium, LineHeight: ui.LettersLineHeight})
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: words})
	return clicked
}

// thumbs are the pictures of a sale's first products, each over the edge of the one before.
func (screens *Screens) thumbs(context *ui.Context, sale Sale) {
	theme := context.Theme
	context.Box(ui.Layout{Gap: thumbOverlap}, func() {
		for index, line := range sale.Lines {
			if index == thumbsShown {
				break
			}
			side := ui.Fixed(thumbSide)
			fill := catalogue.ListFill(screens.positions[line.ProductName])
			context.Frame(ui.Layout{Width: side, Height: side},
				ui.Look{Fill: fill, Border: theme.Surface, BorderWidth: thumbBorder, Radius: thumbRadius}, nil)
		}
	})
}

// sourcePill is where a sale came from, small: a shop's in its own orange, the rest quiet.
func sourcePill(context *ui.Context, words string, channel Channel) {
	theme := context.Theme
	fill, ink := theme.Sunken, theme.InkSoft
	if channel == Marketplace || channel == OnlineShop {
		fill, ink = marketplaceFill, marketplaceInk
	}
	context.Frame(ui.Layout{Padding: ui.PadXY(sourcePaddingX, sourcePaddingY)}, ui.Look{Fill: fill, Radius: sourceSize}, func() {
		context.Text(words, ui.TextStyle{Size: sourceSize, Weight: draw2d.Bold, Color: ink})
	})
}

// piecesWords counts the units of a sale: "1 pieza", "3 piezas".
func piecesWords(sale Sale) string {
	count := 0
	for _, line := range sale.Lines {
		count += line.Quantity
	}
	if count == 1 {
		return "1 pieza"
	}
	return fmt.Sprintf("%d piezas", count)
}

// whoOrdered is the customer's name, or the channel when the order has none.
func whoOrdered(sale Sale) string {
	if sale.Customer != "" {
		return sale.Customer
	}
	return channelWords[sale.Channel]
}

func (screens *Screens) afterAdvance(from Status, problems []Problem, err error) {
	switch {
	case err != nil:
		screens.tell(true, err.Error())
	case len(problems) > 0:
		screens.tell(true, ProblemWords(problems[0]))
	case from == Sent:
		screens.tell(false, "Entregado: el pedido pasa al historial")
	default:
		screens.tell(false, "Preparado: el pedido está en camino")
	}
}
