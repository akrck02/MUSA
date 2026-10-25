package stand

import (
	"fmt"
	"slices"
	"strings"
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

	fairs       []events.Event
	source      Source
	items       []Item
	loadErr     error
	tag         int                               // index in tags()
	icons       map[string]string                 // each tag's icon, by its folded name
	preparing   *preparing                        // set while choosing what to take to an event
	addProducts func()                            // opens the catalogue; nil until the app gives it
	asking      []Item                            // the versions of the product whose version is asked
	search      look.Search                       // by product or version
	shapes      map[draw2d.Color]*look.TopRounded // the pictures' fills, kept so drawing them allocates nothing
	fade        *look.Fade                        // under the floating basket

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
	if screens.loadErr == nil {
		screens.loadErr = screens.readIcons()
	}
}

// readIcons reads the icon of each tag, chosen or guessed.
func (screens *Screens) readIcons() error {
	tags, err := screens.products.Tags()
	if err != nil {
		return err
	}
	screens.icons = map[string]string{}
	for _, tag := range tags {
		screens.icons[catalogue.TagKey(tag.Name)] = tag.Icon()
	}
	return nil
}

// tagIcon is a tag's icon, by its name.
func (screens *Screens) tagIcon(name string) string {
	if icon, found := screens.icons[catalogue.TagKey(name)]; found {
		return icon
	}
	return catalogue.TagIcon(name)
}

// Back closes the dialog open, or leaves the stand for the choice of stands; false when only the
// choice was open.
func (screens *Screens) Back() bool {
	switch {
	case screens.step != browsing:
		screens.step = browsing
	case screens.preparing != nil:
		screens.leavePreparing()
	case screens.source != nil:
		screens.Choose(nil)
	default:
		return false
	}
	return true
}

// AddProductsWith gives the stand what opens the catalogue, for a stand with no products at all.
func (screens *Screens) AddProductsWith(open func()) { screens.addProducts = open }

// nothingToSell says why the stand is empty and offers what fills it: stock taken to the fair, or
// products in the catalogue.
func (screens *Screens) nothingToSell(context *ui.Context) {
	fair, atAFair := screens.source.(Fair)
	if atAFair {
		if look.Empty(context, "package-open", "No hay nada para vender", "Aún no has llevado stock a este evento.", "Configurar stock") {
			back := screens.source
			screens.Prepare(fair.Event, func() { screens.Choose(back) })
		}
		return
	}
	if look.Empty(context, "package-open", "No hay nada para vender", "Aún no tienes productos físicos en el catálogo.", "Añadir productos") && screens.addProducts != nil {
		screens.addProducts()
	}
}

// BasketQuantity is how many of a version are in the basket.
func (screens *Screens) BasketQuantity(variantID string) int {
	return screens.basket.Quantity(variantID)
}

// Choose opens a stand; nil goes back to the choice of stands. The basket starts empty.
func (screens *Screens) Choose(source Source) {
	screens.source, screens.tag, screens.step, screens.preparing = source, 0, browsing, nil
	screens.basket.Clear()
	screens.Reload()
}

// Draw draws the open page of the stand.
func (screens *Screens) Draw(context *ui.Context) {
	if screens.source == nil {
		screens.drawChoice(context)
	} else {
		screens.drawRegister(context)
		screens.drawVersionQuestion(context)
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

// Measures of the choice of stands, in pixels, as in the author's design.
const (
	choiceWidth       = 976
	choiceTop         = 40 // under the page's own top
	choiceTitleSize   = 40
	choiceTitleLine   = 1.2
	choiceLeadSize    = 20
	choiceLeadAbove   = 8
	choiceLeadBelow   = 36
	heroRadius        = 24
	heroPadding       = 32
	heroGap           = 28
	heroBadgeSide     = 104
	heroIconSize      = 52
	heroBadgeAlpha    = 0.12
	heroLiveSize      = 15
	heroNameSize      = 34
	heroNameSpace     = 6
	heroWordsSize     = 18
	heroSoftAlpha     = 0.85
	heroButtonSize    = 20
	heroButtonPadX    = 30
	heroButtonPadY    = 16
	choiceColumns     = 3
	choiceGap         = 16
	choiceCardRadius  = 20
	choiceCardPadding = 26
	choiceCardGap     = 14
	choiceBadge       = 72
	choiceIconSize    = 40
	choiceNameSize    = 22
	choiceWordsSize   = 17
)

// drawChoice is where the artist picks what to sell from: the fair on today, big and dark, then the
// other open fairs by date and the whole stock.
func (screens *Screens) drawChoice(context *ui.Context) {
	theme := context.Theme
	context.Box(ui.Layout{Width: ui.Grow(), AlignX: ui.Center}, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: look.AtMost(context, choiceWidth), Padding: ui.Padding{Top: choiceTop}}, func() {
			context.Text("¿Dónde vendes hoy?", ui.TextStyle{Size: choiceTitleSize, Weight: draw2d.Bold, LineHeight: choiceTitleLine})
			look.Space(context, choiceLeadAbove)
			look.Paragraph(context, "Elige el evento y lo que vendas quedará apuntado en él.", ui.TextStyle{Size: choiceLeadSize, Color: theme.InkSoft})
			look.Space(context, choiceLeadBelow)
			if screens.loadErr != nil {
				look.Empty(context, "triangle-alert", "No se pudieron leer los eventos", screens.loadErr.Error(), "")
				return
			}
			screens.choices(context)
		})
	})
}

// choices are the fair on today, if any, and a card for each other open fair and the whole stock.
func (screens *Screens) choices(context *ui.Context) {
	today := calendar.Of(screens.now())
	others := slices.Clone(screens.fairs)
	slices.SortStableFunc(others, func(first, second events.Event) int {
		return strings.Compare(first.StartDate.String(), second.StartDate.String())
	})
	if index := slices.IndexFunc(others, func(fair events.Event) bool { return today.Within(fair.StartDate, fair.EndDate) }); index >= 0 {
		screens.heroCard(context, others[index])
		look.Space(context, choiceGap)
		others = slices.Delete(others, index, index+1)
	}
	context.Box(ui.Layout{Width: ui.Grow(), Columns: look.Columns(context, choiceColumns), Gap: choiceGap}, func() {
		for _, fair := range others {
			context.Push(fair.ID)
			if screens.choiceCard(context, "calendar", fair.Name, calendar.LongRangeWords(fair.StartDate, fair.EndDate, today.Year)) {
				screens.chooseFair(fair)
			}
			context.Pop()
		}
		if screens.choiceCard(context, "package", "Todo el stock", "Todos tus productos, para vender fuera de un evento") {
			screens.Choose(WholeStock{Sales: screens.sales, Products: screens.products, Now: screens.now})
		}
	})
}

func (screens *Screens) chooseFair(fair events.Event) {
	screens.Choose(Fair{Event: fair, Events: screens.events, Products: screens.products})
}

// heroCard is the fair on today: dark, with what the artist took and «Vender aquí». All of it
// opens the fair's stand. On a phone its parts stack, so a long name keeps its room.
func (screens *Screens) heroCard(context *ui.Context, fair events.Event) {
	colours := heroColoursOf(context.Theme)
	normal := ui.Look{Fill: colours.fill, Border: colours.border, BorderWidth: colours.borderWidth, Radius: heroRadius}
	layout := ui.Layout{Direction: look.Across(context), Width: ui.Grow(), Padding: ui.Pad(heroPadding + colours.borderWidth), Gap: heroGap, AlignY: ui.Center}
	if context.Clickable(fair.Name, layout, ui.Looks{Normal: normal, Hovered: normal, Pressed: normal}, func() {
		side := ui.Fixed(heroBadgeSide)
		badge := ui.Layout{Width: side, Height: side, AlignX: ui.Center, AlignY: ui.Center}
		context.Frame(badge, ui.Look{Fill: colours.badge, Radius: heroBadgeSide / 2}, func() {
			context.IconSized("tent", heroIconSize, colours.badgeIcon)
		})
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow()}, func() {
			context.Text("● EN CURSO", ui.TextStyle{Size: heroLiveSize, Weight: draw2d.Bold, Color: colours.soft})
			look.Space(context, heroNameSpace)
			context.Text(fair.Name, ui.TextStyle{Size: heroNameSize, Weight: draw2d.Bold, Color: colours.ink})
			look.Space(context, heroNameSpace)
			context.Text(screens.heroWords(fair), ui.TextStyle{Size: heroWordsSize, Color: colours.soft})
		})
		pill := ui.Look{Fill: colours.button, Radius: roundRadius}
		context.Frame(ui.Layout{Padding: ui.PadXY(heroButtonPadX, heroButtonPadY)}, pill, func() {
			context.Text("Vender aquí", ui.TextStyle{Size: heroButtonSize, Weight: draw2d.Bold, Color: colours.buttonInk})
		})
	}) {
		screens.chooseFair(fair)
	}
}

// heroColours are the colours of the card of today's fair: on the accent on the light theme, and
// on the surface with an accent line on the dark one, where a light accent would glare.
type heroColours struct {
	fill, border, ink, soft, badge, badgeIcon, button, buttonInk draw2d.Color
	borderWidth                                                  float32
}

// heroDarkBorder is the width of the accent line around today's fair on the dark theme.
const heroDarkBorder = 2

func heroColoursOf(theme ui.Theme) heroColours {
	if look.Dark(theme) {
		return heroColours{fill: theme.Surface, border: theme.Primary, ink: theme.Ink, soft: theme.InkSoft, badge: theme.Sunken,
			badgeIcon: theme.Primary, button: theme.Primary, buttonInk: theme.OnPrimary, borderWidth: heroDarkBorder}
	}
	light := theme.OnPrimary
	return heroColours{fill: theme.Primary, ink: light, soft: theme.Primary.Mix(light, heroSoftAlpha),
		badge: theme.Primary.Mix(light, heroBadgeAlpha), badgeIcon: light, button: light, buttonInk: theme.Primary}
}

// heroWords is how many pieces the artist has at the fair on today.
func (screens *Screens) heroWords(fair events.Event) string {
	detail, err := screens.events.Detail(fair.ID)
	if err != nil {
		return "Hoy" // the count is a nicety; the fair can still be chosen
	}
	left := 0
	for _, line := range detail.Stock {
		left += line.Remaining()
	}
	return fmt.Sprintf("Hoy · llevas %d piezas", left)
}

func (screens *Screens) choiceCard(context *ui.Context, icon, title, subtitle string) bool {
	theme := context.Theme
	normal := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: choiceCardRadius}
	hovered := normal
	hovered.Border = theme.ControlLine
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(choiceCardPadding + 1), Gap: choiceCardGap}
	return context.Clickable(title, layout, look.Answering(theme, normal, hovered), func() {
		side := ui.Fixed(choiceBadge)
		context.Frame(ui.Layout{Width: side, Height: side, AlignX: ui.Center, AlignY: ui.Center}, ui.Look{Fill: theme.Sunken, Radius: choiceBadge / 2}, func() {
			context.IconSized(icon, choiceIconSize, theme.Primary)
		})
		look.Paragraph(context, title, ui.TextStyle{Size: choiceNameSize, Weight: draw2d.Bold})
		look.Paragraph(context, subtitle, ui.TextStyle{Size: choiceWordsSize, Color: theme.InkSoft})
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
	if screens.preparing != nil {
		screens.preparingHeader(context)
	}
	if screens.source == nil {
		return // the header left the preparing
	}
	screens.top(context, tags)
	look.Space(context, topSpaceBelow)
	if len(screens.items) == 0 {
		screens.nothingToSell(context)
		return
	}
	width, _ := context.Size()
	columns := min(tilesColumns, max(2, int(width/tileWidth)))
	context.Box(ui.Layout{Width: ui.Grow(), Columns: columns, Gap: tilesGap}, func() {
		for position, versions := range ByProduct(screens.shown(tags)) {
			context.Push(versions[0].ProductID)
			screens.productTile(context, versions, catalogue.ListFill(position))
			context.Pop()
		}
	})
}

// top is where the stand sells from with «Cambiar», a tab per tag, and the search and «+», in one
// row; on a phone the tabs go on a row of their own under the rest, so they can wrap.
func (screens *Screens) top(context *ui.Context, tags []string) {
	if !look.Narrow(context) {
		context.Box(ui.Layout{Width: ui.Grow(), AlignY: ui.Center}, func() {
			screens.whereAndDivider(context)
			context.Box(ui.Layout{Width: ui.Grow()}, func() { screens.tagTabs(context, tags) })
			screens.tools(context)
		})
		return
	}
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: toolsGap}, func() {
		context.Box(ui.Layout{Width: ui.Grow(), AlignY: ui.Center}, func() {
			if screens.preparing == nil {
				screens.where(context)
			}
			context.Spacer()
			screens.tools(context)
		})
		screens.tagTabs(context, tags)
	})
}

// whereAndDivider is where the stand sells from and a fine line after it; preparing has neither.
func (screens *Screens) whereAndDivider(context *ui.Context) {
	if screens.preparing != nil {
		return
	}
	screens.where(context)
	look.Space(context, whereSpace)
	context.Frame(ui.Layout{Width: ui.Fixed(1), Height: ui.Fixed(dividerHeight)}, ui.Look{Fill: context.Theme.Line}, nil)
	look.Space(context, dividerSpace)
}

// tagTabs is a tab per tag, when the products have tags.
func (screens *Screens) tagTabs(context *ui.Context, tags []string) {
	if len(tags) == 0 {
		return
	}
	tabs := make([]look.Tab, len(tags))
	for index, tag := range tags {
		tabs[index] = look.Tab{Icon: screens.tagIcon(tag), Label: catalogue.Plural(tag)}
	}
	look.PillTabs(context, "tag", &screens.tag, tabs, look.LargeTabs)
}

// tools are the search and «Venta libre».
func (screens *Screens) tools(context *ui.Context) {
	context.Box(ui.Layout{Gap: toolsGap}, func() {
		look.SearchField(context, &screens.search, "Buscar producto")
		look.SearchButton(context, &screens.search)
		if screens.preparing == nil && look.ToolButton(context, "plus", "Venta libre") {
			screens.tell(false, "Pronto: vender algo que no está en el catálogo")
		}
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

// shown is the items of the chosen tag (all of them when there are no tags) that match the search.
func (screens *Screens) shown(tags []string) []Item {
	wanted := catalogue.Fold(strings.TrimSpace(screens.search.Text))
	var shown []Item
	for _, item := range screens.items {
		ofTag := screens.tag >= len(tags) || slices.Contains(item.Tags, tags[screens.tag])
		if ofTag && strings.Contains(catalogue.Fold(item.Label), wanted) {
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
	options := ui.LayerOptions{AlignX: ui.End, AlignY: ui.End, Padding: ui.Padding{Right: basketRight, Bottom: above + basketAbove},
		Leave: look.Leaving(look.DialogRise)}
	shown := context.Appear("basket", look.DialogMotion)
	context.Overlay("basket", options, func() {
		context.Frame(ui.Layout{}, look.Arriving(ui.Look{}, shown, look.DialogRise), func() { screens.basketBar(context, count) })
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
		icon, action := "shopping-bag", "Cobrar"
		if screens.preparing != nil {
			icon, action = "luggage", "Llevar al evento"
		}
		context.IconSized(icon, basketIconSize, light)
		context.Text(piecesWords(count), ui.TextStyle{Size: basketTextSize, Weight: draw2d.Medium, Color: light, LineHeight: normal})
		if screens.preparing == nil {
			context.Text(money.Format(screens.basket.Total()), ui.TextStyle{Size: basketTotalSize, Weight: draw2d.Bold, Color: light, LineHeight: normal})
		}
		pill := ui.Look{Fill: light, Radius: roundRadius}
		button := ui.Layout{Padding: ui.PadXY(basketButtonX, basketButtonY), AlignY: ui.Center}
		clicked := context.Clickable(action, button, ui.Looks{Normal: pill, Hovered: pill, Pressed: pill}, func() {
			context.Text(action, ui.TextStyle{Size: basketTextSize, Weight: draw2d.Bold, Color: theme.Primary, LineHeight: normal})
		})
		switch {
		case clicked && screens.preparing != nil:
			screens.takeAll()
		case clicked:
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
