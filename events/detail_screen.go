package events

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
	// On a wide window the stand gets a register beside it, which holds the figures instead.
	register := look.Wide(context) && screens.detailTab == 0
	if !register {
		look.Figures(context, 4, func() { screens.figures(context) })
		screens.undoBanner(context)
	}
	labels := []string{"Puesto", fmt.Sprintf("Ventas (%d)", len(detail.Sales)), fmt.Sprintf("Gastos (%d)", len(detail.Expenses))}
	context.Tabs("detail-tab", &screens.detailTab, labels...)
	switch {
	case screens.detailTab == 1:
		screens.salesSection(context, open)
	case screens.detailTab == 2:
		screens.expensesSection(context, open)
	case register:
		context.Box(ui.Layout{Width: ui.Grow(), Gap: context.Theme.Gap}, func() {
			context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: context.Theme.Gap}, func() {
				screens.stockSection(context, open, 2)
			})
			context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Fixed(registerWidth), Gap: context.Theme.Gap}, func() {
				screens.registerPanel(context, open)
			})
		})
	default:
		screens.stockSection(context, open, look.Columns(context, 2))
	}
	if screens.sheet.kind != noSheet {
		screens.drawSheet(context)
	}
}

// registerWidth is how wide the register beside the stand is, in pixels.
const registerWidth = 340

// registerShown is how many of the latest sales the register lists; the Ventas tab has them all.
const registerShown = 6

func (screens *Screens) figures(context *ui.Context) {
	summary := Summarise(*screens.detail)
	look.Figure(context, "Bruto", money.Format(summary.Gross), false)
	look.Figure(context, "Coste de lo vendido", money.Format(summary.CostOfGoods), false)
	look.Figure(context, "Gastos", money.Format(summary.Expenses), false)
	look.Figure(context, "Beneficio", money.Format(summary.Profit()), summary.Profit() < 0)
}

// registerPanel is the till beside the stand, as on a shop's tablet: the fair's figures, the
// undo of the last sale and the latest sales, all in sight while selling.
func (screens *Screens) registerPanel(context *ui.Context, open bool) {
	look.Figures(context, 2, func() { screens.figures(context) })
	screens.undoBanner(context)
	look.Section(context, "Últimas ventas", func() {
		if len(screens.detail.Sales) == 0 {
			context.Text("Sin ventas todavía.", ui.TextStyle{Color: context.Theme.InkSoft})
		}
		for index, sale := range screens.detail.Sales {
			if index == registerShown {
				break
			}
			screens.saleRow(context, sale, open)
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

// undoTime is how long the note of a quick sale stays, with its undo button.
const undoTime = 8 * time.Second

// stockSection is the stand: a card per product taken, each sold with one touch at its price,
// which is how a sale happens at a fair; a different quantity or price is one touch further.
func (screens *Screens) stockSection(context *ui.Context, open bool, columns int) {
	theme := context.Theme
	if len(screens.detail.Stock) == 0 {
		if look.Empty(context, "package-plus", "El puesto está vacío",
			"Lleva stock del catálogo para poder vender aquí. Lo que no vendas vuelve al cerrar el evento.",
			map[bool]string{true: "Llevar stock", false: ""}[open]) {
			screens.sheet = sheet{kind: takeSheet, quantity: "1", choices: screens.stockChoices()}
		}
		return
	}
	context.Box(ui.Layout{Width: ui.Grow(), Columns: columns, Gap: theme.Gap}, func() {
		for _, line := range screens.detail.Stock {
			context.Push(line.VariantID)
			screens.standCard(context, line, open)
			context.Pop()
		}
	})
	if open && context.IconButton("package-plus", "Llevar más stock") {
		screens.sheet = sheet{kind: takeSheet, quantity: "1", choices: screens.stockChoices()}
	}
}

func (screens *Screens) standCard(context *ui.Context, line StockLine, open bool) {
	theme := context.Theme
	look.Card(context, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(14), Gap: 8}, func() {
		context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, ui.Look{Clip: true}, func() {
			context.Text(sales.ItemLabel(line.ProductName, line.VariantName), ui.TextStyle{Weight: draw2d.Bold})
			context.Text(fmt.Sprintf("Quedan %d de %d · %d vendidas", line.Remaining(), line.Taken, line.Sold),
				ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
		})
		look.Meter(context, line.Sold, line.Taken)
		if !open {
			return
		}
		if line.Remaining() == 0 {
			look.Pill(context, "Agotado en el puesto", look.Neutral(theme))
			return
		}
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, AlignY: ui.Center}, func() {
			if context.IconButton("shopping-bag", "Vender 1 · "+money.Format(line.Price)) {
				screens.quickSell(line)
			}
			context.Spacer()
			context.MenuButton("ellipsis", "##Más", func() {
				if context.MenuItem("calculator", "Otra cantidad o precio", "") {
					screens.sheet = sheet{kind: sellSheet, variantID: line.VariantID, quantity: "1", price: money.FormatForInput(line.Price)}
				}
				if context.MenuItem("undo-2", "Devolver al catálogo", "") {
					screens.sheet = sheet{kind: returnSheet, variantID: line.VariantID, quantity: strconv.Itoa(line.Remaining())}
				}
			})
		})
	})
}

// quickSell sells one unit at the list price, and offers to undo it for a few seconds: a mistaken
// touch costs one more touch, not a question before every sale.
func (screens *Screens) quickSell(line StockLine) {
	refused, err := screens.store.Sell(screens.detail.Event.ID, line.VariantID, 1, line.Price)
	if !screens.tell(refused, err, "") {
		return
	}
	screens.noteUntil = time.Time{} // the undo banner says it instead
	if len(screens.detail.Sales) > 0 {
		screens.lastSale = screens.detail.Sales[0].ID
		screens.lastWords = "Vendido: " + sales.ItemLabel(line.ProductName, line.VariantName) + " · " + money.Format(line.Price)
		screens.lastUntil = time.Now().Add(undoTime)
	}
}

func (screens *Screens) undoBanner(context *ui.Context) {
	if screens.lastSale == "" || time.Now().After(screens.lastUntil) {
		return
	}
	theme := context.Theme
	context.RequestFrame(time.Until(screens.lastUntil))
	tone := look.Positive(theme)
	context.Frame(ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(14, 8), Gap: 10, AlignY: ui.Center}, ui.Look{Fill: tone.Fill, Radius: theme.PanelRadius}, func() {
		context.IconSized("circle-check", theme.IconSize, theme.Positive)
		context.Frame(ui.Layout{Width: ui.Grow()}, ui.Look{Clip: true}, func() { context.Text(screens.lastWords, ui.TextStyle{Weight: draw2d.Bold}) })
		if context.FlatIconButton("undo-2", "Deshacer") {
			refused, err := screens.store.Undo(screens.lastSale)
			screens.lastSale = ""
			screens.tell(refused, err, "Venta deshecha")
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
			screens.saleRow(context, sale, open)
		}
	})
}

// saleRow is one sale of the fair: what, when and how much, with its undo while the fair is open.
func (screens *Screens) saleRow(context *ui.Context, sale sales.Sale, open bool) {
	theme := context.Theme
	context.Push(sale.ID)
	defer context.Pop()
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
}

func (screens *Screens) expensesSection(context *ui.Context, open bool) {
	look.Section(context, "Gastos", func() {
		for _, expense := range screens.detail.Expenses {
			context.Push(expense.ID)
			context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.Center}, func() {
				context.Box(ui.Layout{Width: ui.Grow()}, func() { context.Text(expense.Concept, ui.TextStyle{}) })
				context.Text(money.Format(expense.Amount), ui.TextStyle{Weight: draw2d.Bold})
				if open && context.DangerFlatIconButton("trash-2", "##Quitar gasto") {
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
				look.Field(context, "Importe (€)", "amount", &screens.newExpense.amount, "En euros", "")
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
