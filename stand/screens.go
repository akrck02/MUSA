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

	fairs     []events.Event
	source    Source
	items     []Item
	loadErr   error
	tag       int // index in tags()
	searching bool
	search    string
	shapes    map[draw2d.Color]*look.TopRounded // the pictures' fills, kept so drawing them allocates nothing
	fade      *look.Fade                        // under the floating basket

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
	saleID        string
	total, change money.Money
	method        sales.Method
	taken         []Taken
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

// Measures of the stand's top, in pixels, as in the author's design.
const (
	topSpaceBelow  = 16
	whereSide      = 48
	whereRadius    = 12
	whereIconSize  = 24
	whereGap       = 12
	whereSpace     = 18
	whereNameSize  = 19
	whereLineScale = 1.25
	whereLinkSize  = 15
	dividerHeight  = 40
	dividerSpace   = 12
	toolsGap       = 10
	tilesGap       = 18
	tilesColumns   = 6
)

// drawRegister is the stand itself: where it sells from, the tags and the products; the basket
// floats over.
func (screens *Screens) drawRegister(context *ui.Context) {
	if screens.loadErr != nil {
		look.Empty(context, "triangle-alert", "No se pudo leer el puesto", screens.loadErr.Error(), "")
		return
	}
	tags := screens.tags()
	screens.top(context, tags)
	look.Space(context, topSpaceBelow)
	if len(screens.items) == 0 {
		look.Empty(context, "package-open", "No hay nada para vender", "Lleva stock a la feria, o añade productos al catálogo.", "")
		return
	}
	width, _ := context.Size()
	columns := min(tilesColumns, max(2, int(width/tileWidth)))
	context.Box(ui.Layout{Width: ui.Grow(), Columns: columns, Gap: tilesGap}, func() {
		for position, item := range screens.shown(tags) {
			context.Push(item.VariantID)
			screens.tile(context, item, catalogue.ListFill(context.Theme, position))
			context.Pop()
		}
	})
}

// top is where the stand sells from with «Cambiar», a tab per tag, and the search and «+».
func (screens *Screens) top(context *ui.Context, tags []string) {
	context.Box(ui.Layout{Width: ui.Grow(), AlignY: ui.Center}, func() {
		screens.where(context)
		look.Space(context, whereSpace)
		context.Frame(ui.Layout{Width: ui.Fixed(1), Height: ui.Fixed(dividerHeight)}, ui.Look{Fill: context.Theme.Line}, nil)
		look.Space(context, dividerSpace)
		context.Box(ui.Layout{Width: ui.Grow()}, func() {
			if len(tags) > 0 {
				tabs := make([]look.Tab, len(tags))
				for index, tag := range tags {
					tabs[index] = look.Tab{Icon: catalogue.TagIcon(tag), Label: catalogue.Plural(tag)}
				}
				look.PillTabs(context, "tag", &screens.tag, tabs, look.LargeTabs)
			}
		})
		context.Box(ui.Layout{Gap: toolsGap}, func() {
			if look.ToolButton(context, "search", "Buscar") {
				screens.searching = !screens.searching
			}
			if look.ToolButton(context, "plus", "Venta libre") {
				screens.tell(false, "Pronto: vender algo que no está en el catálogo")
			}
		})
	})
}

// where is the stand's picture, its name and «Cambiar», which goes back to the choice of stands.
func (screens *Screens) where(context *ui.Context) {
	theme := context.Theme
	context.Box(ui.Layout{Gap: whereGap, AlignY: ui.Center}, func() {
		side := ui.Fixed(whereSide)
		context.Frame(ui.Layout{Width: side, Height: side, AlignX: ui.Center, AlignY: ui.Center}, ui.Look{Fill: theme.Sunken, Radius: whereRadius}, func() {
			context.IconSized(screens.sourceIcon(), whereIconSize, theme.Primary)
		})
		context.Box(ui.Layout{Direction: ui.Vertical}, func() {
			context.Text(screens.source.Name(), ui.TextStyle{Size: whereNameSize, Weight: draw2d.Bold, LineHeight: whereLineScale})
			if look.QuietLink(context, "Cambiar", whereLinkSize, whereLineScale) {
				screens.Choose(nil)
			}
		})
	})
}

// sourceIcon is a box for the whole stock and a tent for a fair.
func (screens *Screens) sourceIcon() string {
	if screens.source.Name() == (WholeStock{}).Name() {
		return "package"
	}
	return "tent"
}

// tags is every tag of the stand's items, the one with most items first.
func (screens *Screens) tags() []string {
	var tags []string
	counts := map[string]int{}
	for _, item := range screens.items {
		for _, tag := range item.Tags {
			if counts[tag] == 0 {
				tags = append(tags, tag)
			}
			counts[tag]++
		}
	}
	slices.SortStableFunc(tags, func(first, second string) int { return counts[second] - counts[first] })
	return tags
}

// shown is the items of the chosen tag, or all of them when there are no tags.
func (screens *Screens) shown(tags []string) []Item {
	if screens.tag >= len(tags) {
		return screens.items
	}
	var shown []Item
	for _, item := range screens.items {
		if slices.Contains(item.Tags, tags[screens.tag]) {
			shown = append(shown, item)
		}
	}
	return shown
}

// DrawBasketBar is the basket floating at the bottom right of the stand, over the products while
// they scroll into a fade: how many pieces, their total and «Cobrar». above is how tall the app's
// bar under it is. It draws nothing while the basket is empty.
func (screens *Screens) DrawBasketBar(context *ui.Context, above float32) {
	count := screens.basket.Count()
	// It stays under a payment dialog too, dimmed with the rest of the stand, and leaves with the
	// sale.
	if screens.source == nil || count == 0 {
		return
	}
	theme := context.Theme
	if screens.fade == nil {
		screens.fade = &look.Fade{Solid: fadeSolid}
	}
	screens.fade.Colour = theme.Background
	fadeOptions := ui.LayerOptions{AlignY: ui.End, Width: ui.Grow(), Padding: ui.Padding{Bottom: above}}
	context.Overlay("fade", fadeOptions, func() {
		context.Frame(ui.Layout{Width: ui.Grow(), Height: ui.Fixed(fadeHeight)}, ui.Look{Backdrop: screens.fade}, nil)
	})
	options := ui.LayerOptions{AlignX: ui.End, AlignY: ui.End, Padding: ui.Padding{Right: basketRight, Bottom: above + basketAbove}}
	context.Overlay("basket", options, func() {
		screens.basketBar(context, count)
	})
}

// Measures of the floating basket, in pixels, as in the author's design.
const (
	basketRight       = 32
	basketAbove       = 22
	basketPadding     = 10
	basketPaddingLeft = 24
	basketGap         = 14
	basketIconSize    = 22
	basketTextSize    = 19
	basketTotalSize   = 22
	basketButtonX     = 26
	basketButtonY     = 14
	fadeHeight        = 120
	fadeSolid         = 0.6
)

func (screens *Screens) basketBar(context *ui.Context, count int) {
	theme := context.Theme
	light := theme.OnPrimary
	normal := ui.LettersLineHeight
	padding := ui.Padding{Left: basketPaddingLeft, Right: basketPadding, Top: basketPadding, Bottom: basketPadding}
	context.Frame(ui.Layout{Padding: padding, Gap: basketGap, AlignY: ui.Center}, ui.Look{Fill: theme.Primary, Radius: roundRadius, Shadow: true}, func() {
		context.IconSized("shopping-bag", basketIconSize, light)
		context.Text(piecesWords(count), ui.TextStyle{Size: basketTextSize, Weight: draw2d.Medium, Color: light, LineHeight: normal})
		context.Text(money.Format(screens.basket.Total()), ui.TextStyle{Size: basketTotalSize, Weight: draw2d.Bold, Color: light, LineHeight: normal})
		pill := ui.Look{Fill: light, Radius: roundRadius}
		button := ui.Layout{Padding: ui.PadXY(basketButtonX, basketButtonY), AlignY: ui.Center}
		if context.Clickable("Cobrar", button, ui.Looks{Normal: pill, Hovered: pill, Pressed: pill}, func() {
			context.Text("Cobrar", ui.TextStyle{Size: basketTextSize, Weight: draw2d.Bold, Color: theme.Primary, LineHeight: normal})
		}) {
			screens.step = choosingMethod
		}
	})
}

func piecesWords(count int) string {
	switch count {
	case 1:
		return "1 artículo"
	}
	return fmt.Sprintf("%d artículos", count)
}

// ChooseTodaysFair opens the stand of an open fair happening today; false when there is none.
func (screens *Screens) ChooseTodaysFair() bool {
	screens.Reload()
	today := calendar.Of(screens.now())
	for _, fair := range screens.fairs {
		if today.Within(fair.StartDate, fair.EndDate) {
			screens.chooseFair(fair)
			return true
		}
	}
	return false
}

// ShowBasket puts the first item shown twice and the second once in the basket, for captures of
// the stand.
func (screens *Screens) ShowBasket() {
	shown := screens.shown(screens.tags())
	for index, times := range []int{2, 1} {
		for range times {
			if index < len(shown) {
				screens.basket.Add(shown[index])
			}
		}
	}
}

// ShowPaymentChoice opens the choice of cash or card over ShowBasket's basket, for captures.
func (screens *Screens) ShowPaymentChoice() {
	screens.ShowBasket()
	screens.step = choosingMethod
}

// ShowCashPayment opens the cash dialog over ShowBasket's basket with the pieces given handed
// over, for captures of the stand.
func (screens *Screens) ShowCashPayment(given ...money.Money) {
	screens.ShowBasket()
	screens.received.Clear()
	for _, piece := range given {
		screens.received.Add(piece)
	}
	screens.step = payingCash
}

// ShowCardPayment opens the card dialog over ShowBasket's basket, for captures.
func (screens *Screens) ShowCardPayment() {
	screens.ShowBasket()
	screens.step = payingCard
}

// ShowPaid sells ShowBasket's basket for the given cash and opens the dialog that closes the sale,
// for captures.
func (screens *Screens) ShowPaid(given money.Money) {
	screens.ShowBasket()
	screens.received.Clear()
	screens.received.Add(given)
	change, _ := cash.Change(screens.basket.Total(), screens.received)
	screens.sell(sales.Cash, change)
}

// Selling tells whether a stand is open, rather than the choice of stands.
func (screens *Screens) Selling() bool { return screens.source != nil }
