package sales

import (
	"fmt"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
)

// noteTime is how long a note about what an action did stays on screen.
const noteTime = 3 * time.Second

// Screens are the orders pages: the board of orders on the way, the sales history with its
// filters and totals, an order's detail over them, and the order editor.
type Screens struct {
	store    *Store
	products *catalogue.Store
	newID    func() string
	now      func() time.Time

	sales   []Sale
	loadErr error
	view    int // 0 the board of orders on the way, 1 the sales history
	period  int // index in Periods
	channel int // 0 every channel, then 1 + index in channelChoices
	opened  string
	editor  *orderEditor

	note        string
	noteWarning bool
	noteUntil   time.Time
}

// channelChoices are the channels the list filters by: every one, fairs included.
var channelChoices = append([]Channel{AtEvent}, SelectableChannels...)

// NewScreens draws the sales kept in store, picking what they sell from products.
func NewScreens(store *Store, products *catalogue.Store, newID func() string, now func() time.Time) *Screens {
	return &Screens{store: store, products: products, newID: newID, now: now}
}

// Reload reads the sales again.
func (screens *Screens) Reload() { screens.sales, screens.loadErr = screens.store.All() }

// Back closes what is open over the list; false when only the list was open.
func (screens *Screens) Back() bool {
	switch {
	case screens.editor != nil:
		screens.editor = nil
	case screens.opened != "":
		screens.opened = ""
	default:
		return false
	}
	return true
}

// Draw draws the open page of the sales.
func (screens *Screens) Draw(context *ui.Context) {
	if screens.editor != nil {
		screens.drawEditor(context)
	} else {
		screens.drawList(context)
	}
	if time.Now().Before(screens.noteUntil) {
		context.Toast(screens.note, screens.noteWarning)
		context.RequestFrame(time.Until(screens.noteUntil))
	}
}

func (screens *Screens) tell(warning bool, words string) {
	screens.note, screens.noteWarning, screens.noteUntil = words, warning, time.Now().Add(noteTime)
	screens.Reload()
}

func (screens *Screens) query() Query {
	query := Query{Period: Periods[screens.period]}
	if screens.channel > 0 {
		query.Channel = channelChoices[screens.channel-1]
	}
	return query
}

// views are the two tabs of the orders page.
var views = []string{"Por enviar", "Historial de ventas"}

func (screens *Screens) drawList(context *ui.Context) {
	onTheWay := OnTheWay(screens.sales)
	look.PageHeader(context, "Pedidos", boardWords(onTheWay), func() {
		if context.PrimaryIconButton("plus", "Nuevo pedido") {
			screens.StartOrder()
		}
	})
	context.Tabs("view", &screens.view, views...)
	switch {
	case screens.loadErr != nil:
		look.Empty(context, "triangle-alert", "No se pudieron leer los pedidos", screens.loadErr.Error(), "")
	case screens.view == 0:
		screens.drawBoard(context, onTheWay)
	default:
		screens.drawHistory(context)
	}
	if screens.opened != "" {
		screens.detailSheet(context)
	}
}

// drawHistory is the delivered orders, local sales among them, of a period and a channel.
func (screens *Screens) drawHistory(context *ui.Context) {
	theme := context.Theme
	now := screens.now()
	shown := Filter(Done(screens.sales), screens.query(), calendar.Of(now), now.Location())
	context.Box(ui.Layout{Width: ui.Grow(), Gap: theme.Gap, AlignY: ui.Center}, func() {
		var periods []string
		for _, period := range Periods {
			periods = append(periods, periodWords[period])
		}
		context.Box(ui.Layout{Width: ui.Grow()}, func() { context.Tabs("period", &screens.period, periods...) })
		channels := []string{"Todos los canales"}
		for _, channel := range channelChoices {
			channels = append(channels, channelWords[channel])
		}
		context.Select("channel", &screens.channel, channels)
	})
	summary := Summarise(shown)
	look.Figures(context, 4, func() {
		look.Figure(context, "Ventas", fmt.Sprint(summary.Count), false)
		look.Figure(context, "Ingresos", money.Format(summary.Revenue), false)
		look.Figure(context, "Coste", money.Format(summary.Cost), false)
		look.Figure(context, "Te queda", money.Format(summary.Margin()), summary.Margin() < 0)
	})
	if len(shown) == 0 {
		look.Empty(context, "receipt", "Sin ventas en este periodo", "Mira otro periodo, u otro canal.", "")
		return
	}
	for _, sale := range shown {
		context.Push(sale.ID)
		if screens.row(context, sale) {
			screens.opened = sale.ID
		}
		context.Pop()
	}
}

// row is one sale in the list; it returns true on the frame it is clicked.
func (screens *Screens) row(context *ui.Context, sale Sale) bool {
	theme := context.Theme
	return look.ClickableCard(context, screens.headline(sale), ui.Layout{Width: ui.Grow(), Padding: ui.Pad(12), Gap: 12, AlignY: ui.Center}, func() {
		context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, ui.Look{Clip: true}, func() {
			context.Text(screens.headline(sale), ui.TextStyle{Weight: draw2d.Bold})
			context.Text(calendar.Words(calendar.Of(sale.SoldAt))+" · "+channelWords[sale.Channel]+customerWords(sale),
				ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
		})
		context.Box(ui.Layout{Direction: ui.Vertical, Gap: 2, AlignX: ui.End}, func() {
			context.Text(money.Format(sale.Total()), ui.TextStyle{Weight: draw2d.Bold})
			colour := theme.InkSoft
			if sale.Margin() < 0 {
				colour = theme.Danger
			}
			context.Text("margen "+money.Format(sale.Margin()), ui.TextStyle{Size: theme.SmallTextSize, Color: colour})
		})
	})
}

// headline is the first line of a sale, or "first line +2" when it has more.
func (screens *Screens) headline(sale Sale) string {
	if len(sale.Lines) == 0 {
		return "Venta"
	}
	first := sale.Lines[0]
	words := fmt.Sprintf("%d × %s", first.Quantity, ItemLabel(first.ProductName, first.VariantName))
	if len(sale.Lines) > 1 {
		words += fmt.Sprintf(" +%d", len(sale.Lines)-1)
	}
	return words
}

func customerWords(sale Sale) string {
	if sale.Customer == "" {
		return ""
	}
	return " · " + sale.Customer
}

func (screens *Screens) detailSheet(context *ui.Context) {
	sale, found := screens.find(screens.opened)
	if !found {
		screens.opened = ""
		return
	}
	theme := context.Theme
	if context.Sheet("Venta del "+calendar.Words(calendar.Of(sale.SoldAt)), ui.SheetOptions{}, func() {
		context.Text(channelWords[sale.Channel]+customerWords(sale), ui.TextStyle{Color: theme.InkSoft})
		if sale.Note != "" {
			context.Text(sale.Note, ui.TextStyle{})
		}
		for _, line := range sale.Lines {
			amountRow(context, fmt.Sprintf("%d × %s", line.Quantity, ItemLabel(line.ProductName, line.VariantName)), line.Total(), false)
		}
		if sale.ShippingCharged != 0 || sale.ShippingCost != 0 {
			amountRow(context, "Envío cobrado", sale.ShippingCharged, false)
			amountRow(context, "Envío pagado", -sale.ShippingCost, false)
		}
		context.Separator()
		amountRow(context, "Total", sale.Total(), true)
		amountRow(context, "Coste", sale.Cost(), false)
		amountRow(context, "Margen", sale.Margin(), true)
		if sale.EventID != "" {
			context.Text(ProblemWords(Problem{Code: CodeManagedByEvent}), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
			return
		}
		if context.IconButton("trash-2", "Eliminar venta") {
			problems, err := screens.store.Delete(sale.ID)
			screens.opened = ""
			screens.afterDelete(problems, err)
		}
	}) {
		screens.opened = ""
	}
}

func (screens *Screens) afterDelete(problems []Problem, err error) {
	switch {
	case err != nil:
		screens.tell(true, err.Error())
	case len(problems) > 0:
		screens.tell(true, ProblemWords(problems[0]))
	default:
		screens.tell(false, "Venta eliminada; el stock ha vuelto al catálogo")
	}
}

// amountRow is a label on the left and an amount on the right.
func amountRow(context *ui.Context, label string, amount money.Money, bold bool) {
	weight := draw2d.Regular
	if bold {
		weight = draw2d.Bold
	}
	colour := context.Theme.Ink
	if amount < 0 && bold {
		colour = context.Theme.Danger
	}
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 12}, func() {
		context.Box(ui.Layout{Width: ui.Grow()}, func() { context.Text(label, ui.TextStyle{Weight: weight}) })
		context.Text(money.Format(amount), ui.TextStyle{Weight: weight, Color: colour})
	})
}

func (screens *Screens) find(id string) (Sale, bool) {
	for _, sale := range screens.sales {
		if sale.ID == id {
			return sale, true
		}
	}
	return Sale{}, false
}

// StartOrder opens the editor of a new order.
func (screens *Screens) StartOrder() {
	editing := newOrderEditor(screens.newID, calendar.Of(screens.now()))
	screens.editor = &editing
}
