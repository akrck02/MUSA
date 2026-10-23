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
	addingTag   bool   // the field to add a tag is open
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
	sectionTitleScale = 1.4
	backButtonSide    = 48
	moneyColumnWidth  = 130
	titleScale        = 1.6
	amountFieldWidth  = 64
	stockColumnWidth  = 250
	stepCountWidth    = 56
	stepCountScale    = 1.5
)

func (screens *Screens) drawEditor(context *ui.Context) {
	screens.editorHeader(context)
	if screens.editor == nil {
		return
	}
	theme := context.Theme
	side := ui.Fixed(editorSideWidth)
	if look.Narrow(context) {
		side = ui.Grow()
	}
	context.Box(ui.Layout{Direction: look.Across(context), Width: ui.Grow(), Gap: theme.Gap * 1.5, AlignY: ui.Start}, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: side, Gap: theme.Gap}, func() {
			screens.basics(context)
			screens.tagSection(context)
		})
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: theme.Gap}, func() {
			screens.variantSection(context)
			screens.moreDetails(context)
		})
	})
	if screens.editor.asking == "delete" {
		screens.askDelete(context)
	}
}

// editorHeader is a round back button and the product's name, with Duplicar and Guardar.
func (screens *Screens) editorHeader(context *ui.Context) {
	theme := context.Theme
	edited := &screens.editor.form
	title := nameOr(edited.product.Name, "Nuevo producto")
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 14, AlignY: ui.Center}, func() {
		round := ui.Look{Fill: theme.Sunken, Radius: backButtonSide / 2}
		layout := ui.Layout{Width: ui.Fixed(backButtonSide), Height: ui.Fixed(backButtonSide), AlignX: ui.Center, AlignY: ui.Center}
		if context.Clickable("Volver", layout, ui.Looks{Normal: round, Hovered: round, Pressed: round}, func() {
			context.IconSized("chevron-left", theme.IconSize, theme.Ink)
		}) {
			screens.editor = nil
			return
		}
		context.Text(title, ui.TextStyle{Size: theme.HeadingSize * titleScale, Weight: draw2d.Bold})
		context.Spacer()
		if !edited.isNew && context.Button("Duplicar") {
			screens.duplicate()
			return
		}
		if context.PrimaryButton("Guardar") {
			screens.save()
		}
	})
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
		if context.FlatIconButton("archive", map[bool]string{true: "Restaurar", false: "Archivar"}[archived]) {
			screens.tell(screens.store.SetArchived(edited.product.ID, !archived), map[bool]string{true: "Restaurado", false: "Archivado"}[archived])
			screens.editor = nil
			return
		}
		if context.FlatIconButton("trash-2", "Eliminar") {
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
	screens.editorPicture(context)
	softLabel(context, "Nombre")
	context.TextFieldWith("name", &edited.name, ui.TextFieldOptions{Placeholder: "Gato astronauta"})
	if words := editing.problemFor(CodeBlankName, "", ""); words != "" {
		context.Text(words, ui.TextStyle{Size: theme.SmallTextSize, Color: theme.Danger})
	}
	context.Box(ui.Layout{Gap: 8, AlignY: ui.Center}, func() {
		context.IconSized("lock", theme.IconSize*0.8, theme.InkSoft)
		context.Text(kindWords[edited.product.Kind]+" · el tipo no se cambia", ui.TextStyle{Color: theme.InkSoft})
	})
}

func softLabel(context *ui.Context, words string) {
	context.Text(words, ui.TextStyle{Color: context.Theme.InkSoft})
}

// editorPicture is the product's picture, square, with its first letter until it has a photo.
func (screens *Screens) editorPicture(context *ui.Context) {
	theme := context.Theme
	product := screens.editor.form.product
	name := nameOr(product.Name, screens.editor.form.name)
	tone := TagTone(theme, PictureColor(Product{Name: name}))
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Fixed(editorSideWidth), Padding: ui.Pad(theme.Padding), AlignX: ui.Center}
	context.Frame(layout, ui.Look{Fill: tone.Fill, Radius: theme.PanelRadius}, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Height: ui.Grow(), AlignX: ui.Center, AlignY: ui.Center}, func() {
			context.Text(Initial(nameOr(name, "?")), ui.TextStyle{Size: cardLetterSize * 2.5, Weight: draw2d.Bold, Color: theme.Surface.WithAlpha(0.7)})
		})
		pill := ui.Look{Fill: theme.Surface, Radius: theme.ControlHeight / 2}
		layout := ui.Layout{Height: ui.Fixed(theme.ControlHeight), Padding: ui.PadXY(18, 0), Gap: 8, AlignY: ui.Center}
		if context.Clickable("Cambiar foto", layout, ui.Looks{Normal: pill, Hovered: pill, Pressed: pill}, func() {
			context.IconSized("image", theme.IconSize, theme.Ink)
			context.Text("Cambiar foto", ui.TextStyle{})
		}) {
			// TODO(akrck02): choose a photo of the work from the device's pictures (proposal 3, R12).
			screens.tell(nil, "Las fotos llegarán pronto")
		}
	})
}

// tagSection is the product's tags as chips, and «+ Añadir», which opens a field with suggestions.
func (screens *Screens) tagSection(context *ui.Context) {
	editing := screens.editor
	edited := &editing.form
	theme := context.Theme
	softLabel(context, "Etiquetas")
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, Wrap: true}, func() {
		for index, name := range edited.tags {
			context.Push(name)
			if _, removed := context.RemovableChip("", name, chipLook(look.Neutral(theme))); removed {
				edited.tags = append(edited.tags[:index:index], edited.tags[index+1:]...)
			}
			context.Pop()
		}
		if context.Chip("plus", "Añadir", chipLook(look.Neutral(theme))) {
			editing.addingTag = !editing.addingTag
		}
	})
	if !editing.addingTag {
		return
	}
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.Center}, func() {
		context.TextFieldWith("tag", &edited.tagTyped, ui.TextFieldOptions{Placeholder: "Lámina, pegatina, gatos…"})
		if context.IconButton("check", "##Añadir etiqueta") {
			edited.addTag(edited.tagTyped)
		}
	})
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, Wrap: true}, func() {
		for _, tag := range SuggestTags(screens.tags, edited.tags, edited.tagTyped) {
			context.Push("suggested-" + tag.ID)
			if context.Chip("plus", tag.Name, chipLook(look.Neutral(theme))) {
				edited.addTag(tag.Name)
			}
			context.Pop()
		}
	})
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
	physical := edited.product.Kind == Physical
	look.Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(theme.Padding * 1.5), Gap: 10}, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Gap: 2}, func() {
			context.Text("Versiones", ui.TextStyle{Size: theme.TextSize * sectionTitleScale, Weight: draw2d.Bold})
			context.Text(map[bool]string{true: "Cada versión tiene su precio y su stock.", false: "Cada versión tiene su precio."}[physical],
				ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
		})
		heading := func(words string) func() {
			return func() { context.Text(words, ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft}) }
		}
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 8}, func() {
			versionLine(context, physical, heading("Versión"), heading("Precio"), heading("Te cuesta"), heading("En stock"))
			if len(edited.variants) > 1 {
				context.Box(ui.Layout{Width: ui.Fixed(theme.ControlHeight)}, nil) // above the rows' «×»
			}
		})
		if words := editing.problemFor(CodeNoVariants, "", ""); words != "" {
			context.Text(words, ui.TextStyle{Color: theme.Danger})
		}
		for index := range edited.variants {
			context.Separator()
			context.Push(edited.variants[index].id)
			removed := screens.variantRow(context, &edited.variants[index])
			context.Pop()
			if removed {
				edited.variants = append(edited.variants[:index:index], edited.variants[index+1:]...)
				break
			}
		}
		context.Separator()
		if context.FlatIconButton("plus", "Añadir versión") {
			edited.variants = append(edited.variants, variantForm{id: screens.newID(), stock: "0"})
		}
	})
}

// versionLine lays out one line of the versions' table: the name takes what the price, the cost
// and the stock leave.
func versionLine(context *ui.Context, physical bool, name, price, cost, stock func()) {
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 12, AlignY: ui.Center}, func() {
		context.Box(ui.Layout{Width: ui.Grow()}, name)
		context.Box(ui.Layout{Width: ui.Fixed(moneyColumnWidth)}, price)
		context.Box(ui.Layout{Width: ui.Fixed(moneyColumnWidth)}, cost)
		if physical {
			context.Box(ui.Layout{Width: ui.Fixed(stockColumnWidth), AlignY: ui.Center}, stock)
		}
	})
}

// variantRow is one version's line; it returns true on the frame it is removed.
func (screens *Screens) variantRow(context *ui.Context, typed *variantForm) bool {
	editing := screens.editor
	theme := context.Theme
	id := typed.id
	physical := editing.form.product.Kind == Physical
	removed := false
	problem := firstOf(editing.problemFor(CodeBlankVariantName, id, ""), editing.problemFor(CodeDuplicateVariantName, id, ""),
		editing.problemFor(CodeNegativePrice, id, "price"), editing.problemFor(CodeNegativeCost, id, "cost"))
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.Center}, func() {
		context.Box(ui.Layout{Width: ui.Grow()}, func() {
			versionLine(context, physical,
				func() { plainField(context, "variant-name", &typed.name, "Única", draw2d.Bold) },
				func() { moneyField(context, "price", &typed.price, "12,50") },
				func() { moneyField(context, "cost", &typed.cost, "3,50") },
				func() { stockCell(context, typed) })
		})
		removed = len(editing.form.variants) > 1 && context.FlatIconButton("x", "##Quitar versión")
	})
	if problem != "" {
		context.Text(problem, ui.TextStyle{Size: theme.SmallTextSize, Color: theme.Danger})
	}
	return removed
}

// plainField is a field that reads as the table's text until it is touched.
func plainField(context *ui.Context, name string, text *string, placeholder string, weight draw2d.Weight) {
	context.TextFieldWith(name, text, ui.TextFieldOptions{Placeholder: placeholder, Plain: true, Weight: weight})
}

// moneyField is a plain field for an amount, with the euro sign after it.
func moneyField(context *ui.Context, name string, text *string, placeholder string) {
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 2, AlignY: ui.Center}, func() {
		context.TextFieldWith(name, text, ui.TextFieldOptions{Placeholder: placeholder, Plain: true, Width: ui.Fixed(amountFieldWidth)})
		context.Text("€", ui.TextStyle{})
	})
}

// stockCell is − count + for a version's stock, with «Pocos» when it reaches the warning.
func stockCell(context *ui.Context, typed *variantForm) {
	theme := context.Theme
	count, err := strconv.Atoi(typed.stock)
	if err != nil {
		context.TextFieldWith("stock", &typed.stock, ui.TextFieldOptions{Placeholder: "0"})
		return
	}
	context.Box(ui.Layout{Gap: 8, AlignY: ui.Center}, func() {
		if context.IconButton("minus", "##Quitar uno") && count > 0 {
			count--
		}
		context.Box(ui.Layout{Width: ui.Fixed(stepCountWidth), AlignX: ui.Center}, func() {
			context.Text(strconv.Itoa(count), ui.TextStyle{Size: theme.TextSize * stepCountScale, Weight: draw2d.Bold})
		})
		if context.IconButton("plus", "##Añadir uno") {
			count++
		}
		typed.stock = strconv.Itoa(count)
		minimum, _ := strconv.Atoi(typed.minimum) // an unreadable minimum is told when saving
		if (Variant{Stock: count, MinimumStock: minimum}).LowStock() {
			context.Text(map[bool]string{true: "Agotado", false: "Pocos"}[count == 0],
				ui.TextStyle{Size: theme.SmallTextSize, Weight: draw2d.Bold, Color: theme.Warning})
		}
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
			screens.lifeButtons(context)
		}
	})
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
