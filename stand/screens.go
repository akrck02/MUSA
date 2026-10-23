package stand

import (
	"fmt"
	"slices"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/cash"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
	"github.com/Itros97/MUSA/sales"
)

// noteTime is how long a note about what an action did stays on screen.
const noteTime = 3 * time.Second

// tileWidth is the width a product tile wants; the grid fits as many as the window allows.
const tileWidth = 220

// step is where the checkout is: browsing the stand, or one of the payment dialogs over it.
type step int

const (
	browsing step = iota
	choosingMethod
	payingCash
	payingCard
	paid
)

// Screens are the stand's pages: choosing a stand, and selling from it.
type Screens struct {
	events   *events.Store
	products *catalogue.Store
	sales    *sales.Store
	newID    func() string
	now      func() time.Time

	fairs   []events.Event
	source  Source
	items   []Item
	loadErr error
	tag     int // 0 every tag, then 1 + index in tags()

	basket   Basket
	step     step
	received cash.Received
	receipt  receipt

	note        string
	noteWarning bool
	noteUntil   time.Time
}

// receipt is what the last sale took and gave back, for the dialog that closes it.
type receipt struct {
	total, change money.Money
	method        sales.Method
}

// NewScreens sells from the open fairs kept in fairs, or from the catalogue in products.
func NewScreens(fairs *events.Store, products *catalogue.Store, sold *sales.Store, newID func() string, now func() time.Time) *Screens {
	return &Screens{events: fairs, products: products, sales: sold, newID: newID, now: now}
}

// Reload reads the open fairs and what the chosen stand has left.
func (screens *Screens) Reload() {
	all, err := screens.events.All()
	screens.fairs, screens.loadErr = nil, err
	for _, event := range all {
		if event.Status == events.Open {
			screens.fairs = append(screens.fairs, event)
		}
	}
	if screens.source != nil && err == nil {
		screens.items, screens.loadErr = screens.source.Items()
	}
}

// Back closes the dialog open, or leaves the stand for the choice of stands; false when only the
// choice was open.
func (screens *Screens) Back() bool {
	switch {
	case screens.step != browsing:
		screens.step = browsing
	case screens.source != nil:
		screens.Choose(nil)
	default:
		return false
	}
	return true
}

// Choose opens a stand; nil goes back to the choice of stands. The basket starts empty.
func (screens *Screens) Choose(source Source) {
	screens.source, screens.tag, screens.step = source, 0, browsing
	screens.basket.Clear()
	screens.Reload()
}

// Draw draws the open page of the stand.
func (screens *Screens) Draw(context *ui.Context) {
	if screens.source == nil {
		screens.drawChoice(context)
	} else {
		screens.drawRegister(context)
		screens.drawCheckout(context)
	}
	if time.Now().Before(screens.noteUntil) {
		context.Toast(screens.note, screens.noteWarning)
		context.RequestFrame(time.Until(screens.noteUntil))
	}
}

func (screens *Screens) tell(warning bool, words string) {
	screens.note, screens.noteWarning, screens.noteUntil = words, warning, time.Now().Add(noteTime)
}

// drawChoice is where the artist picks what to sell from: an open fair, or the whole stock.
func (screens *Screens) drawChoice(context *ui.Context) {
	look.PageHeader(context, "Puesto", "¿Desde dónde vendes hoy?", nil)
	if screens.loadErr != nil {
		look.Empty(context, "triangle-alert", "No se pudieron leer las ferias", screens.loadErr.Error(), "")
		return
	}
	context.Grid(look.Columns(context, 2), func() {
		for _, fair := range screens.fairs {
			context.Push(fair.ID)
			if screens.choiceCard(context, "tent", fair.Name, "Lo que llevaste a la feria") {
				screens.Choose(Fair{Event: fair, Events: screens.events, Products: screens.products})
			}
			context.Pop()
		}
		if screens.choiceCard(context, "store", "Todo el stock", "Todo lo que tienes en el catálogo") {
			screens.Choose(WholeStock{Sales: screens.sales, Products: screens.products, Now: screens.now})
		}
	})
}

func (screens *Screens) choiceCard(context *ui.Context, icon, title, subtitle string) bool {
	theme := context.Theme
	return look.ClickableCard(context, title, ui.Layout{Width: ui.Grow(), Padding: ui.Pad(20), Gap: 16, AlignY: ui.Center}, func() {
		look.Badge(context, icon, 56)
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, func() {
			context.Text(title, ui.TextStyle{Weight: draw2d.Bold})
			context.Text(subtitle, ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
		})
	})
}

// drawRegister is the stand itself: the tags, the products and the basket's bar at the bottom.
func (screens *Screens) drawRegister(context *ui.Context) {
	look.PageHeader(context, screens.source.Name(), "Toca un producto para añadirlo", func() {
		if context.IconButton("arrow-left-right", "Cambiar") {
			screens.Choose(nil)
		}
	})
	if screens.loadErr != nil {
		look.Empty(context, "triangle-alert", "No se pudo leer el puesto", screens.loadErr.Error(), "")
		return
	}
	if len(screens.items) == 0 {
		look.Empty(context, "package-open", "No hay nada para vender", "Lleva stock a la feria, o añade productos al catálogo.", "")
		return
	}
	tags := screens.tags()
	if len(tags) > 0 {
		context.Tabs("tag", &screens.tag, append([]string{"Todo"}, tags...)...)
	}
	width, _ := context.Size()
	context.Grid(max(2, int(width/tileWidth)-1), func() {
		for _, item := range screens.shown(tags) {
			context.Push(item.VariantID)
			screens.tile(context, item)
			context.Pop()
		}
	})
	screens.basketBar(context)
}

// tags is every tag of the stand's items, in the order they first appear.
func (screens *Screens) tags() []string {
	var tags []string
	for _, item := range screens.items {
		for _, tag := range item.Tags {
			if !slices.Contains(tags, tag) {
				tags = append(tags, tag)
			}
		}
	}
	return tags
}

// shown is the items of the chosen tag, or all of them.
func (screens *Screens) shown(tags []string) []Item {
	if screens.tag == 0 || screens.tag > len(tags) {
		return screens.items
	}
	var shown []Item
	for _, item := range screens.items {
		if slices.Contains(item.Tags, tags[screens.tag-1]) {
			shown = append(shown, item)
		}
	}
	return shown
}

// tile is one product: tapping it adds one to the basket; once chosen, it shows how many and a
// button to take one back.
func (screens *Screens) tile(context *ui.Context, item Item) {
	theme := context.Theme
	chosen := screens.basket.Quantity(item.VariantID)
	label := fmt.Sprintf("%s · %s", item.Label, money.Format(item.Price))
	if look.ClickableCard(context, label, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(14), Gap: 6}, func() {
		context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, ui.Look{Clip: true}, func() {
			look.Paragraph(context, item.Label, ui.TextStyle{Weight: draw2d.Bold})
			context.Text(money.Format(item.Price), ui.TextStyle{Color: theme.InkSoft})
		})
		left := item.Left - chosen
		context.Text(fmt.Sprintf("Quedan %d", left), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
		if chosen > 0 {
			screens.chosenRow(context, item, chosen)
		}
	}) && !screens.basket.Add(item) {
		screens.tell(true, "No quedan más")
	}
}

// chosenRow is the slate band of a chosen tile: how many are in the basket, and «−».
func (screens *Screens) chosenRow(context *ui.Context, item Item, chosen int) {
	theme := context.Theme
	context.Frame(ui.Layout{Width: ui.Grow(), Padding: ui.Pad(6), Gap: 8, AlignY: ui.Center},
		ui.Look{Fill: theme.Primary, Radius: theme.ControlRadius}, func() {
			context.Box(ui.Layout{Width: ui.Grow()}, func() {
				context.Text(fmt.Sprintf("× %d", chosen), ui.TextStyle{Weight: draw2d.Bold, Color: theme.OnPrimary})
			})
			if context.IconButton("minus", "##Quitar uno") {
				screens.basket.Remove(item.VariantID)
			}
		})
}

// basketBar is the bottom of the stand: what the basket holds, its total and «Cobrar».
func (screens *Screens) basketBar(context *ui.Context) {
	theme := context.Theme
	count := screens.basket.Count()
	look.Card(context, ui.Layout{Width: ui.Grow(), Padding: ui.Pad(16), Gap: 16, AlignY: ui.Center}, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, func() {
			context.Text(piecesWords(count), ui.TextStyle{Color: theme.InkSoft})
			context.Text(money.Format(screens.basket.Total()), ui.TextStyle{Size: theme.TextSize * 1.6, Weight: draw2d.Bold})
		})
		if count == 0 {
			return
		}
		if context.FlatButton("Vaciar") {
			screens.basket.Clear()
		}
		if context.PrimaryIconButton("hand-coins", "Cobrar") {
			screens.step = choosingMethod
		}
	})
}

func piecesWords(count int) string {
	switch count {
	case 0:
		return "Toca los productos que se lleva"
	case 1:
		return "1 pieza"
	}
	return fmt.Sprintf("%d piezas", count)
}

// ShowCashPayment opens the cash dialog with the first item twice in the basket and 20 € handed
// over, for captures of the stand.
func (screens *Screens) ShowCashPayment() {
	if len(screens.items) == 0 {
		return
	}
	screens.basket.Add(screens.items[0])
	screens.basket.Add(screens.items[0])
	screens.received.Clear()
	screens.received.Add(2000)
	screens.step = payingCash
}
