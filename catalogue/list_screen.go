package catalogue

import (
	"fmt"
	"slices"
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/look"
)

// Measures of the list, in pixels, as in the author's design.
const (
	cardMinimumWidth     = 300
	cardsOnAPhone        = 1
	cardsWhenThereIsRoom = 4
	gridGap              = 18
	shownRestockNames    = 3
	kindBarSpace         = 20
	kindBarSpaceBelow    = 4
	bannerSpaceAbove     = 12
	bannerSpaceBelow     = 16
	tagBarSpaceBelow     = 18
	bannerRadius         = 16
	bannerPaddingX       = 18
	bannerPaddingY       = 14
	bannerGap            = 16
	bannerIconSide       = 44
	bannerIconSize       = 22
	bannerTitleSize      = 17
)

// kindTabs are the two halves of the catalogue, in the order of their tabs.
var kindTabs = []Kind{Physical, Digital}

func (screens *Screens) drawList(context *ui.Context) {
	screens.applyTabs()
	screens.query.Text = screens.search.Text
	shown := Filter(screens.products, screens.query)
	look.PageHeader(context, "Mis productos", screens.subtitle(), func() {
		look.SearchField(context, &screens.search, "Buscar por nombre o SKU")
		look.SearchButton(context, &screens.search)
		if look.MainButton(context, "plus", "Nuevo producto") {
			screens.choosingKind = true
		}
	})
	switch {
	case screens.loadErr != nil:
		look.Empty(context, "triangle-alert", "No se pudieron leer tus productos", screens.loadErr.Error(), "")
	case len(screens.products) == 0:
		if look.Empty(context, "palette", "Aún no tienes productos",
			"Añade tu primera lámina, pegatina o comisión y empieza a llevar la cuenta.", "Crear tu primer producto") {
			screens.choosingKind = true
		}
	default:
		look.Space(context, kindBarSpace)
		screens.kindBar(context)
		look.Space(context, kindBarSpaceBelow)
		if screens.query.Kind == Digital {
			screens.digitalList(context, shown)
			break
		}
		screens.restockBanner(context)
		look.Space(context, bannerSpaceBelow)
		screens.tagBar(context)
		look.Space(context, tagBarSpaceBelow)
		screens.cards(context, shown)
	}
	if screens.choosingKind {
		screens.kindSheet(context)
	}
	if screens.stockOf != "" {
		screens.stockSheet(context)
	}
}

// subtitle counts the artist's works: every product that is not archived.
func (screens *Screens) subtitle() string {
	count := len(Filter(screens.products, Query{}))
	if count == 1 {
		return "1 obra"
	}
	return fmt.Sprintf("%d obras", count)
}

// applyTabs turns the tabs the artist chose into the query.
func (screens *Screens) applyTabs() {
	screens.query.Kind = kindTabs[screens.kindTab]
	tags := screens.tagsOfKind()
	screens.query.TagIDs, screens.query.Archived = nil, false
	switch {
	case screens.tagTab > len(tags):
		screens.query.Archived = true
	case screens.tagTab > 0:
		screens.query.TagIDs = []string{tags[screens.tagTab-1].ID}
	}
}

// kindIcons are the icons of the kinds' tabs.
var kindIcons = map[Kind]string{Physical: "package", Digital: "cloud"}

// kindBar is Físico and Digital, each with how many products it holds.
func (screens *Screens) kindBar(context *ui.Context) {
	tabs := make([]look.Tab, len(kindTabs))
	for index, kind := range kindTabs {
		tabs[index] = look.Tab{Icon: kindIcons[kind], Label: kindWords[kind], Count: len(Filter(screens.products, Query{Kind: kind})), HasCount: true}
	}
	if look.Segmented(context, "kind", &screens.kindTab, tabs, look.Halves) {
		screens.tagTab = 0
		screens.query.LowStockOnly = false
	}
}

// restockBanner says how many versions run low and names the first few; «Ver cuáles» shows only
// those products.
func (screens *Screens) restockBanner(context *ui.Context) {
	names := Restock(screens.products)
	if len(names) == 0 || screens.query.Kind != Physical {
		return
	}
	theme := context.Theme
	look.Space(context, bannerSpaceAbove)
	// The line takes room inside the box, as in CSS.
	layout := ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(bannerPaddingX+cardBorder, bannerPaddingY+cardBorder), Gap: bannerGap, AlignY: ui.Center}
	context.Frame(layout, ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: cardBorder, Radius: bannerRadius}, func() {
		side := ui.Fixed(bannerIconSide)
		context.Frame(ui.Layout{Width: side, Height: side, AlignX: ui.Center, AlignY: ui.Center}, ui.Look{Fill: theme.Sunken, Radius: bannerIconSide / 2}, func() {
			context.IconSized("triangle-alert", bannerIconSize, theme.Warning)
		})
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow()}, func() {
			context.Text(fmt.Sprintf("%d %s por reponer", len(names), map[bool]string{true: "producto", false: "productos"}[len(names) == 1]),
				ui.TextStyle{Size: bannerTitleSize, Weight: draw2d.Bold})
			context.Text(restockSummary(names), ui.TextStyle{Color: theme.InkSoft})
		})
		action := map[bool]string{true: "Ver todos", false: "Ver cuáles"}[screens.query.LowStockOnly]
		if look.Link(context, action) {
			screens.query.LowStockOnly = !screens.query.LowStockOnly
		}
	})
}

// restockSummary names the first few and counts the rest: "A3 de Gato, Luna y 2 más".
func restockSummary(names []string) string {
	if len(names) <= shownRestockNames {
		return strings.Join(names, ", ")
	}
	return fmt.Sprintf("%s y %d más", strings.Join(names[:shownRestockNames], ", "), len(names)-shownRestockNames)
}

// tagsOfKind are the tags some product of the chosen kind carries, the most used first.
func (screens *Screens) tagsOfKind() []Tag {
	var used []Tag
	for _, tag := range screens.tags {
		if screens.countWithTag(tag.ID) > 0 {
			used = append(used, tag)
		}
	}
	slices.SortStableFunc(used, func(first, second Tag) int {
		return screens.countWithTag(second.ID) - screens.countWithTag(first.ID)
	})
	return used
}

func (screens *Screens) countWithTag(id string) int {
	return len(Filter(screens.products, Query{Kind: kindTabs[screens.kindTab], TagIDs: []string{id}}))
}

// tagBar is Todo, one tab per tag with its count, and Archivados when something is archived.
func (screens *Screens) tagBar(context *ui.Context) {
	kind := kindTabs[screens.kindTab]
	tabs := []look.Tab{{Label: "Todo", Count: len(Filter(screens.products, Query{Kind: kind})), HasCount: true}}
	for _, tag := range screens.tagsOfKind() {
		tabs = append(tabs, look.Tab{Label: Plural(tag.Name), Count: screens.countWithTag(tag.ID), HasCount: true})
	}
	if archived := len(Filter(screens.products, Query{Kind: kind, Archived: true})); archived > 0 {
		tabs = append(tabs, look.Tab{Label: "Archivados", Count: archived, HasCount: true})
	}
	screens.tagTab = min(screens.tagTab, len(tabs)-1)
	look.PillTabs(context, "tag", &screens.tagTab, tabs, look.GroupTabs)
}

func (screens *Screens) cards(context *ui.Context, shown []Product) {
	if len(shown) == 0 {
		look.Empty(context, "search", "Nada por aquí", "Ningún producto coincide con la búsqueda.", "")
		return
	}
	width, _ := context.Size()
	columns := min(cardsWhenThereIsRoom, max(cardsOnAPhone, int(width/cardMinimumWidth)))
	context.Box(ui.Layout{Width: ui.Grow(), Columns: columns, Gap: gridGap}, func() {
		for _, product := range shown {
			context.Push(product.ID)
			screens.card(context, product, ListFill(screens.positions[product.ID]))
			context.Pop()
		}
	})
}

// Measures of a product's card, in pixels, as in the author's design.
const (
	cardRadius        = 18
	cardBorder        = 1
	cardPictureHeight = 150
	cardLetterSize    = 54
	cardLetterAlpha   = 0.6
	cardBadgeInset    = 10
	cardBadgeSize     = 13
	cardBadgePaddingX = 10
	cardBadgePaddingY = 4
	cardBadgeGap      = 6
	cardBadgeIconSize = 14
	cardPaddingX      = 16
	cardPaddingTop    = 14
	cardPaddingBottom = 16
	cardNameSize      = 18
	cardPriceSize     = 15
	cardPriceSpace    = 2
	cardPillsSpace    = 12
	versionGap        = 6
	versionTextSize   = 14
	versionPaddingX   = 11
	versionPaddingY   = 5
	versionCountSpace = 4
	roundRadius       = 999
)

// card is one product: its picture in fill with its first tag, its name, its prices and one pill
// per version with what is left. Tapping it opens the product; a version's pill changes its stock.
func (screens *Screens) card(context *ui.Context, product Product, fill draw2d.Color) {
	theme := context.Theme
	normal := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: cardBorder, Radius: cardRadius}
	hovered := normal
	hovered.Border = theme.ControlLine
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(cardBorder)}
	padding := ui.Padding{Left: cardPaddingX, Right: cardPaddingX, Top: cardPaddingTop, Bottom: cardPaddingBottom}
	clicked := context.Clickable(product.Name, layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
		screens.cardPicture(context, product, fill)
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: padding}, func() {
			look.Paragraph(context, product.Name, ui.TextStyle{Size: cardNameSize, Weight: draw2d.Bold})
			look.Space(context, cardPriceSpace)
			context.Text(PriceRangeWords(product), ui.TextStyle{Size: cardPriceSize, Color: theme.InkSoft})
			if product.Kind == Digital {
				screens.digitalFacts(context, product)
				return
			}
			look.Space(context, cardPillsSpace)
			screens.versionPills(context, product)
		})
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: product.Name})
	if clicked {
		screens.edit(product)
	}
}

// pictureShape is the top-rounded fill of a colour, kept from frame to frame.
func (screens *Screens) pictureShape(fill draw2d.Color) *look.TopRounded {
	if screens.shapes == nil {
		screens.shapes = map[draw2d.Color]*look.TopRounded{}
	}
	shape, found := screens.shapes[fill]
	if !found {
		shape = &look.TopRounded{Fill: fill, Radius: cardRadius - cardBorder}
		screens.shapes[fill] = shape
	}
	return shape
}

// cardPicture is the coloured top of a card with the product's first letter in the middle and
// its first tag in a corner.
func (screens *Screens) cardPicture(context *ui.Context, product Product, fill draw2d.Color) {
	layout := ui.Layout{Direction: ui.Stacked, Width: ui.Grow(), Height: ui.Fixed(cardPictureHeight)}
	context.Frame(layout, ui.Look{Backdrop: screens.pictureShape(fill)}, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow(), AlignX: ui.Center, AlignY: ui.Center}, func() {
			context.Text(Initial(product.Name), ui.TextStyle{Size: cardLetterSize, Weight: draw2d.Bold, Color: draw2d.Color{1, 1, 1, cardLetterAlpha}})
		})
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow(), Padding: ui.Pad(cardBadgeInset), Gap: versionGap}, func() {
			switch {
			case product.Kind == Digital:
				cardBadge(context, "cloud", "Digital")
			case len(product.Tags) > 0:
				cardBadge(context, "", product.Tags[0].Name)
			}
			if product.Status == Draft {
				cardBadge(context, "", "Borrador")
			}
		})
	})
}

// cardBadge is a small soft label on a card's picture, with an icon before it when icon is set.
func cardBadge(context *ui.Context, icon, words string) {
	theme := context.Theme
	layout := ui.Layout{Padding: ui.PadXY(cardBadgePaddingX, cardBadgePaddingY), Gap: cardBadgeGap, AlignY: ui.Center}
	context.Frame(layout, ui.Look{Fill: theme.Surface, Radius: roundRadius}, func() {
		if icon != "" {
			context.IconSized(icon, cardBadgeIconSize, theme.InkSoft)
		}
		context.Text(words, ui.TextStyle{Size: cardBadgeSize, Color: theme.InkSoft})
	})
}

// versionPills shows each version with what is left, the count warm when it runs low and the
// words red when sold out; a digital product is made to order instead.
func (screens *Screens) versionPills(context *ui.Context, product Product) {
	theme := context.Theme
	context.Box(ui.Layout{Width: ui.Grow(), Gap: versionGap, Wrap: true}, func() {
		if product.Kind == Digital {
			look.Pill(context, "Por encargo", look.Neutral(theme))
			return
		}
		for _, variant := range product.Variants {
			context.Push(variant.ID)
			if versionPill(context, variant) {
				screens.stockOf = product.ID
			}
			context.Pop()
		}
	})
}

// versionPill is "A4 9", "Única 8" or "A3 Agotado", as a button.
func versionPill(context *ui.Context, variant Variant) bool {
	theme := context.Theme
	name := nameOr(variant.Name, "Única")
	ink, countInk, weight := theme.Ink, theme.InkSoft, draw2d.Regular
	count := fmt.Sprint(variant.Stock)
	switch {
	case variant.Stock == 0:
		ink, countInk, count = theme.Danger, theme.Danger, "Agotado"
	case variant.LowStock():
		countInk, weight = theme.Warning, draw2d.Bold
	}
	normal := ui.Look{Fill: theme.Sunken, Radius: roundRadius}
	pressed := normal
	pressed.Border, pressed.BorderWidth = theme.ControlLine, cardBorder
	layout := ui.Layout{Padding: ui.PadXY(versionPaddingX, versionPaddingY), Gap: versionCountSpace, AlignY: ui.Center}
	return context.Clickable(name, layout, ui.Looks{Normal: normal, Hovered: pressed, Pressed: pressed}, func() {
		context.Text(name, ui.TextStyle{Size: versionTextSize, Color: ink})
		context.Text(count, ui.TextStyle{Size: versionTextSize, Color: countInk, Weight: weight})
	})
}

// Measures of the digital products' list, in pixels, as in the author's design.
const (
	digitalColumns    = 3
	digitalHintSpace  = 6
	digitalGridSpace  = 20
	digitalHintSize   = 15
	factsSpace        = 14
	factsGap          = 10
	factRadius        = 12
	factPaddingX      = 12
	factPaddingY      = 10
	factLabelSize     = 13
	factValueSize     = 16
	digitalSoldSpace  = 6
	digitalPhoneLimit = 2
)

// digitalList is the digital products: nothing to count, so a line saying so and a card each.
func (screens *Screens) digitalList(context *ui.Context, shown []Product) {
	theme := context.Theme
	look.Space(context, digitalHintSpace)
	context.Text("Sin stock que contar: se venden tantas veces como quieras.", ui.TextStyle{Size: digitalHintSize, Color: theme.InkSoft})
	look.Space(context, digitalGridSpace)
	if len(shown) == 0 {
		look.Empty(context, "search", "Nada por aquí", "Ningún producto coincide con la búsqueda.", "")
		return
	}
	width, _ := context.Size()
	columns := min(digitalColumns, max(cardsOnAPhone, int(width/cardMinimumWidth)))
	context.Box(ui.Layout{Width: ui.Grow(), Columns: columns, Gap: gridGap}, func() {
		for _, product := range shown {
			context.Push(product.ID)
			screens.card(context, product, ListFill(screens.positions[product.ID]))
			context.Pop()
		}
	})
}

// digitalFacts are a digital product's format and delivery side by side, and how many times it
// sold this month.
func (screens *Screens) digitalFacts(context *ui.Context, product Product) {
	theme := context.Theme
	look.Space(context, factsSpace)
	context.Box(ui.Layout{Width: ui.Grow(), Columns: 2, Gap: factsGap}, func() {
		fact(context, "Formato", nameOr(product.DeliveryFormat, "—"))
		fact(context, "Entrega", DeliveryWords(product.TurnaroundDays))
	})
	look.Space(context, digitalSoldSpace)
	context.Text(SoldWords(screens.soldOf(product)), ui.TextStyle{Size: digitalHintSize, Color: theme.InkSoft})
}

// fact is a label and its value in a sunken well.
func fact(context *ui.Context, label, value string) {
	theme := context.Theme
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.PadXY(factPaddingX, factPaddingY)}
	context.Frame(layout, ui.Look{Fill: theme.Sunken, Radius: factRadius}, func() {
		context.Text(label, ui.TextStyle{Size: factLabelSize, Color: theme.InkSoft})
		context.Text(value, ui.TextStyle{Size: factValueSize, Weight: draw2d.Bold})
	})
}

func kindIcon(kind Kind) string {
	if kind == Digital {
		return "brush"
	}
	return "package"
}

func (screens *Screens) kindSheet(context *ui.Context) {
	theme := context.Theme
	if context.Sheet("¿Qué vas a crear?", ui.SheetOptions{}, func() {
		for _, option := range []struct {
			kind               Kind
			title, description string
		}{
			{Physical, "Producto físico", "Láminas, pegatinas, merch… Con stock y envío."},
			{Digital, "Producto digital", "Comisiones y encargos: busto, cuerpo entero…"},
		} {
			layout := ui.Layout{Width: ui.Grow(), Padding: ui.Pad(theme.Padding), Gap: 14, AlignY: ui.Center}
			if look.ClickableCard(context, option.title, layout, func() {
				context.IconSized(kindIcon(option.kind), theme.IconSize*1.5, theme.Primary)
				context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, func() {
					context.Text(option.title, ui.TextStyle{Weight: draw2d.Bold})
					look.Paragraph(context, option.description, ui.TextStyle{Color: theme.InkSoft})
				})
			}) {
				screens.choosingKind = false
				editing := newEditor(newForm(option.kind, screens.newID))
				screens.editor = &editing
			}
		}
	}) {
		screens.choosingKind = false
	}
}

func (screens *Screens) stockSheet(context *ui.Context) {
	product, found := screens.product(screens.stockOf)
	if !found {
		screens.stockOf = ""
		return
	}
	theme := context.Theme
	if context.Sheet(product.Name, ui.SheetOptions{}, func() {
		context.Text(StockWords(product), ui.TextStyle{Color: theme.InkSoft})
		for _, variant := range product.Variants {
			context.Push(variant.ID)
			context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.Center}, func() {
				context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow()}, func() {
					context.Text(nameOr(variant.Name, "Unidades"), ui.TextStyle{Weight: draw2d.Bold})
					if variant.MinimumStock > 0 {
						context.Text(fmt.Sprintf("Mínimo %d", variant.MinimumStock), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
					}
				})
				if delta := stepper(context, variant.Stock); delta != 0 {
					screens.tell(screens.store.AdjustStock(variant.ID, delta), "Stock cambiado")
				}
			})
			context.Pop()
		}
		if context.PrimaryButton("Listo") {
			screens.stockOf = ""
		}
	}) {
		screens.stockOf = ""
	}
}

// stepper is − count +; it returns the step clicked in this frame, or zero.
func stepper(context *ui.Context, count int) int {
	step := 0
	context.Box(ui.Layout{Gap: 8, AlignY: ui.Center}, func() {
		if context.IconButton("minus", "##Quitar uno") && count > 0 {
			step = -1
		}
		context.Text(fmt.Sprint(count), ui.TextStyle{Weight: draw2d.Bold})
		if context.IconButton("plus", "##Añadir uno") {
			step = 1
		}
	})
	return step
}

func nameOr(name, otherwise string) string {
	if name == "" {
		return otherwise
	}
	return name
}

func (screens *Screens) edit(product Product) {
	editing := newEditor(formOf(product))
	screens.editor = &editing
}
