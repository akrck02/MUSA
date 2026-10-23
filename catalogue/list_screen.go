package catalogue

import (
	"fmt"
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/look"
)

// Measures of the list, in pixels.
const (
	cardMinimumWidth     = 300
	cardPictureHeight    = 150
	cardLetterSize       = 56
	shownRestockNames    = 3
	kindTabsWidth        = 340
	searchFieldWidth     = 320
	cardsOnAPhone        = 1
	cardsWhenThereIsRoom = 4
	cardNameScale        = 1.15
	versionPillHeight    = 30
	versionPillTextScale = 1.15
)

// kindTabs are the two halves of the catalogue, in the order of their tabs.
var kindTabs = []Kind{Physical, Digital}

func (screens *Screens) drawList(context *ui.Context) {
	screens.applyTabs()
	shown := Filter(screens.products, screens.query)
	look.PageHeader(context, "Mis productos", screens.subtitle(), func() {
		if screens.searching {
			context.Box(ui.Layout{Width: ui.Fixed(searchFieldWidth)}, func() {
				context.TextFieldWith("search", &screens.query.Text, ui.TextFieldOptions{Placeholder: "Buscar por nombre o SKU"})
			})
		}
		if context.IconButton("search", "##Buscar") {
			screens.searching = !screens.searching
			screens.query.Text = ""
		}
		if context.PrimaryIconButton("plus", "Nuevo producto") {
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
		screens.kindBar(context)
		screens.restockBanner(context)
		screens.tagBar(context)
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
	if look.Segmented(context, "kind", &screens.kindTab, tabs) {
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
	look.Card(context, ui.Layout{Width: ui.Grow(), Padding: ui.Pad(theme.Padding), Gap: 14, AlignY: ui.Center}, func() {
		look.Badge(context, "triangle-alert", theme.IconSize*1.6)
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, func() {
			context.Text(fmt.Sprintf("%d %s por reponer", len(names), map[bool]string{true: "producto", false: "productos"}[len(names) == 1]),
				ui.TextStyle{Weight: draw2d.Bold})
			context.Text(restockSummary(names), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
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

// tagsOfKind are the tags some product of the chosen kind carries.
func (screens *Screens) tagsOfKind() []Tag {
	var used []Tag
	for _, tag := range screens.tags {
		if screens.countWithTag(tag.ID) > 0 {
			used = append(used, tag)
		}
	}
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
		tabs = append(tabs, look.Tab{Label: tag.Name, Count: screens.countWithTag(tag.ID), HasCount: true})
	}
	if archived := len(Filter(screens.products, Query{Kind: kind, Archived: true})); archived > 0 {
		tabs = append(tabs, look.Tab{Label: "Archivados", Count: archived, HasCount: true})
	}
	screens.tagTab = min(screens.tagTab, len(tabs)-1)
	look.PillTabs(context, "tag", &screens.tagTab, tabs)
}

func (screens *Screens) cards(context *ui.Context, shown []Product) {
	theme := context.Theme
	if len(shown) == 0 {
		look.Empty(context, "search", "Nada por aquí", "Ningún producto coincide con la búsqueda.", "")
		return
	}
	width, _ := context.Size()
	columns := min(cardsWhenThereIsRoom, max(cardsOnAPhone, int(width/cardMinimumWidth)))
	context.Box(ui.Layout{Width: ui.Grow(), Columns: columns, Gap: theme.Gap}, func() {
		for _, product := range shown {
			context.Push(product.ID)
			screens.card(context, product)
			context.Pop()
		}
	})
}

// card is one product: its picture with its first tag, its name, its prices and one pill per
// version with what is left. Tapping it opens the product; a version's pill changes its stock.
func (screens *Screens) card(context *ui.Context, product Product) {
	theme := context.Theme
	normal := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: theme.PanelRadius}
	hovered := normal
	hovered.Border = theme.ControlLine
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(1)}
	if context.Clickable(product.Name, layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
		screens.cardPicture(context, product)
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(16), Gap: 4}, func() {
			look.Paragraph(context, product.Name, ui.TextStyle{Size: theme.TextSize * cardNameScale, Weight: draw2d.Bold})
			context.Text(PriceRangeWords(product), ui.TextStyle{Color: theme.InkSoft})
			context.Box(ui.Layout{Height: ui.Fixed(6)}, nil)
			screens.versionPills(context, product)
		})
	}) {
		screens.edit(product)
	}
}

// pictureShape is the top-rounded fill of a colour, kept from frame to frame.
func (screens *Screens) pictureShape(theme ui.Theme, color TagColor) *look.TopRounded {
	if screens.shapes == nil {
		screens.shapes = map[TagColor]*look.TopRounded{}
	}
	shape, found := screens.shapes[color]
	if !found {
		shape = &look.TopRounded{}
		screens.shapes[color] = shape
	}
	shape.Fill, shape.Radius = TagTone(theme, color).Fill, theme.PanelRadius-1
	return shape
}

// cardPicture is the coloured top of a card with the product's first letter and first tag.
func (screens *Screens) cardPicture(context *ui.Context, product Product) {
	theme := context.Theme
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Fixed(cardPictureHeight), Padding: ui.Pad(10)}
	context.Frame(layout, ui.Look{Backdrop: screens.pictureShape(theme, PictureColor(product))}, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 6}, func() {
			if len(product.Tags) > 0 {
				look.Pill(context, product.Tags[0].Name, look.Tone{Fill: theme.Surface, Ink: theme.Ink})
			}
			if product.Status == Draft {
				look.Pill(context, "Borrador", look.Tone{Fill: theme.Surface, Ink: theme.InkSoft})
			}
		})
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow(), AlignX: ui.Center, AlignY: ui.Center}, func() {
			context.Text(Initial(product.Name), ui.TextStyle{Size: cardLetterSize, Weight: draw2d.Bold, Color: theme.Surface.WithAlpha(0.7)})
		})
	})
}

// versionPills shows each version with what is left, the count warm when it runs low and the
// whole pill red when sold out; a digital product is made to order instead.
func (screens *Screens) versionPills(context *ui.Context, product Product) {
	theme := context.Theme
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, Wrap: true}, func() {
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
	fill, ink, countInk, weight := theme.Sunken, theme.Ink, theme.InkSoft, draw2d.Regular
	count := fmt.Sprint(variant.Stock)
	switch {
	case variant.Stock == 0:
		fill, ink, countInk, count = theme.Danger.WithAlpha(0.1), theme.Danger, theme.Danger, "Agotado"
	case variant.LowStock():
		countInk, weight = theme.Warning, draw2d.Bold
	}
	normal := ui.Look{Fill: fill, Radius: versionPillHeight / 2}
	pressed := normal
	pressed.Border, pressed.BorderWidth = theme.ControlLine, 1
	layout := ui.Layout{Height: ui.Fixed(versionPillHeight), Padding: ui.PadXY(10, 0), Gap: 6, AlignY: ui.Center}
	return context.Clickable(name, layout, ui.Looks{Normal: normal, Hovered: pressed, Pressed: pressed}, func() {
		context.Text(name, ui.TextStyle{Size: theme.SmallTextSize * versionPillTextScale, Color: ink})
		context.Text(count, ui.TextStyle{Size: theme.SmallTextSize * versionPillTextScale, Color: countInk, Weight: weight})
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
