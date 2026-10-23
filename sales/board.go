package sales

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

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

// drawBoard is the orders still to deliver, in one column per status.
func (screens *Screens) drawBoard(context *ui.Context, onTheWay []Sale) {
	if len(onTheWay) == 0 {
		look.Empty(context, "package-check", "Nada por enviar",
			"Los pedidos entregados y las ventas locales están en el historial.", "")
		return
	}
	context.Grid(look.Columns(context, len(boardColumns)), func() {
		for _, status := range boardColumns {
			screens.drawColumn(context, status, onTheWay)
		}
	})
}

func (screens *Screens) drawColumn(context *ui.Context, status Status, onTheWay []Sale) {
	var orders []Sale
	for _, sale := range onTheWay {
		if sale.Status == status {
			orders = append(orders, sale)
		}
	}
	look.Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(16), Gap: 12}, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 12, AlignY: ui.Center}, func() {
			look.Badge(context, columnIcons[status], 36)
			context.Box(ui.Layout{Width: ui.Grow()}, func() {
				context.Text(StatusWords(status), ui.TextStyle{Weight: draw2d.Bold})
			})
			context.Text(fmt.Sprint(len(orders)), ui.TextStyle{Color: context.Theme.InkSoft})
		})
		for _, sale := range orders {
			context.Separator()
			context.Push(sale.ID)
			screens.boardRow(context, sale)
			context.Pop()
		}
	})
}

// boardRow is one order on the board: who and what on the left, the button that moves it on the right.
func (screens *Screens) boardRow(context *ui.Context, sale Sale) {
	theme := context.Theme
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 12, AlignY: ui.Center}, func() {
		context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, ui.Look{Clip: true}, func() {
			context.Text(whoOrdered(sale), ui.TextStyle{Weight: draw2d.Bold})
			context.Text(fmt.Sprintf("%s · %s", screens.headline(sale), money.Format(sale.Total())),
				ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
		})
		if context.Button(actionWords[sale.Status]) {
			problems, err := screens.store.Advance(sale.ID)
			screens.afterAdvance(sale.Status, problems, err)
		}
	})
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
