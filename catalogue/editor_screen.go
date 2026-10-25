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

func (screens *Screens) duplicate() {
	edited := &screens.editor.form
	id, err := screens.store.Duplicate(edited.product.ID, edited.product.Name+" (copia)")
	screens.tell(err, "Copia creada como borrador")
	if copied, found := screens.product(id); err == nil && found {
		screens.edit(copied)
	}
}

// lifeButtons archive or restore the product, or delete it, at the end of «Más detalles».
func (screens *Screens) lifeButtons(context *ui.Context) {
	edited := &screens.editor.form
	if edited.isNew {
		return
	}
	archived := edited.product.Status == Archived
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 8}, func() {
		if context.FlatIconButton("archive", archiveWords(archived)) {
			screens.tell(screens.store.SetArchived(edited.product.ID, !archived), archivedWords(archived))
			screens.editor = nil
			return
		}
		if context.DangerFlatIconButton("trash-2", "Eliminar") {
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
			if context.DangerButton("Eliminar") {
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

func (edited *form) addTag(name string) {
	edited.tags = DistinctTagNames(append(edited.tags, name))
	edited.tagTyped = ""
}

func (screens *Screens) detailsSummary() string {
	if screens.editor.form.product.Kind == Physical {
		return "SKU, medidas, peso y aviso de stock bajo"
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
			look.Field(context, "Formato", "delivery", &edited.delivery, "Ej.: PNG 4000 px", "")
			look.Field(context, "Plazo en días", "turnaround", &edited.turnaround, "En días",
				firstOf(editing.unreadable["turnaround"], editing.problemFor(CodeNegativeTurnaround, "", "")))
		})
	} else {
		look.Field(context, "Material", "material", &edited.material, "Ej.: Papel de algodón 300 g", "")
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
		look.Field(context, "SKU", "sku", &typed.sku, "Ej.: LAM-GAT-A4", editing.problemFor(CodeDuplicateSKU, id, "sku"))
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
