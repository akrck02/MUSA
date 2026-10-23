package catalogue

import (
	"strconv"
	"strings"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
)

// editor is the product editor: the form and what the last save attempt found wrong.
type editor struct {
	form        form
	problems    []Problem
	unreadable  map[string]string
	asking      string // "delete" while the delete question is open
	detailsOpen bool   // «Más detalles» is unfolded
}

func newEditor(edited form) editor { return editor{form: edited, unreadable: map[string]string{}} }

// problemFor is the words to show beside a field: unreadable text first, then the draft's problems.
func (editing *editor) problemFor(code Code, variantID, field string) string {
	if words, found := editing.unreadable[fieldKey(variantID, field)]; found {
		return words
	}
	for _, problem := range editing.problems {
		if problem.Code == code && problem.VariantID == variantID {
			return ProblemWords(problem)
		}
	}
	return ""
}

// Measures of the editor, in pixels.
const (
	editorSideWidth   = 400
	moneyFieldWidth   = 130
	stockColumnWidth  = 230
	sectionTitleScale = 1.4
)

func (screens *Screens) drawEditor(context *ui.Context) {
	edited := &screens.editor.form
	title := "Nuevo producto"
	if !edited.isNew {
		title = edited.product.Name
	}
	look.PageHeader(context, title, "", func() {
		if context.FlatIconButton("arrow-left", "Volver") {
			screens.editor = nil
		}
		if !edited.isNew {
			screens.editorMenu(context)
		}
		if context.PrimaryButton("Guardar") {
			screens.save()
		}
	})
	if screens.editor == nil {
		return
	}
	side := ui.Fixed(editorSideWidth)
	if look.Narrow(context) {
		side = ui.Grow()
	}
	context.Box(ui.Layout{Direction: look.Across(context), Width: ui.Grow(), Gap: context.Theme.Gap * 1.5, AlignY: ui.Start}, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: side, Gap: context.Theme.Gap}, func() {
			screens.basics(context)
			screens.tagSection(context)
		})
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: context.Theme.Gap}, func() {
			screens.variantSection(context)
			screens.moreDetails(context)
		})
	})
	if screens.editor.asking == "delete" {
		screens.askDelete(context)
	}
}

func (screens *Screens) editorMenu(context *ui.Context) {
	edited := &screens.editor.form
	if context.Button("Duplicar") {
		id, err := screens.store.Duplicate(edited.product.ID, edited.product.Name+" (copia)")
		screens.tell(err, "Copia creada como borrador")
		if copied, found := screens.product(id); err == nil && found {
			screens.edit(copied)
		}
	}
	context.MenuButton("ellipsis", "##Más acciones", func() {
		archived := edited.product.Status == Archived
		if context.MenuItem("archive", map[bool]string{true: "Restaurar", false: "Archivar"}[archived], "") {
			screens.tell(screens.store.SetArchived(edited.product.ID, !archived), map[bool]string{true: "Restaurado", false: "Archivado"}[archived])
			screens.editor = nil
		}
		if context.MenuItem("trash-2", "Eliminar", "") {
			screens.editor.asking = "delete"
		}
	})
}

func (screens *Screens) askDelete(context *ui.Context) {
	edited := &screens.editor.form
	if context.Sheet("¿Eliminar «"+edited.product.Name+"»?", ui.SheetOptions{}, func() {
		look.Paragraph(context, "Dejará de aparecer en el catálogo. Las ventas ya hechas lo conservan.", ui.TextStyle{})
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignX: ui.End}, func() {
			if context.Button("Cancelar") {
				screens.editor.asking = ""
			}
			if context.PrimaryButton("Eliminar") {
				screens.tell(screens.store.Delete(edited.product.ID), "Producto eliminado")
				screens.editor = nil
			}
		})
	}) && screens.editor != nil {
		screens.editor.asking = ""
	}
}

func (screens *Screens) save() {
	editing := screens.editor
	draft, unreadable := editing.form.draft()
	editing.unreadable = unreadable
	if len(unreadable) > 0 {
		editing.problems = nil
		return
	}
	product, problems, err := screens.store.Save(draft)
	editing.problems = problems
	if err != nil || len(problems) > 0 {
		if err != nil {
			screens.tell(err, "")
		}
		return
	}
	screens.tell(nil, "«"+product.Name+"» guardado")
	screens.editor = nil
}

// basics is the picture, the name and the kind, which never changes once made.
func (screens *Screens) basics(context *ui.Context) {
	editing := screens.editor
	edited := &editing.form
	theme := context.Theme
	shown := edited.product
	shown.Name = edited.name
	for _, name := range edited.tags[:min(1, len(edited.tags))] {
		shown.Tags = []Tag{{Name: name, Color: TagColorFor(TagKey(name))}}
	}
	editorPicture(context, shown)
	look.Field(context, "Nombre", "name", &edited.name, "Lámina «Gato astronauta»", editing.problemFor(CodeBlankName, "", ""))
	context.Box(ui.Layout{Gap: 6, AlignY: ui.Center}, func() {
		context.IconSized("lock", theme.SmallTextSize, theme.InkSoft)
		context.Text("Producto "+strings.ToLower(kindWords[shown.Kind])+" · el tipo no se cambia",
			ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
	})
}

// editorPicture is the product's picture, square, with its first letter until it has a photo.
func editorPicture(context *ui.Context, product Product) {
	theme := context.Theme
	tone := TagTone(theme, PictureColor(product))
	layout := ui.Layout{Width: ui.Grow(), Height: ui.Fixed(editorSideWidth * 0.75), AlignX: ui.Center, AlignY: ui.Center}
	context.Frame(layout, ui.Look{Fill: tone.Fill, Radius: theme.PanelRadius}, func() {
		context.Text(Initial(nameOr(product.Name, "?")), ui.TextStyle{Size: cardLetterSize * 2, Weight: draw2d.Bold, Color: theme.Surface.WithAlpha(0.7)})
	})
	// TODO(akrck02): a photo of the work, chosen from the device's pictures (proposal 3, R12).
}

func (screens *Screens) tagSection(context *ui.Context) {
	edited := &screens.editor.form
	theme := context.Theme
	look.Label(context, "Etiquetas")
	func() {
		if len(edited.tags) > 0 {
			context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, Wrap: true}, func() {
				for index, name := range edited.tags {
					context.Push(name)
					tone := TagTone(theme, TagColorFor(TagKey(name)))
					if _, removed := context.RemovableChip("tag", name, chipLook(tone)); removed {
						edited.tags = append(edited.tags[:index:index], edited.tags[index+1:]...)
					}
					context.Pop()
				}
			})
		}
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.Center}, func() {
			context.TextFieldWith("tag", &edited.tagTyped, ui.TextFieldOptions{Placeholder: "Añade una etiqueta: tema, fandom, evento…"})
			if context.IconButton("plus", "Añadir") {
				edited.addTag(edited.tagTyped)
			}
		})
		suggestions := SuggestTags(screens.tags, edited.tags, edited.tagTyped)
		if len(suggestions) > 0 {
			context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, Wrap: true}, func() {
				for _, tag := range suggestions {
					context.Push("suggested-" + tag.ID)
					if context.Chip("plus", tag.Name, chipLook(look.Neutral(theme))) {
						edited.addTag(tag.Name)
					}
					context.Pop()
				}
			})
		}
	}()
}

func (edited *form) addTag(name string) {
	edited.tags = DistinctTagNames(append(edited.tags, name))
	edited.tagTyped = ""
}

// variantSection is the table of versions: each with its name, price, cost and stock.
func (screens *Screens) variantSection(context *ui.Context) {
	editing := screens.editor
	edited := &editing.form
	theme := context.Theme
	look.Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(theme.Padding), Gap: theme.Gap}, func() {
		context.Text("Versiones", ui.TextStyle{Size: theme.TextSize * sectionTitleScale, Weight: draw2d.Bold})
		look.Paragraph(context, "Cada versión tiene su precio y su stock. Con una sola, no hace falta nombre.", ui.TextStyle{Color: theme.InkSoft})
		if words := editing.problemFor(CodeNoVariants, "", ""); words != "" {
			context.Text(words, ui.TextStyle{Color: theme.Danger})
		}
		for index := range edited.variants {
			context.Push(edited.variants[index].id)
			removed := screens.variantRow(context, &edited.variants[index])
			context.Pop()
			if removed {
				edited.variants = append(edited.variants[:index:index], edited.variants[index+1:]...)
				break
			}
		}
		if context.FlatIconButton("plus", "Añadir versión") {
			edited.variants = append(edited.variants, variantForm{id: screens.newID(), stock: "0"})
		}
	})
}

// variantRow is one version's line; it returns true on the frame it is removed.
func (screens *Screens) variantRow(context *ui.Context, typed *variantForm) bool {
	editing := screens.editor
	id := typed.id
	removed := false
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 12, AlignY: ui.End}, func() {
		look.Field(context, "Versión", "variant-name", &typed.name, "A4, Mate, Busto…",
			firstOf(editing.problemFor(CodeBlankVariantName, id, ""), editing.problemFor(CodeDuplicateVariantName, id, "")))
		context.Box(ui.Layout{Width: ui.Fixed(moneyFieldWidth)}, func() {
			look.Field(context, "Precio (€)", "price", &typed.price, "12,50", editing.problemFor(CodeNegativePrice, id, "price"))
		})
		context.Box(ui.Layout{Width: ui.Fixed(moneyFieldWidth)}, func() {
			look.Field(context, "Te cuesta (€)", "cost", &typed.cost, "3,50", editing.problemFor(CodeNegativeCost, id, "cost"))
		})
		if editing.form.product.Kind == Physical {
			stockColumn(context, typed)
		}
		removed = len(editing.form.variants) > 1 && context.FlatIconButton("trash-2", "##Quitar versión")
	})
	return removed
}

// stockColumn is − count + for a version's stock, with «Pocos» when it reaches the warning.
func stockColumn(context *ui.Context, typed *variantForm) {
	theme := context.Theme
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(stockColumnWidth), Gap: 4}, func() {
		context.Text("En stock", ui.TextStyle{Size: theme.SmallTextSize, Weight: draw2d.Bold})
		count, err := strconv.Atoi(typed.stock)
		if err != nil {
			context.TextFieldWith("stock", &typed.stock, ui.TextFieldOptions{Placeholder: "0"})
			return
		}
		context.Box(ui.Layout{Gap: 8, AlignY: ui.Center}, func() {
			count += stepper(context, count)
			typed.stock = strconv.Itoa(count)
			minimum, _ := strconv.Atoi(typed.minimum) // an unreadable minimum is told when saving
			if (Variant{Stock: count, MinimumStock: minimum}).LowStock() {
				context.Text(map[bool]string{true: "Agotado", false: "Pocos"}[count == 0], ui.TextStyle{Size: theme.SmallTextSize, Weight: draw2d.Bold, Color: theme.Warning})
			}
		})
	})
}

// moreDetails folds away what is set once: the description, the status, what it is made of or how
// it is delivered, and each version's SKU, warning and measures.
func (screens *Screens) moreDetails(context *ui.Context) {
	editing := screens.editor
	theme := context.Theme
	look.Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(theme.Padding), Gap: theme.Gap}, func() {
		layout := ui.Layout{Width: ui.Grow(), AlignY: ui.Center}
		if context.Clickable("Más detalles", layout, ui.Looks{}, func() {
			context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, func() {
				context.Text("Más detalles", ui.TextStyle{})
				context.Text(screens.detailsSummary(), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
			})
			context.IconSized(map[bool]string{true: "chevron-up", false: "chevron-down"}[editing.detailsOpen], theme.IconSize, theme.InkSoft)
		}) {
			editing.detailsOpen = !editing.detailsOpen
		}
		if editing.detailsOpen {
			screens.details(context)
		}
	})
}

func (screens *Screens) detailsSummary() string {
	if screens.editor.form.product.Kind == Physical {
		return "Descripción, estado, SKU, medidas, peso y aviso de stock bajo"
	}
	return "Descripción, estado, formato y plazo de entrega"
}

func (screens *Screens) details(context *ui.Context) {
	editing := screens.editor
	edited := &editing.form
	theme := context.Theme
	look.Field(context, "Descripción", "description", &edited.description, "Opcional", "")
	look.Label(context, "Estado")
	context.Tabs("status", &edited.status, statusWords[Draft], statusWords[Active], statusWords[Archived])
	if edited.product.Kind == Digital {
		context.Box(ui.Layout{Width: ui.Grow(), Columns: look.Columns(context, 2), Gap: theme.Gap}, func() {
			look.Field(context, "Formato", "delivery", &edited.delivery, "PNG 4000 px", "")
			look.Field(context, "Plazo en días", "turnaround", &edited.turnaround, "14",
				firstOf(editing.unreadable["turnaround"], editing.problemFor(CodeNegativeTurnaround, "", "")))
		})
	} else {
		look.Field(context, "Material", "material", &edited.material, "Papel de algodón 300 g", "")
	}
	for index := range edited.variants {
		typed := &edited.variants[index]
		context.Push(typed.id)
		context.Text(nameOr(typed.name, "Única"), ui.TextStyle{Weight: draw2d.Bold})
		marginWords(context, typed)
		if edited.product.Kind == Physical {
			screens.physicalFields(context, typed)
		}
		context.Pop()
	}
}

// marginWords shows what each unit leaves, as the artist types the price and the cost.
func marginWords(context *ui.Context, typed *variantForm) {
	theme := context.Theme
	price, priceRead := money.Parse(nameOr(typed.price, "0"))
	cost, costRead := money.Parse(nameOr(typed.cost, "0"))
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 4}, func() {
		context.Text("Margen", ui.TextStyle{Size: theme.SmallTextSize, Weight: draw2d.Bold})
		if !priceRead || !costRead {
			context.Text("—", ui.TextStyle{Color: theme.InkSoft})
			return
		}
		margin := money.MarginOf(price, cost)
		words := money.Format(margin.Amount)
		if margin.HasPercent {
			words += " · " + strconv.Itoa(margin.Percent) + " %"
		}
		colour := theme.Ink
		if margin.Negative() {
			colour = theme.Danger
		}
		context.Text(words, ui.TextStyle{Weight: draw2d.Bold, Color: colour})
	})
}

func (screens *Screens) physicalFields(context *ui.Context, typed *variantForm) {
	editing := screens.editor
	theme := context.Theme
	id := typed.id
	context.Box(ui.Layout{Width: ui.Grow(), Columns: look.Columns(context, 3), Gap: theme.Gap, AlignY: ui.Start}, func() {
		look.Field(context, "SKU", "sku", &typed.sku, "LAM-GAT-A4", editing.problemFor(CodeDuplicateSKU, id, "sku"))
		look.Field(context, "Avisar con", "minimum", &typed.minimum, "Opcional", editing.problemFor(CodeNegativeMinimumStock, id, "minimum"))
	})
	context.Box(ui.Layout{Width: ui.Grow(), Columns: look.Columns(context, 4), Gap: theme.Gap, AlignY: ui.Start}, func() {
		measure := editing.problemFor(CodeInvalidMeasure, id, "")
		look.Field(context, "Peso (g)", "weight", &typed.weight, "", firstOf(editing.unreadable[fieldKey(id, "weight")], measure))
		look.Field(context, "Ancho (cm)", "width", &typed.width, "", editing.unreadable[fieldKey(id, "width")])
		look.Field(context, "Alto (cm)", "height", &typed.height, "", editing.unreadable[fieldKey(id, "height")])
		look.Field(context, "Fondo (cm)", "depth", &typed.depth, "", editing.unreadable[fieldKey(id, "depth")])
	})
}

func firstOf(words ...string) string {
	for _, each := range words {
		if each != "" {
			return each
		}
	}
	return ""
}
