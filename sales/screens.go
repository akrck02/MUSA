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
	// positions is each product's place in the catalogue, by name, which picks its colour.
	positions map[string]int
	view      int // 0 the board of orders on the way, 1 the sales history
	period    int // index in Periods
	// choosingPeriod is true while the list of periods is open, under the period's figure.
	choosingPeriod bool
	kind           int // index in historyKinds
	// eventName names a fair by its id, empty when unknown; sales cannot read fairs themselves.
	eventName func(id string) string
	opened    string
	editor    *orderEditor

	note        string
	noteWarning bool
	noteUntil   time.Time
}

// NewScreens draws the sales kept in store, picking what they sell from products.
func NewScreens(store *Store, products *catalogue.Store, newID func() string, now func() time.Time) *Screens {
	return &Screens{store: store, products: products, newID: newID, now: now}
}

// Reload reads the sales again, and the catalogue's order for their colours.
func (screens *Screens) Reload() {
	screens.sales, screens.loadErr = screens.store.All()
	all, err := screens.products.All()
	if err != nil {
		screens.loadErr = err
		return
	}
	screens.positions = make(map[string]int, len(all))
	for index, product := range all {
		screens.positions[product.Name] = index
	}
}

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

// views are the two tabs of the orders page.
var views = []string{"Por enviar", "Historial de ventas"}

func (screens *Screens) drawList(context *ui.Context) {
	onTheWay := OnTheWay(screens.sales)
	look.PageHeader(context, "Pedidos", boardWords(onTheWay), func() {
		if look.MainButton(context, "plus", "Nuevo pedido") {
			screens.StartOrder()
		}
	})
	tabs := make([]look.Tab, len(views))
	for index, words := range views {
		tabs[index] = look.Tab{Label: words}
	}
	look.Space(context, viewsTop)
	context.Box(ui.Layout{Width: ui.Grow(), AlignY: ui.Center}, func() {
		look.PillTabs(context, "view", &screens.view, tabs, viewTabs)
		context.Spacer()
		// TODO(akrck02): say when the orders last synced with the shops, once there is an API to sync with.
		look.SyncStatus(context, look.Unsynced)
	})
	look.Space(context, viewsBottom)
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
		if context.DangerIconButton("trash-2", "Eliminar venta") {
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

// ShowHistory opens the sales history tab, for captures.
func (screens *Screens) ShowHistory() { screens.view = 1 }

// NameEvents tells the screens how to name a fair by its id, so a fair's sales show its name.
func (screens *Screens) NameEvents(name func(id string) string) { screens.eventName = name }
