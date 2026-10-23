package stand

import (
	"fmt"
	"slices"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
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
const tileWidth = 210

// How much bigger than a heading the choice's title and the fair on today are.
const (
	choiceTitleScale = 1.4
	heroTitleScale   = 1.6
)

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

// Measures of the choice of stands, in pixels.
const (
	choiceWidth   = 980
	heroBadgeSide = 100
	choiceBadge   = 64
	choiceColumns = 3
	choiceHeight  = 230
)

// drawChoice is where the artist picks what to sell from: the fair on today, big and dark, then the
// other open fairs and the whole stock.
func (screens *Screens) drawChoice(context *ui.Context) {
	theme := context.Theme
	context.Box(ui.Layout{Width: ui.Grow(), AlignX: ui.Center}, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(choiceWidth), Gap: theme.Gap, Padding: ui.Padding{Top: theme.Padding}}, func() {
			context.Text("¿Dónde vendes hoy?", ui.TextStyle{Size: theme.HeadingSize * choiceTitleScale, Weight: draw2d.Bold})
			context.Text("Elige la feria y lo que vendas quedará apuntado en ella.", ui.TextStyle{Color: theme.InkSoft})
			if screens.loadErr != nil {
				look.Empty(context, "triangle-alert", "No se pudieron leer las ferias", screens.loadErr.Error(), "")
				return
			}
			today := calendar.Of(screens.now())
			others := screens.fairs
			if index := slices.IndexFunc(screens.fairs, func(fair events.Event) bool { return today.Within(fair.StartDate, fair.EndDate) }); index >= 0 {
				screens.heroCard(context, screens.fairs[index])
				others = slices.Delete(slices.Clone(screens.fairs), index, index+1)
			}
			context.Grid(look.Columns(context, choiceColumns), func() {
				for _, fair := range others {
					context.Push(fair.ID)
					if screens.choiceCard(context, "calendar", fair.Name, calendar.RangeWords(fair.StartDate, fair.EndDate)) {
						screens.chooseFair(fair)
					}
					context.Pop()
				}
				if screens.choiceCard(context, "package", "Todo el stock", "Todos tus productos, para vender fuera de una feria") {
					screens.Choose(WholeStock{Sales: screens.sales, Products: screens.products, Now: screens.now})
				}
			})
		})
	})
}

func (screens *Screens) chooseFair(fair events.Event) {
	screens.Choose(Fair{Event: fair, Events: screens.events, Products: screens.products})
}

// heroCard is the fair on today: dark, with what the artist took and «Vender aquí».
func (screens *Screens) heroCard(context *ui.Context, fair events.Event) {
	theme := context.Theme
	light := theme.OnPrimary
	layout := ui.Layout{Width: ui.Grow(), Padding: ui.Pad(theme.Padding * 1.8), Gap: 28, AlignY: ui.Center}
	context.Frame(layout, ui.Look{Fill: theme.Primary, Radius: theme.PanelRadius}, func() {
		badge := ui.Layout{Width: ui.Fixed(heroBadgeSide), Height: ui.Fixed(heroBadgeSide), AlignX: ui.Center, AlignY: ui.Center}
		context.Frame(badge, ui.Look{Fill: light.WithAlpha(0.12), Radius: heroBadgeSide / 2}, func() {
			context.IconSized("tent", theme.IconSize*1.5, light)
		})
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 4}, func() {
			context.Text("● EN CURSO", ui.TextStyle{Size: theme.SmallTextSize, Weight: draw2d.Bold, Color: light.WithAlpha(0.8)})
			context.Text(fair.Name, ui.TextStyle{Size: theme.HeadingSize * heroTitleScale, Weight: draw2d.Bold, Color: light})
			context.Text(screens.heroWords(fair), ui.TextStyle{Color: light.WithAlpha(0.8)})
		})
		pill := ui.Look{Fill: theme.Surface, Radius: theme.ControlHeight / 2}
		button := ui.Layout{Height: ui.Fixed(theme.ControlHeight * 1.1), Padding: ui.PadXY(22, 0), AlignY: ui.Center}
		if context.Clickable("Vender aquí", button, ui.Looks{Normal: pill, Hovered: pill, Pressed: pill}, func() {
			context.Text("Vender aquí", ui.TextStyle{Weight: draw2d.Bold})
		}) {
			screens.chooseFair(fair)
		}
	})
}

// heroWords is where the fair is and how many pieces the artist has there.
func (screens *Screens) heroWords(fair events.Event) string {
	words := "Hoy"
	if fair.Location != "" {
		words += " · " + fair.Location
	}
	detail, err := screens.events.Detail(fair.ID)
	if err != nil {
		return words // the count is a nicety; the fair can still be chosen
	}
	left := 0
	for _, line := range detail.Stock {
		left += line.Remaining()
	}
	return words + fmt.Sprintf(" · llevas %d piezas", left)
}

func (screens *Screens) choiceCard(context *ui.Context, icon, title, subtitle string) bool {
	theme := context.Theme
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Fixed(choiceHeight), Padding: ui.Pad(18), Gap: 10}
	return look.ClickableCard(context, title, layout, func() {
		look.Badge(context, icon, choiceBadge)
		look.Paragraph(context, title, ui.TextStyle{Weight: draw2d.Bold})
		look.Paragraph(context, subtitle, ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
	})
}

// drawRegister is the stand itself: its name, the tags and the products; the basket floats over.
func (screens *Screens) drawRegister(context *ui.Context) {
	theme := context.Theme
	if screens.loadErr != nil {
		look.Empty(context, "triangle-alert", "No se pudo leer el puesto", screens.loadErr.Error(), "")
		return
	}
	tags := screens.tags()
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 14, AlignY: ui.Center}, func() {
		look.Badge(context, map[bool]string{true: "package", false: "tent"}[screens.source.Name() == WholeStock{}.Name()], choiceBadge*0.8)
		context.Box(ui.Layout{Direction: ui.Vertical, Gap: 2}, func() {
			context.Text(screens.source.Name(), ui.TextStyle{Weight: draw2d.Bold})
			if look.Link(context, "Cambiar") {
				screens.Choose(nil)
			}
		})
		context.Frame(ui.Layout{Width: ui.Fixed(1), Height: ui.Fixed(theme.ControlHeight)}, ui.Look{Fill: theme.Line}, nil)
		if len(tags) > 0 {
			tabs := []look.Tab{{Label: "Todo"}}
			for _, tag := range tags {
				tabs = append(tabs, look.Tab{Label: tag})
			}
			look.PillTabs(context, "tag", &screens.tag, tabs)
		}
	})
	if len(screens.items) == 0 {
		look.Empty(context, "package-open", "No hay nada para vender", "Lleva stock a la feria, o añade productos al catálogo.", "")
		return
	}
	width, _ := context.Size()
	context.Grid(max(2, int(width/tileWidth)), func() {
		for _, item := range screens.shown(tags) {
			context.Push(item.VariantID)
			screens.tile(context, item)
			context.Pop()
		}
	})
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

// DrawBasketBar is the basket floating at the bottom right of the stand, over the products while
// they scroll: how many pieces, their total and «Cobrar». It draws nothing while the basket is empty.
func (screens *Screens) DrawBasketBar(context *ui.Context) {
	count := screens.basket.Count()
	if screens.source == nil || count == 0 || screens.step != browsing {
		return
	}
	theme := context.Theme
	light := theme.OnPrimary
	options := ui.LayerOptions{AlignX: ui.End, AlignY: ui.End, Padding: ui.Padding{Right: basketMargin / 3, Bottom: basketMargin}}
	context.Overlay("basket", options, func() {
		layout := ui.Layout{Padding: ui.Padding{Left: 22, Right: 6, Top: 6, Bottom: 6}, Gap: 14, AlignY: ui.Center}
		context.Frame(layout, ui.Look{Fill: theme.Primary, Radius: basketHeight / 2, Shadow: true}, func() {
			context.IconSized("shopping-basket", theme.IconSize, light)
			context.Text(piecesWords(count), ui.TextStyle{Color: light})
			context.Text(money.Format(screens.basket.Total()), ui.TextStyle{Size: theme.TextSize * 1.2, Weight: draw2d.Bold, Color: light})
			pill := ui.Look{Fill: theme.Surface, Radius: basketHeight / 2}
			button := ui.Layout{Height: ui.Fixed(basketHeight - 12), Padding: ui.PadXY(22, 0), AlignY: ui.Center}
			if context.Clickable("Cobrar", button, ui.Looks{Normal: pill, Hovered: pill, Pressed: pill}, func() {
				context.Text("Cobrar", ui.TextStyle{Weight: draw2d.Bold})
			}) {
				screens.step = choosingMethod
			}
		})
	})
}

// Measures of the floating basket, in pixels.
const (
	basketHeight = 60
	basketMargin = 100 // above the app's bar
)

func piecesWords(count int) string {
	switch count {
	case 1:
		return "1 artículo"
	}
	return fmt.Sprintf("%d artículos", count)
}

// ShowBasket puts the first item twice and the second once in the basket, for captures of the stand.
func (screens *Screens) ShowBasket() {
	for index, times := range []int{2, 1} {
		for range times {
			if index < len(screens.items) {
				screens.basket.Add(screens.items[index])
			}
		}
	}
}

// ShowPaymentChoice opens the choice of cash or card over ShowBasket's basket, for captures.
func (screens *Screens) ShowPaymentChoice() {
	screens.ShowBasket()
	screens.step = choosingMethod
}

// ShowCashPayment opens the cash dialog over ShowBasket's basket with 20 € handed over, for
// captures of the stand.
func (screens *Screens) ShowCashPayment() {
	screens.ShowBasket()
	screens.received.Clear()
	screens.received.Add(2000)
	screens.received.Add(2000)
	screens.step = payingCash
}
