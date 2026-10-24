package catalogue

import (
	"strconv"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/look"
)

// Measures of the versions' panel, in pixels, as in the author's design.
const (
	panelRadius       = 20
	panelPadding      = 22 + 1
	panelTitleSize    = 22
	panelTitleHeight  = 1.2
	panelTitleBottom  = 4
	panelNoteBottom   = 16
	boxGap            = 16
	boxRadius         = 14
	boxPadding        = 16
	boxValueSize      = 26
	headerRowBottom   = 6
	headerRowSize     = 14
	versionRowY       = 14
	versionColumnsGap = 14
	versionNameSize   = 19
	versionMoneySize  = 18
	nameShare         = 1.1
	moneyShare        = 1
	stockShare        = 1.6
	stepSide          = 48
	stepRadius        = 12
	stepGap           = 10
	stepIconSize      = 20
	stepCountSize     = 22
	lowWordsSize      = 13
	addVersionSize    = 17
	addVersionGap     = 10
	addVersionTop     = 16
	addVersionBottom  = 4
	addOtherTop       = 18
	addOtherBottom    = 2
	addOtherHintSize  = 15
	moreTop           = 16
	moreRadius        = 16
	morePaddingX      = 22 + 1
	morePaddingY      = 18 + 1
	moreSize          = 18
	moreHintSize      = 15
	moreIconSize      = 20
	removeVersionSide = 32
	rowTop            = 6 // the design's rows keep this space above their line
)

// variantSection is the price and stock of a product of one version, or the table of its
// versions when it has several.
func (screens *Screens) variantSection(context *ui.Context) {
	theme := context.Theme
	card := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: panelRadius}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(panelPadding)}, card, func() {
		if len(screens.editor.form.variants) == 1 {
			screens.singleVersion(context)
			return
		}
		screens.versionTable(context)
	})
}

// panelHeading is the panel's title over a soft line.
func panelHeading(context *ui.Context, title, note string) {
	context.Text(title, ui.TextStyle{Size: panelTitleSize, Weight: draw2d.Bold, LineHeight: panelTitleHeight})
	look.Space(context, panelTitleBottom)
	context.Text(note, ui.TextStyle{Color: context.Theme.InkSoft})
	look.Space(context, panelNoteBottom)
}

// singleVersion is a product of one version: its price and cost in two boxes, its stock in a
// third with − and +, and a way to add another version.
func (screens *Screens) singleVersion(context *ui.Context) {
	editing := screens.editor
	physical := editing.form.product.Kind == Physical
	typed := &editing.form.variants[0]
	panelHeading(context, map[bool]string{true: "Precio y stock", false: "Precio"}[physical], "Una sola versión.")
	context.Push(typed.id)
	context.Box(ui.Layout{Width: ui.Grow(), Gap: boxGap}, func() {
		valueBox(context, "Precio", func() { amountField(context, "price", &typed.price, "12,50", boxValueSize, draw2d.Bold) })
		valueBox(context, "Te cuesta", func() { amountField(context, "cost", &typed.cost, "3,50", boxValueSize, draw2d.Bold) })
	})
	if physical {
		look.Space(context, boxGap)
		stockBox(context, typed)
	}
	screens.versionProblem(context, typed.id)
	context.Pop()
	context.Box(ui.Layout{Padding: ui.Padding{Top: addOtherTop, Bottom: addOtherBottom}}, func() {
		if addVersionLink(context, "Añadir otra versión", "· por ejemplo, otro color") {
			screens.addVersion()
		}
	})
}

// valueBox is a soft box with a label over a big value.
func valueBox(context *ui.Context, label string, value func()) {
	box := ui.Look{Fill: context.Theme.Background, Radius: boxRadius}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Share(1), Padding: ui.Pad(boxPadding)}, box, func() {
		fieldLabel(context, label)
		value()
	})
}

// stockBox is the stock of a product of one version, with − and + to count it.
func stockBox(context *ui.Context, typed *variantForm) {
	theme := context.Theme
	count, err := strconv.Atoi(typed.stock)
	box := ui.Look{Fill: theme.Background, Radius: boxRadius}
	context.Frame(ui.Layout{Width: ui.Grow(), Padding: ui.Pad(boxPadding), AlignY: ui.Center}, box, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow()}, func() {
			fieldLabel(context, "En stock")
			if err != nil {
				context.TextFieldWith("stock", &typed.stock, ui.TextFieldOptions{Placeholder: "0"})
				return
			}
			context.Text(strconv.Itoa(count), ui.TextStyle{Size: boxValueSize, Weight: draw2d.Bold})
		})
		if err == nil {
			context.Box(ui.Layout{Gap: stepGap, AlignY: ui.Center}, func() {
				typed.stock = strconv.Itoa(stepped(context, count))
			})
		}
	})
}

// stepped is − and + around nothing: the count after this frame's taps.
func stepped(context *ui.Context, count int) int {
	if stepButton(context, "minus", "Quitar uno") && count > 0 {
		count--
	}
	if stepButton(context, "plus", "Añadir uno") {
		count++
	}
	return count
}

// stepButton is a square button with an icon, on the page's colour with a fine line.
func stepButton(context *ui.Context, icon, label string) bool {
	theme := context.Theme
	normal := ui.Look{Fill: theme.Background, Border: theme.Line, BorderWidth: 1, Radius: stepRadius}
	hovered := normal
	hovered.Fill = theme.Sunken
	side := ui.Fixed(stepSide)
	clicked := context.Clickable("##"+label, ui.Layout{Width: side, Height: side, AlignX: ui.Center, AlignY: ui.Center},
		ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
			context.IconSized(icon, stepIconSize, theme.Ink)
		})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: label})
	return clicked
}

// amountField is an amount the artist can type over, with the euro sign right after it.
func amountField(context *ui.Context, name string, text *string, placeholder string, size float32, weight draw2d.Weight) {
	theme := context.Theme
	context.Box(ui.Layout{AlignY: ui.Center, Padding: ui.Padding{Left: -fieldInset}}, func() {
		context.TextFieldWith(name, text, ui.TextFieldOptions{Placeholder: placeholder, Plain: true, FitText: true,
			Size: size, Weight: weight, Height: ui.Fixed(size * theme.LineHeight)})
		context.Text("€", ui.TextStyle{Size: size, Weight: weight})
	})
}

// addVersionLink is a plus and bold words, with a soft hint after them when given.
func addVersionLink(context *ui.Context, words, hint string) bool {
	theme := context.Theme
	clicked := context.Clickable(words, ui.Layout{Gap: addVersionGap, AlignY: ui.Center}, ui.Looks{}, func() {
		context.IconSized("plus", stepIconSize, theme.Ink)
		context.Text(words, ui.TextStyle{Size: addVersionSize, Weight: draw2d.Bold})
		if hint != "" {
			context.Text(hint, ui.TextStyle{Size: addOtherHintSize, Color: theme.InkSoft})
		}
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: words})
	return clicked
}

func (screens *Screens) addVersion() {
	edited := &screens.editor.form
	edited.variants = append(edited.variants, variantForm{id: screens.newID(), stock: "0"})
}

// versionTable is the versions of a product, one line each: its name, price, cost and stock.
func (screens *Screens) versionTable(context *ui.Context) {
	editing := screens.editor
	edited := &editing.form
	theme := context.Theme
	physical := edited.product.Kind == Physical
	panelHeading(context, "Versiones", map[bool]string{true: "Cada versión tiene su precio y su stock.", false: "Cada versión tiene su precio."}[physical])
	heading := func(words string) func() {
		return func() { context.Text(words, ui.TextStyle{Size: headerRowSize, Color: theme.InkSoft}) }
	}
	look.Space(context, rowTop)
	context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Bottom: headerRowBottom}}, func() {
		versionLine(context, physical, heading("Versión"), heading("Precio"), heading("Te cuesta"), heading("En stock"))
	})
	if words := editing.problemFor(CodeNoVariants, "", ""); words != "" {
		context.Text(words, ui.TextStyle{Color: theme.Danger})
	}
	for index := range edited.variants {
		look.Space(context, rowTop)
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
	context.Box(ui.Layout{Padding: ui.Padding{Top: addVersionTop, Bottom: addVersionBottom}}, func() {
		if addVersionLink(context, "Añadir versión", "") {
			screens.addVersion()
		}
	})
}

// versionLine lays out one line of the versions' table in shares of its width, as the design's
// grid: the name, the price, the cost and the stock.
func versionLine(context *ui.Context, physical bool, name, price, cost, stock func()) {
	context.Box(ui.Layout{Width: ui.Grow(), Gap: versionColumnsGap, AlignY: ui.Center}, func() {
		context.Box(ui.Layout{Width: ui.Share(nameShare)}, name)
		context.Box(ui.Layout{Width: ui.Share(moneyShare)}, price)
		context.Box(ui.Layout{Width: ui.Share(moneyShare)}, cost)
		if physical {
			context.Box(ui.Layout{Width: ui.Share(stockShare), AlignY: ui.Center}, stock)
		}
	})
}

// variantRow is one version's line; it returns true on the frame it is removed.
func (screens *Screens) variantRow(context *ui.Context, typed *variantForm) bool {
	theme := context.Theme
	removed := false
	context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(0, versionRowY)}, func() {
		versionLine(context, screens.editor.form.product.Kind == Physical,
			func() {
				context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Left: -fieldInset}}, func() {
					context.TextFieldWith("variant-name", &typed.name, ui.TextFieldOptions{Placeholder: "Única", Plain: true,
						Size: versionNameSize, Weight: draw2d.Bold, Height: ui.Fixed(versionNameSize * theme.LineHeight)})
				})
			},
			func() { amountField(context, "price", &typed.price, "12,50", versionMoneySize, draw2d.Regular) },
			func() { amountField(context, "cost", &typed.cost, "3,50", versionMoneySize, draw2d.Regular) },
			func() { removed = stockCell(context, typed) })
	})
	screens.versionProblem(context, typed.id)
	return removed
}

// versionProblem is what is wrong with a version, in red, after saving failed.
func (screens *Screens) versionProblem(context *ui.Context, id string) {
	editing := screens.editor
	problem := firstOf(editing.problemFor(CodeBlankVariantName, id, ""), editing.problemFor(CodeDuplicateVariantName, id, ""),
		editing.problemFor(CodeNegativePrice, id, "price"), editing.problemFor(CodeNegativeCost, id, "cost"))
	if problem != "" {
		context.Text(problem, ui.TextStyle{Size: headerRowSize, Color: context.Theme.Danger})
	}
}

// stockCell is − count + for a version's stock, with «Pocos» when it reaches the warning, and a
// quiet × at the end that removes the version; it returns true on the frame it is removed.
func stockCell(context *ui.Context, typed *variantForm) bool {
	theme := context.Theme
	count, err := strconv.Atoi(typed.stock)
	if err != nil {
		context.TextFieldWith("stock", &typed.stock, ui.TextFieldOptions{Placeholder: "0"})
		return false
	}
	removed := false
	context.Box(ui.Layout{Width: ui.Grow(), Gap: stepGap, AlignY: ui.Center}, func() {
		if stepButton(context, "minus", "Quitar uno") && count > 0 {
			count--
		}
		context.Box(ui.Layout{Width: ui.Fixed(stepSide), AlignX: ui.Center}, func() {
			context.Text(strconv.Itoa(count), ui.TextStyle{Size: stepCountSize, Weight: draw2d.Bold})
		})
		if stepButton(context, "plus", "Añadir uno") {
			count++
		}
		typed.stock = strconv.Itoa(count)
		minimum, _ := strconv.Atoi(typed.minimum) // an unreadable minimum is told when saving
		if (Variant{Stock: count, MinimumStock: minimum}).LowStock() {
			context.Text(map[bool]string{true: "Agotado", false: "Pocos"}[count == 0],
				ui.TextStyle{Size: lowWordsSize, Weight: draw2d.Bold, Color: theme.Warning})
		}
		context.Spacer()
		removed = removeButton(context)
	})
	return removed
}

// removeButton is a quiet × that removes a version.
func removeButton(context *ui.Context) bool {
	side := ui.Fixed(removeVersionSide)
	clicked := context.Clickable("##Quitar versión", ui.Layout{Width: side, Height: side, AlignX: ui.Center, AlignY: ui.Center}, ui.Looks{}, func() {
		context.IconSized("x", lowWordsSize, context.Theme.InkQuiet)
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: "Quitar versión"})
	return clicked
}

// moreDetails folds away what is set once: the description, the status, what it is made of or how
// it is delivered, and each version's SKU, warning and measures.
func (screens *Screens) moreDetails(context *ui.Context) {
	editing := screens.editor
	theme := context.Theme
	look.Space(context, moreTop)
	card := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: moreRadius}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.PadXY(morePaddingX, morePaddingY), Gap: theme.Gap}, card, func() {
		if context.Clickable("Más detalles", ui.Layout{Width: ui.Grow(), AlignY: ui.Center}, ui.Looks{}, func() {
			context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow()}, func() {
				context.Text("Más detalles", ui.TextStyle{Size: moreSize})
				// The hint is a second line of the title's paragraph in the design, as tall as its lines.
				hintHeight := moreSize * theme.LineHeight / moreHintSize
				context.Text(screens.detailsSummary(), ui.TextStyle{Size: moreHintSize, Color: theme.InkSoft, LineHeight: hintHeight})
			})
			context.IconSized(map[bool]string{true: "chevron-up", false: "chevron-down"}[editing.detailsOpen], moreIconSize, theme.Ink)
		}) {
			editing.detailsOpen = !editing.detailsOpen
		}
		context.Fold("details", editing.detailsOpen, look.DialogMotion, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: theme.Gap}, func() {
			screens.details(context)
			screens.lifeButtons(context)
		})
	})
}
