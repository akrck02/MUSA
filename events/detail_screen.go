package events

import (
	"fmt"
	"strconv"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
	"github.com/Itros97/MUSA/sales"
)

// sheetKind is which sheet is open over a fair's detail.
type sheetKind int

const (
	noSheet sheetKind = iota
	takeSheet
	returnSheet
	closeSheet
	deleteSheet
)

// sheet is the open sheet and its fields, as typed.
type sheet struct {
	kind      sheetKind
	variantID string
	choice    int
	choices   []stockChoice
	quantity  string
	problem   string
}

// stockChoice is a physical variant of the catalogue that can be taken to a fair.
type stockChoice struct {
	variantID, words string
}

// expenseForm is the new expense's fields, as typed.
type expenseForm struct{ concept, amount string }

// Measures of an event's page, in pixels.
const (
	pageGap       = 16 // between the figures, the cards and the columns
	figuresTop    = 26 // from the header to the figures
	cardPadding   = 20
	rowPaddingY   = 12
	rowGap        = 12
	dotSide       = 10
	meterWidth    = 140
	countWidth    = 70
	timeWidth     = 50
	stockShare    = 1.3 // how much wider the stock card is than the column of sales and expenses
	amountWidth   = 120 // the amount of a new expense
	cardTitleSize = 18
)

// drawDetail is an event's page: its header, its four figures, what was taken to it on the left,
// and its sales and expenses on the right; on a narrow window, one card under another.
func (screens *Screens) drawDetail(context *ui.Context) {
	detail := screens.detail
	event := detail.Event
	open := event.Status == Open
	screens.detailHeader(context, event, open)
	if screens.detail == nil {
		return
	}
	look.Space(context, figuresTop)
	screens.figureCards(context)
	look.Space(context, pageGap)
	if look.Narrow(context) {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: pageGap}, func() {
			screens.stockCard(context, open)
			screens.salesCard(context, open)
			screens.expensesCard(context, open)
		})
	} else {
		context.Box(ui.Layout{Width: ui.Grow(), Gap: pageGap}, func() {
			context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.GrowBy(stockShare)}, func() { screens.stockCard(context, open) })
			context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: pageGap}, func() {
				screens.salesCard(context, open)
				screens.expensesCard(context, open)
			})
		})
	}
	if screens.sheet.kind != noSheet {
		screens.drawSheet(context)
	}
}

// detailHeader is the event's name, when and where with its state, «Volver», «Abrir el puesto»
// while it is open, and «⋯» with the rest.
func (screens *Screens) detailHeader(context *ui.Context, event Event, open bool) {
	look.PageHeader(context, event.Name, event.WhenAndWhere(), func() {
		if look.ToolButton(context, "arrow-left", "Volver") {
			screens.detail = nil
			screens.Reload()
			return
		}
		statusPill(context, event.Status)
		if open && screens.openStand != nil && look.MainButton(context, "store", "Abrir el puesto") {
			screens.openStand(event)
		}
		screens.detailMenu(context, event, open)
	})
}

// figureCards are the event's accounts, a card each: sold, the cost of what was sold, expenses and
// profit, red below zero.
func (screens *Screens) figureCards(context *ui.Context) {
	summary := Summarise(*screens.detail)
	figures := []struct {
		label    string
		value    money.Money
		negative bool
	}{
		{"Vendido", summary.Gross, false},
		{"Coste de lo vendido", summary.CostOfGoods, false},
		{"Gastos", summary.Expenses, false},
		{"Beneficio", summary.Profit(), summary.Profit() < 0},
	}
	context.Box(ui.Layout{Width: ui.Grow(), Columns: look.Columns(context, len(figures)), Gap: pageGap}, func() {
		for _, figure := range figures {
			look.Card(context, ui.Layout{Width: ui.Grow(), Padding: ui.Pad(cardPadding)}, func() {
				look.Figure(context, figure.label, money.Format(figure.value), figure.negative)
			})
		}
	})
}

func (screens *Screens) detailMenu(context *ui.Context, event Event, open bool) {
	context.MenuButton("ellipsis", "##Más acciones", func() {
		if context.MenuItem("pencil", "Editar", "") {
			screens.editor = &eventEditor{draft: Draft{ID: event.ID, Name: event.Name, Location: event.Location},
				start: calendar.Typed(event.StartDate), end: calendar.Typed(event.EndDate)}
		}
		if open && context.MenuItem("lock", "Cerrar evento", "") {
			screens.sheet = sheet{kind: closeSheet}
		}
		if context.DangerMenuItem("trash-2", "Eliminar", "") {
			screens.sheet = sheet{kind: deleteSheet}
		}
	})
}

// detailCard is one card of the page: its title, an action on its right when there is one, and
// its rows. It returns true on the frame the action is tapped.
func detailCard(context *ui.Context, title, action string, rows func()) bool {
	acted := false
	look.Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(cardPadding)}, func() {
		context.Box(ui.Layout{Width: ui.Grow(), AlignY: ui.Center}, func() {
			context.Box(ui.Layout{Width: ui.Grow()}, func() {
				context.Text(title, ui.TextStyle{Size: cardTitleSize, Weight: draw2d.Bold})
			})
			if action != "" {
				acted = context.IconButton("plus", action)
			}
		})
		rows()
	})
	return acted
}

// row is one line of a card, with a fine line above it but the first.
func row(context *ui.Context, first bool, content func()) {
	if !first {
		context.Separator()
	}
	context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(0, rowPaddingY), Gap: rowGap, AlignY: ui.Center}, content)
}

// stockCard is what was taken to the event: a row per product with how many are left of how many
// and a thin bar. Tapping a row of an open event offers to take more or return what is left.
func (screens *Screens) stockCard(context *ui.Context, open bool) {
	action := ""
	if open {
		action = "Llevar"
	}
	if detailCard(context, "Lo que llevas", action, func() {
		if len(screens.detail.Stock) == 0 {
			look.Space(context, rowPaddingY)
			context.Text("Aún no has llevado nada. Lo que no vendas vuelve al catálogo al cerrar el evento.",
				ui.TextStyle{Color: context.Theme.InkSoft})
		}
		for index, line := range screens.detail.Stock {
			context.Push(line.VariantID)
			row(context, index == 0, func() { screens.stockRow(context, line, index, open) })
			context.Pop()
		}
	}) {
		screens.takeMore()
	}
}

// takeMore opens the choice of what to take to the event: the stand's, when the app gave it, or
// the sheet.
func (screens *Screens) takeMore() {
	if screens.prepare != nil {
		screens.prepare(screens.detail.Event)
		return
	}
	screens.sheet = sheet{kind: takeSheet, quantity: "1", choices: screens.stockChoices()}
}

// stockRow is one product taken: its colour, its name and version, a bar of what was sold and
// how many are left; chosen, «Llevar más» and «Devolver» instead of the bar.
func (screens *Screens) stockRow(context *ui.Context, line StockLine, index int, open bool) {
	theme := context.Theme
	side := ui.Fixed(dotSide)
	context.Frame(ui.Layout{Width: side, Height: side}, ui.Look{Fill: catalogue.ListFill(index), Radius: dotSide / 2}, nil)
	name := sales.ItemLabel(line.ProductName, line.VariantName)
	if open && context.Clickable(name, ui.Layout{Width: ui.Grow()}, ui.Looks{}, func() {
		context.Text(name, ui.TextStyle{})
	}) {
		screens.toggleLine(line.VariantID)
	}
	if !open {
		context.Box(ui.Layout{Width: ui.Grow()}, func() { context.Text(name, ui.TextStyle{}) })
	}
	if screens.chosenLine == line.VariantID && open {
		if screens.tonalButton(context, "Llevar más") {
			screens.sheet = sheet{kind: takeSheet, quantity: "1", choices: screens.stockChoices(), choice: screens.choiceOf(line.VariantID)}
		}
		if screens.tonalButton(context, "Devolver") {
			screens.sheet = sheet{kind: returnSheet, variantID: line.VariantID, quantity: strconv.Itoa(line.Remaining())}
		}
		return
	}
	context.Box(ui.Layout{Width: ui.Fixed(meterWidth)}, func() { look.Meter(context, line.Remaining(), line.Taken) })
	context.Box(ui.Layout{Width: ui.Fixed(countWidth), AlignX: ui.End}, func() {
		context.Text(fmt.Sprintf("%d de %d", line.Remaining(), line.Taken), ui.TextStyle{Color: theme.InkSoft})
	})
}

// toggleLine shows the actions of a stock row, or hides them when they were shown.
func (screens *Screens) toggleLine(variantID string) {
	if screens.chosenLine == variantID {
		screens.chosenLine = ""
		return
	}
	screens.chosenLine = variantID
}

// choiceOf is where a variant is among the stock that can be taken, or the first.
func (screens *Screens) choiceOf(variantID string) int {
	for index, choice := range screens.stockChoices() {
		if choice.variantID == variantID {
			return index
		}
	}
	return 0
}

// stockChoices lists the physical variants with stock in the catalogue.
func (screens *Screens) stockChoices() []stockChoice {
	products, err := screens.products.All()
	if err != nil {
		screens.tell(nil, err, "")
		return nil
	}
	var choices []stockChoice
	for _, product := range catalogue.Filter(products, catalogue.Query{Kind: catalogue.Physical}) {
		for _, variant := range product.Variants {
			if variant.Stock > 0 {
				words := fmt.Sprintf("%s (%d)", sales.ItemLabel(product.Name, variant.Name), variant.Stock)
				choices = append(choices, stockChoice{variantID: variant.ID, words: words})
			}
		}
	}
	return choices
}

// salesCard is every sale of the event, newest first, each with its undo while it is open.
func (screens *Screens) salesCard(context *ui.Context, open bool) {
	detailCard(context, "Ventas", "", func() {
		if len(screens.detail.Sales) == 0 {
			look.Space(context, rowPaddingY)
			context.Text("Sin ventas todavía.", ui.TextStyle{Color: context.Theme.InkSoft})
		}
		for index, sale := range screens.detail.Sales {
			context.Push(sale.ID)
			row(context, index == 0, func() { screens.saleRow(context, sale, open) })
			context.Pop()
		}
	})
}

// saleRow is one sale: when, what and how much, with its undo while the event is open.
func (screens *Screens) saleRow(context *ui.Context, sale sales.Sale, open bool) {
	theme := context.Theme
	line := sale.Lines[0]
	context.Box(ui.Layout{Width: ui.Fixed(timeWidth)}, func() {
		context.Text(sale.SoldAt.Format("15:04"), ui.TextStyle{Color: theme.InkSoft})
	})
	context.Frame(ui.Layout{Width: ui.Grow()}, ui.Look{Clip: true}, func() {
		context.Text(fmt.Sprintf("%d × %s", line.Quantity, sales.ItemLabel(line.ProductName, line.VariantName)), ui.TextStyle{})
	})
	context.Text(money.Format(sale.Total()), ui.TextStyle{Weight: draw2d.Bold})
	if open && context.FlatIconButton("undo-2", "##Deshacer venta") {
		refused, err := screens.store.Undo(sale.ID)
		screens.tell(refused, err, "Venta deshecha")
	}
}

// expensesCard is what the event cost, each with its red × while it is open, and «Añadir», which
// opens a small form for one more.
func (screens *Screens) expensesCard(context *ui.Context, open bool) {
	action := ""
	if open {
		action = "Añadir gasto"
	}
	if detailCard(context, "Gastos", action, func() {
		if len(screens.detail.Expenses) == 0 && !screens.addingExpense {
			look.Space(context, rowPaddingY)
			context.Text("Sin gastos apuntados.", ui.TextStyle{Color: context.Theme.InkSoft})
		}
		for index, expense := range screens.detail.Expenses {
			context.Push(expense.ID)
			row(context, index == 0, func() { screens.expenseRow(context, expense, open) })
			context.Pop()
		}
		if open && screens.addingExpense {
			screens.expenseForm(context)
		}
	}) {
		screens.addingExpense = !screens.addingExpense
	}
}

// expenseRow is one expense: what for and how much, with its red × while the event is open.
func (screens *Screens) expenseRow(context *ui.Context, expense Expense, open bool) {
	context.Box(ui.Layout{Width: ui.Grow()}, func() { context.Text(expense.Concept, ui.TextStyle{}) })
	context.Text(money.Format(expense.Amount), ui.TextStyle{Weight: draw2d.Bold})
	if open && context.DangerFlatIconButton("trash-2", "##Quitar gasto") {
		refused, err := screens.store.DeleteExpense(screens.detail.Event.ID, expense.ID)
		screens.tell(refused, err, "Gasto quitado")
	}
}

// expenseForm is the new expense's concept and amount, and the button that adds it.
func (screens *Screens) expenseForm(context *ui.Context) {
	look.Space(context, rowPaddingY)
	context.Box(ui.Layout{Direction: look.Across(context), Width: ui.Grow(), Gap: rowGap, AlignY: ui.End}, func() {
		look.Field(context, "Concepto", "concept", &screens.newExpense.concept, "Stand, gasolina, comida…", "")
		context.Box(ui.Layout{Width: ui.Fixed(amountWidth)}, func() {
			look.Field(context, "Importe (€)", "amount", &screens.newExpense.amount, "En euros", "")
		})
		if context.PrimaryButton("Guardar##expense") {
			screens.addExpense()
		}
	})
}

func (screens *Screens) addExpense() {
	amount, found := money.Parse(screens.newExpense.amount)
	if !found {
		screens.tell([]Refusal{{Code: CodeNegativeAmount}}, nil, "")
		return
	}
	refused, err := screens.store.AddExpense(screens.detail.Event.ID, screens.newExpense.concept, amount)
	if screens.tell(refused, err, "Gasto añadido") {
		screens.newExpense, screens.addingExpense = expenseForm{}, false
	}
}
