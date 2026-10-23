package catalogue

import (
	"strconv"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
)

// editor is the product editor: the form and what the last save attempt found wrong.
type editor struct {
	form       form
	problems   []Problem
	unreadable map[string]string
	asking     string // "delete" while the delete question is open
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

func (screens *Screens) drawEditor(context *ui.Context) {
	editing := screens.editor
	edited := &editing.form
	title := "Nuevo producto"
	if !edited.isNew {
		title = edited.product.Name
	}
	look.PageHeader(context, title, kindWords[edited.product.Kind], func() {
		if context.FlatIconButton("arrow-left", "Volver") {
			screens.editor = nil
		}
		if !edited.isNew {
			screens.editorMenu(context)
		}
		if context.PrimaryIconButton("check", "Guardar") {
			screens.save()
		}
	})
	if screens.editor == nil {
		return
	}
	screens.basics(context)
	screens.tagSection(context)
	screens.details(context)
	screens.variantSection(context)
	if editing.asking == "delete" {
		screens.askDelete(context)
	}
}

func (screens *Screens) editorMenu(context *ui.Context) {
	edited := &screens.editor.form
	context.MenuButton("ellipsis", "##Más acciones", func() {
		if context.MenuItem("copy", "Duplicar", "") {
			id, err := screens.store.Duplicate(edited.product.ID, edited.product.Name+" (copia)")
			screens.tell(err, "Copia creada como borrador")
			if copied, found := screens.product(id); err == nil && found {
				screens.edit(copied)
			}
		}
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

func (screens *Screens) basics(context *ui.Context) {
	editing := screens.editor
	edited := &editing.form
	look.Section(context, "Lo básico", func() {
		look.Field(context, "Nombre", "name", &edited.name, "Lámina «Gato astronauta»", editing.problemFor(CodeBlankName, "", ""))
		look.Field(context, "Descripción", "description", &edited.description, "Opcional", "")
		context.Text("Estado", ui.TextStyle{Size: context.Theme.SmallTextSize, Weight: draw2d.Bold})
		context.Tabs("status", &edited.status, statusWords[Draft], statusWords[Active], statusWords[Archived])
	})
}

func (screens *Screens) tagSection(context *ui.Context) {
	edited := &screens.editor.form
	theme := context.Theme
	look.Section(context, "Etiquetas", func() {
		if len(edited.tags) > 0 {
			context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, Wrap: true}, func() {
				for index, name := range edited.tags {
					context.Push(name)
					tone := tagTone(theme, TagColorFor(TagKey(name)))
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
	})
}

func (edited *form) addTag(name string) {
	edited.tags = DistinctTagNames(append(edited.tags, name))
	edited.tagTyped = ""
}

func (screens *Screens) details(context *ui.Context) {
	editing := screens.editor
	edited := &editing.form
	if edited.product.Kind == Physical {
		look.Section(context, "Material", func() {
			look.Field(context, "De qué está hecho", "material", &edited.material, "Papel de algodón 300 g", "")
		})
		return
	}
	look.Section(context, "Entrega", func() {
		context.Box(ui.Layout{Width: ui.Grow(), Columns: look.Columns(context, 2), Gap: context.Theme.Gap}, func() {
			look.Field(context, "Formato", "delivery", &edited.delivery, "PNG 4000 px", "")
			look.Field(context, "Plazo en días", "turnaround", &edited.turnaround, "14",
				firstOf(editing.unreadable["turnaround"], editing.problemFor(CodeNegativeTurnaround, "", "")))
		})
	})
}

func (screens *Screens) variantSection(context *ui.Context) {
	editing := screens.editor
	edited := &editing.form
	theme := context.Theme
	look.Section(context, "Variantes", func() {
		look.Paragraph(context, "Tamaños, formatos o niveles de detalle, cada uno con su precio. Con una sola, no hace falta nombre.",
			ui.TextStyle{Color: theme.InkSoft})
		if words := editing.problemFor(CodeNoVariants, "", ""); words != "" {
			context.Text(words, ui.TextStyle{Color: theme.Danger})
		}
		for index := range edited.variants {
			context.Push(edited.variants[index].id)
			if screens.variantCard(context, &edited.variants[index]) {
				edited.variants = append(edited.variants[:index:index], edited.variants[index+1:]...)
				context.Pop()
				break
			}
			context.Pop()
		}
		if context.IconButton("plus", "Añadir variante") {
			stock := ""
			if edited.product.Kind == Physical {
				stock = "0"
			}
			edited.variants = append(edited.variants, variantForm{id: screens.newID(), stock: stock})
		}
	})
}

// variantCard draws one variant's fields; it returns true on the frame it is removed.
func (screens *Screens) variantCard(context *ui.Context, typed *variantForm) bool {
	editing := screens.editor
	theme := context.Theme
	id := typed.id
	removed := false
	look.Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(12), Gap: theme.Gap}, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.End}, func() {
			look.Field(context, "Nombre", "variant-name", &typed.name, "A4, Busto…",
				firstOf(editing.problemFor(CodeBlankVariantName, id, ""), editing.problemFor(CodeDuplicateVariantName, id, "")))
			removed = len(editing.form.variants) > 1 && context.FlatIconButton("trash-2", "##Quitar variante")
		})
		context.Box(ui.Layout{Width: ui.Grow(), Columns: look.Columns(context, 3), Gap: theme.Gap, AlignY: ui.Start}, func() {
			look.Field(context, "Precio (€)", "price", &typed.price, "12,50", editing.problemFor(CodeNegativePrice, id, "price"))
			look.Field(context, "Coste (€)", "cost", &typed.cost, "3,50", editing.problemFor(CodeNegativeCost, id, "cost"))
			marginWords(context, typed)
		})
		if editing.form.product.Kind == Physical {
			screens.physicalFields(context, typed)
		}
	})
	return removed
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
		look.Field(context, "Stock", "stock", &typed.stock, "0", editing.problemFor(CodeNegativeStock, id, "stock"))
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
