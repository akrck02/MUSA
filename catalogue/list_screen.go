package catalogue

import (
	"fmt"
	"slices"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/look"
)

// Measures of the list, in pixels.
const (
	cardMinimumWidth = 340
	shownTags        = 3
	thumbnailSide    = 56
)

func (screens *Screens) drawList(context *ui.Context) {
	shown := Filter(screens.products, screens.query)
	look.PageHeader(context, "Catálogo", screens.subtitle(len(shown)), func() {
		if context.PrimaryIconButton("plus", "Nuevo producto") {
			screens.choosingKind = true
		}
	})
	switch {
	case screens.loadErr != nil:
		look.Empty(context, "triangle-alert", "No se pudo leer el catálogo", screens.loadErr.Error(), "")
	case len(screens.products) == 0:
		if look.Empty(context, "palette", "Tu catálogo está vacío",
			"Añade tu primera lámina, pegatina o comisión y empieza a llevar la cuenta.", "Crear tu primer producto") {
			screens.choosingKind = true
		}
	default:
		screens.filterBar(context)
		screens.cards(context, shown)
	}
	if screens.choosingKind {
		screens.kindSheet(context)
	}
	if screens.stockOf != "" {
		screens.stockSheet(context)
	}
}

func (screens *Screens) subtitle(count int) string {
	noun := map[bool]string{true: "producto", false: "productos"}
	if screens.query.Filtering() {
		noun = map[bool]string{true: "resultado", false: "resultados"}
	}
	if len(screens.products) == 0 {
		return "Todo lo que vendes, en un sitio"
	}
	return fmt.Sprintf("%d %s", count, noun[count == 1])
}

func (screens *Screens) filterBar(context *ui.Context) {
	query := &screens.query
	context.TextFieldWith("search", &query.Text, ui.TextFieldOptions{Placeholder: "Buscar productos o SKU"})
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, Wrap: true}, func() {
		theme := context.Theme
		toggle := func(icon, label string, on bool) bool {
			tone := look.Neutral(theme)
			if on {
				tone = look.Tone{Fill: theme.PrimaryTint, Ink: theme.Ink}
			}
			return context.Chip(icon, label, chipLook(tone))
		}
		for _, kind := range []Kind{Physical, Digital} {
			if toggle(kindIcon(kind), pluralKindWords[kind], query.Kind == kind) {
				query.Kind = map[bool]Kind{true: "", false: kind}[query.Kind == kind]
			}
		}
		if toggle("triangle-alert", "Stock bajo", query.LowStockOnly) {
			query.LowStockOnly = !query.LowStockOnly
		}
		if toggle("archive", "Archivados", query.Archived) {
			query.Archived = !query.Archived
		}
		for _, tag := range screens.tags {
			context.Push(tag.ID)
			chosen := slices.Contains(query.TagIDs, tag.ID)
			tone := TagTone(theme, tag.Color)
			if !chosen {
				tone = look.Neutral(theme)
			}
			if context.Chip("tag", tag.Name, chipLook(tone)) {
				query.TagIDs = toggled(query.TagIDs, tag.ID)
			}
			context.Pop()
		}
	})
	if len(Filter(screens.products, *query)) == 0 && context.FlatButton("Quitar filtros") {
		*query = Query{}
	}
}

func toggled(ids []string, id string) []string {
	if index := slices.Index(ids, id); index >= 0 {
		return slices.Delete(slices.Clone(ids), index, index+1)
	}
	return append(slices.Clone(ids), id)
}

func (screens *Screens) cards(context *ui.Context, shown []Product) {
	theme := context.Theme
	if len(shown) == 0 {
		look.Empty(context, "search", "Nada por aquí", "Ningún producto coincide con la búsqueda o los filtros.", "")
		return
	}
	width, _ := context.Size()
	columns := max(1, int(width/cardMinimumWidth)-1)
	context.Box(ui.Layout{Width: ui.Grow(), Columns: columns, Gap: theme.Gap}, func() {
		for _, product := range shown {
			context.Push(product.ID)
			screens.card(context, product)
			context.Pop()
		}
	})
}

func (screens *Screens) card(context *ui.Context, product Product) {
	theme := context.Theme
	layout := ui.Layout{Width: ui.Grow(), Padding: ui.Pad(12), Gap: 14}
	opened := look.ClickableCard(context, product.Name, layout, func() {
		thumbnail(context, product)
		context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 4}, ui.Look{Clip: true}, func() {
			context.Text(product.Name, ui.TextStyle{Weight: draw2d.Bold})
			context.Text(SummaryWords(product), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
			tagRow(context, product)
			context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, AlignY: ui.Center, Wrap: true}, func() {
				context.Text(PriceWords(product), ui.TextStyle{Weight: draw2d.Bold, Color: theme.Primary})
				context.Spacer()
				screens.statusPills(context, product)
			})
		})
	})
	if opened {
		screens.edit(product)
	}
}

// thumbnail is a square with the product's first letter, tinted by its kind, until products have
// pictures.
func thumbnail(context *ui.Context, product Product) {
	theme := context.Theme
	fill := theme.PrimaryTint
	if product.Kind == Digital {
		fill = theme.Sunken
	}
	side := ui.Fixed(thumbnailSide)
	context.Frame(ui.Layout{Width: side, Height: side, AlignX: ui.Center, AlignY: ui.Center}, ui.Look{Fill: fill, Radius: theme.ControlRadius}, func() {
		context.IconSized(kindIcon(product.Kind), theme.IconSize*1.2, theme.Ink)
	})
}

func tagRow(context *ui.Context, product Product) {
	if len(product.Tags) == 0 {
		return
	}
	theme := context.Theme
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 4, Wrap: true}, func() {
		for index, tag := range product.Tags {
			if index == shownTags {
				look.Pill(context, fmt.Sprintf("+%d", len(product.Tags)-shownTags), look.Neutral(theme))
				break
			}
			look.Pill(context, tag.Name, TagTone(theme, tag.Color))
		}
	})
}

func (screens *Screens) statusPills(context *ui.Context, product Product) {
	theme := context.Theme
	if product.Status == Draft {
		look.Pill(context, "Borrador", look.Neutral(theme))
	}
	if product.Kind == Digital {
		look.Pill(context, "Por encargo", look.Neutral(theme))
		return
	}
	tone := look.Positive(theme)
	switch {
	case product.TotalStock() == 0:
		tone = look.Danger(theme)
	case product.LowStock():
		tone = look.Warning(theme)
	}
	if context.Chip("package", StockWords(product), chipLook(tone)) {
		screens.stockOf = product.ID
	}
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
