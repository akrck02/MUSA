package sales

import (
	"fmt"
	"strconv"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
)

// orderEditor is the new order's text, as typed, and what the last attempt found wrong.
type orderEditor struct {
	channel         int // index in SelectableChannels
	day             string
	customer, note  string
	lines           []lineForm
	shippingCharged string
	shippingCost    string

	choices  []choice
	problems []Problem
	typing   map[string]string // unreadable fields, by name
}

// lineForm is one line as typed: which variant (an index in the editor's choices), how many, at what price.
type lineForm struct {
	id       string
	choice   int
	quantity string
	price    string
}

// choice is a variant the order can sell, as the picker lists it.
type choice struct {
	variantID string
	words     string
	price     money.Money
	cost      money.Money
}

func newOrderEditor(newID func() string, today calendar.Date) orderEditor {
	return orderEditor{day: calendar.Typed(today), lines: []lineForm{{id: newID(), choice: -1, quantity: "1"}}, typing: map[string]string{}}
}

// choicesOf lists every variant of the products that are not archived, with their stock.
func choicesOf(products []catalogue.Product) []choice {
	var choices []choice
	for _, product := range catalogue.Filter(products, catalogue.Query{}) {
		for _, variant := range product.Variants {
			words := ItemLabel(product.Name, variant.Name)
			if product.Kind == catalogue.Physical {
				words += fmt.Sprintf(" (%d)", variant.Stock)
			}
			choices = append(choices, choice{variantID: variant.ID, words: words, price: variant.Price, cost: variant.Cost})
		}
	}
	return choices
}

func (screens *Screens) drawEditor(context *ui.Context) {
	editing := screens.editor
	theme := context.Theme
	if editing.choices == nil {
		products, err := screens.products.All()
		if err != nil {
			screens.tell(true, err.Error())
		}
		editing.choices = choicesOf(products)
	}
	look.PageHeader(context, "Nueva venta", "Un pedido de uno o más productos", func() {
		if context.FlatIconButton("arrow-left", "Volver") {
			screens.editor = nil
		}
		if context.PrimaryIconButton("check", "Guardar") {
			screens.record()
		}
	})
	if screens.editor == nil {
		return
	}
	for _, problem := range editing.problems {
		if problem.LineID == "" {
			context.Text(ProblemWords(problem), ui.TextStyle{Color: theme.Danger})
		}
	}
	look.Section(context, "Dónde y quién", func() {
		var channels []string
		for _, channel := range SelectableChannels {
			channels = append(channels, channelWords[channel])
		}
		context.Box(ui.Layout{Width: ui.Grow(), Columns: look.Columns(context, 2), Gap: theme.Gap, AlignY: ui.End}, func() {
			context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 4}, func() {
				context.Text("Canal", ui.TextStyle{Size: theme.SmallTextSize, Weight: draw2d.Bold})
				context.Select("channel", &editing.channel, channels)
			})
			look.Field(context, "Día", "day", &editing.day, "dd/mm/aaaa", editing.typing["day"])
			look.Field(context, "Cliente", "customer", &editing.customer, "Opcional", "")
			look.Field(context, "Nota", "note", &editing.note, "Ej.: Pedido web 0007", "")
		})
	})
	screens.lineSection(context)
	look.Section(context, "Envío", func() {
		context.Box(ui.Layout{Width: ui.Grow(), Columns: look.Columns(context, 2), Gap: theme.Gap}, func() {
			look.Field(context, "Cobrado al cliente (€)", "charged", &editing.shippingCharged, "0", editing.typing["charged"])
			look.Field(context, "Pagado por ti (€)", "cost", &editing.shippingCost, "0", editing.typing["cost"])
		})
	})
	if draft, ok := editing.draft(); ok {
		sale := screens.preview(draft)
		look.Section(context, "Resumen", func() {
			amountRow(context, "Total", sale.Total(), true)
			amountRow(context, "Coste", sale.Cost(), false)
			amountRow(context, "Margen", sale.Margin(), true)
		})
	}
}

func (screens *Screens) lineSection(context *ui.Context) {
	editing := screens.editor
	theme := context.Theme
	var words []string
	for _, each := range editing.choices {
		words = append(words, each.words)
	}
	look.Section(context, "Productos", func() {
		if len(words) == 0 {
			look.Paragraph(context, "El catálogo está vacío: crea antes un producto.", ui.TextStyle{Color: theme.InkSoft})
			return
		}
		for index := range editing.lines {
			line := &editing.lines[index]
			context.Push(line.id)
			if line.choice < 0 {
				line.choice = 0
				line.price = money.FormatForInput(editing.choices[0].price)
			}
			removed := false
			context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.End}, func() {
				context.Box(ui.Layout{Width: ui.Grow()}, func() {
					if context.Select("variant", &line.choice, words) {
						line.price = money.FormatForInput(editing.choices[line.choice].price)
					}
				})
				context.Box(ui.Layout{Width: ui.Fixed(90)}, func() {
					look.Field(context, "Cantidad", "quantity", &line.quantity, "1", screens.lineProblem(line.id, CodeInvalidQuantity))
				})
				context.Box(ui.Layout{Width: ui.Fixed(120)}, func() {
					look.Field(context, "Precio (€)", "price", &line.price, "", screens.lineProblem(line.id, CodeNegativePrice))
				})
				removed = len(editing.lines) > 1 && context.DangerFlatIconButton("trash-2", "##Quitar línea")
			})
			for _, problem := range editing.problems {
				if problem.Code == CodeNotEnoughStock && problem.VariantID == editing.choices[line.choice].variantID {
					context.Text(ProblemWords(problem), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.Danger})
				}
			}
			context.Pop()
			if removed {
				editing.lines = append(editing.lines[:index:index], editing.lines[index+1:]...)
				break
			}
		}
		if context.IconButton("plus", "Añadir producto") {
			editing.lines = append(editing.lines, lineForm{id: screens.newID(), choice: -1, quantity: "1"})
		}
	})
}

func (screens *Screens) lineProblem(lineID string, code Code) string {
	if words, found := screens.editor.typing[lineID+"/"+string(code)]; found {
		return words
	}
	for _, problem := range screens.editor.problems {
		if problem.LineID == lineID && problem.Code == code {
			return ProblemWords(problem)
		}
	}
	return ""
}

// draft reads the typed text into a draft; ok is false while a field cannot be read, and the
// editor's typing then says which.
func (editing *orderEditor) draft() (draft Draft, ok bool) {
	editing.typing = map[string]string{}
	day, found := calendar.ParseTyped(editing.day)
	if !found {
		editing.typing["day"] = "Escribe el día como 18/4/2026"
	}
	draft = Draft{Channel: SelectableChannels[editing.channel], Customer: editing.customer, Note: editing.note,
		ShippingCharged: editing.amount(editing.shippingCharged, "charged"), ShippingCost: editing.amount(editing.shippingCost, "cost")}
	for _, line := range editing.lines {
		if line.choice < 0 || line.choice >= len(editing.choices) {
			continue
		}
		quantity, err := strconv.Atoi(line.quantity)
		if err != nil {
			editing.typing[line.id+"/"+string(CodeInvalidQuantity)] = "Escribe un número entero"
		}
		draft.Lines = append(draft.Lines, LineDraft{ID: line.id, VariantID: editing.choices[line.choice].variantID,
			Quantity: quantity, UnitPrice: editing.amount(line.price, line.id+"/"+string(CodeNegativePrice))})
	}
	draft.SoldAt = day.At(noon, time.Local) // record() sets the real moment
	return draft, len(editing.typing) == 0
}

func (editing *orderEditor) amount(text, key string) money.Money {
	if text == "" {
		return 0
	}
	amount, found := money.Parse(text)
	if !found {
		editing.typing[key] = "Escribe un importe como 4 o 4,50"
	}
	return amount
}

// preview is the sale the draft would make, with each variant's cost, for the summary.
func (screens *Screens) preview(draft Draft) Sale {
	costs := map[string]money.Money{}
	for _, each := range screens.editor.choices {
		costs[each.variantID] = each.cost
	}
	sale := Sale{ShippingCharged: draft.ShippingCharged, ShippingCost: draft.ShippingCost}
	for _, line := range draft.Lines {
		sale.Lines = append(sale.Lines, Line{Quantity: line.Quantity, UnitPrice: line.UnitPrice, UnitCost: costs[line.VariantID]})
	}
	return sale
}

func (screens *Screens) record() {
	editing := screens.editor
	draft, ok := editing.draft()
	if !ok {
		return
	}
	day, _ := calendar.ParseTyped(editing.day) // draft read it
	draft.SoldAt = SoldAt(day, screens.now())
	_, problems, err := screens.store.Record(draft)
	editing.problems = problems
	if err != nil {
		screens.tell(true, err.Error())
		return
	}
	if len(problems) > 0 {
		return
	}
	screens.editor = nil
	screens.tell(false, "Venta guardada")
}
