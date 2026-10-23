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
	sellSheet
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
	price     string
	problem   string
}

// stockChoice is a physical variant of the catalogue that can be taken to a fair.
type stockChoice struct {
	variantID, words string
}

// expenseForm is the new expense's fields, as typed.
type expenseForm struct{ concept, amount string }

func (screens *Screens) drawDetail(context *ui.Context) {
	detail := screens.detail
	event := detail.Event
	open := event.Status == Open
	theme := context.Theme
	look.PageHeader(context, event.Name, event.WhenAndWhere(), func() {
		if context.FlatIconButton("arrow-left", "Volver") {
			screens.detail = nil
			screens.Reload()
		}
		statusPill(context, event.Status)
		screens.detailMenu(context, event, open)
	})
	if screens.detail == nil {
		return
	}
	summary := Summarise(*detail)
	context.Box(ui.Layout{Width: ui.Grow(), Columns: look.Columns(context, 4), Gap: theme.Gap}, func() {
		look.Figure(context, "Bruto", money.Format(summary.Gross), false)
		look.Figure(context, "Coste de lo vendido", money.Format(summary.CostOfGoods), false)
		look.Figure(context, "Gastos", money.Format(summary.Expenses), false)
		look.Figure(context, "Beneficio", money.Format(summary.Profit()), summary.Profit() < 0)
	})
	screens.stockSection(context, open)
	screens.salesSection(context, open)
	screens.expensesSection(context, open)
	if screens.sheet.kind != noSheet {
		screens.drawSheet(context)
	}
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
		if context.MenuItem("trash-2", "Eliminar", "") {
			screens.sheet = sheet{kind: deleteSheet}
		}
	})
}

func (screens *Screens) stockSection(context *ui.Context, open bool) {
	theme := context.Theme
	look.Section(context, "Stock en el puesto", func() {
		if len(screens.detail.Stock) == 0 {
			look.Paragraph(context, "Aún no has llevado nada. Lleva stock del catálogo para poder vender aquí.", ui.TextStyle{Color: theme.InkSoft})
		}
		for _, line := range screens.detail.Stock {
			context.Push(line.VariantID)
			context.Box(ui.Layout{Direction: look.Across(context), Width: ui.Grow(), Gap: 8}, func() {
				context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, func() {
					context.Text(sales.ItemLabel(line.ProductName, line.VariantName), ui.TextStyle{Weight: draw2d.Bold})
					context.Text(fmt.Sprintf("Llevadas %d · vendidas %d · quedan %d · %s", line.Taken, line.Sold, line.Remaining(), money.Format(line.Price)),
						ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
				})
				if open && line.Remaining() > 0 {
					screens.lineActions(context, line)
				}
			})
			context.Pop()
		}
		if open && context.IconButton("package-plus", "Llevar stock") {
			screens.sheet = sheet{kind: takeSheet, quantity: "1", choices: screens.stockChoices()}
		}
	})
}

// lineActions are the buttons of a line still on the stand: sell some, or bring them back.
func (screens *Screens) lineActions(context *ui.Context, line StockLine) {
	context.Box(ui.Layout{Gap: 8, AlignY: ui.Center}, func() {
		if context.IconButton("shopping-bag", "Vender") {
			screens.sheet = sheet{kind: sellSheet, variantID: line.VariantID, quantity: "1", price: money.FormatForInput(line.Price)}
		}
		if context.FlatIconButton("undo-2", "Devolver") {
			screens.sheet = sheet{kind: returnSheet, variantID: line.VariantID, quantity: strconv.Itoa(line.Remaining())}
		}
	})
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

func (screens *Screens) salesSection(context *ui.Context, open bool) {
	theme := context.Theme
	look.Section(context, "Ventas", func() {
		if len(screens.detail.Sales) == 0 {
			context.Text("Sin ventas todavía.", ui.TextStyle{Color: theme.InkSoft})
		}
		for _, sale := range screens.detail.Sales {
			context.Push(sale.ID)
			line := sale.Lines[0]
			context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.Center}, func() {
				context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, ui.Look{Clip: true}, func() {
					context.Text(fmt.Sprintf("%d × %s", line.Quantity, sales.ItemLabel(line.ProductName, line.VariantName)), ui.TextStyle{})
					context.Text(sale.SoldAt.Format("15:04")+" · "+calendar.Words(calendar.Of(sale.SoldAt)), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
				})
				context.Text(money.Format(sale.Total()), ui.TextStyle{Weight: draw2d.Bold})
				if open && context.FlatIconButton("undo-2", "##Deshacer venta") {
					refused, err := screens.store.Undo(sale.ID)
					screens.tell(refused, err, "Venta deshecha")
				}
			})
			context.Pop()
		}
	})
}

func (screens *Screens) expensesSection(context *ui.Context, open bool) {
	look.Section(context, "Gastos", func() {
		for _, expense := range screens.detail.Expenses {
			context.Push(expense.ID)
			context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.Center}, func() {
				context.Box(ui.Layout{Width: ui.Grow()}, func() { context.Text(expense.Concept, ui.TextStyle{}) })
				context.Text(money.Format(expense.Amount), ui.TextStyle{Weight: draw2d.Bold})
				if open && context.FlatIconButton("trash-2", "##Quitar gasto") {
					refused, err := screens.store.DeleteExpense(screens.detail.Event.ID, expense.ID)
					screens.tell(refused, err, "Gasto quitado")
				}
			})
			context.Pop()
		}
		if !open {
			return
		}
		context.Box(ui.Layout{Direction: look.Across(context), Width: ui.Grow(), Gap: 8}, func() {
			look.Field(context, "Concepto", "concept", &screens.newExpense.concept, "Stand, gasolina, comida…", "")
			context.Box(ui.Layout{Width: ui.Fixed(120)}, func() {
				look.Field(context, "Importe (€)", "amount", &screens.newExpense.amount, "40", "")
			})
			if context.IconButton("plus", "Añadir gasto") {
				screens.addExpense()
			}
		})
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
		screens.newExpense = expenseForm{}
	}
}
