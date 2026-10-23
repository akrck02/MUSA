package events

import (
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/sales"
	"github.com/Itros97/MUSA/storage"
)

var (
	springStart = calendar.Date{Year: 2026, Month: time.April, Day: 18}
	springEnd   = calendar.Date{Year: 2026, Month: time.April, Day: 19}
)

// bench is a catalogue with one print (variant "v1": 10 in stock, price 1000, cost 300) and an
// open fair "e1", over a database in memory.
type bench struct {
	events   *Store
	products *catalogue.Store
	sales    *sales.Store
}

func newBench(t *testing.T) bench {
	t.Helper()
	database, err := storage.Open(storage.InMemory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	counter, moment := 0, time.UnixMilli(1_700_000_000_000)
	ids := func() string { counter++; return fmt.Sprintf("id-%d", counter) }
	clock := func() time.Time { moment = moment.Add(time.Second); return moment }
	products := catalogue.NewStore(database, clock, ids)
	sold := sales.NewStore(database, products, clock, ids)
	test := bench{events: NewStore(database, sold, clock, ids), products: products, sales: sold}
	_, problems, err := products.Save(catalogue.ProductDraft{Product: catalogue.Product{ID: "p1", Kind: catalogue.Physical,
		Name: "Lámina gato", Status: catalogue.Active, Variants: []catalogue.Variant{{ID: "v1", Price: 1000, Cost: 300, Stock: 10}}}})
	if err != nil || problems != nil {
		t.Fatal(problems, err)
	}
	test.must(t)(test.events.Save(Draft{ID: "e1", Name: "Mercado de primavera", StartDate: springStart, EndDate: springEnd}))
	return test
}

// must fails the test when an action is refused or fails.
func (bench) must(t *testing.T) func([]Refusal, error) {
	return func(refused []Refusal, err error) {
		t.Helper()
		if err != nil || refused != nil {
			t.Fatalf("refused %v, %v", refused, err)
		}
	}
}

func (test bench) catalogueStock(t *testing.T) int {
	stock, _, err := test.products.Stock("v1")
	if err != nil {
		t.Fatal(err)
	}
	return stock
}

func (test bench) line(t *testing.T) StockLine {
	detail, err := test.events.Detail("e1")
	if err != nil {
		t.Fatal(err)
	}
	line, _ := detail.LineOf("v1")
	return line
}

func refusedWith(t *testing.T, want Refusal) func([]Refusal, error) {
	return func(refused []Refusal, err error) {
		t.Helper()
		if err != nil || !slices.Equal(refused, []Refusal{want}) {
			t.Errorf("refused %v, %v; want %v", refused, err, want)
		}
	}
}

func TestADraftNeedsANameAndAnEndAfterItsStart(t *testing.T) {
	draft := Draft{Name: "Feria", StartDate: springStart, EndDate: springEnd}
	if Validate(draft) != nil {
		t.Error("a valid draft is refused")
	}
	draft.EndDate = springStart
	if Validate(draft) != nil {
		t.Error("a one-day fair is refused")
	}
	if got := Validate(Draft{Name: "  ", StartDate: springEnd, EndDate: springStart}); !slices.Equal(got, []Code{CodeBlankName, CodeEndBeforeStart}) {
		t.Errorf("problems %v", got)
	}
}

func TestSavingKeepsTheStatusAndCreationTimeOfAnEditedFair(t *testing.T) {
	test := newBench(t)
	first, _, _ := test.events.Find("e1")
	if first.Status != Open {
		t.Fatalf("a new fair is %s", first.Status)
	}
	test.must(t)(test.events.Close("e1"))
	test.must(t)(test.events.Save(Draft{ID: "e1", Name: "  Otro nombre ", Location: " Pabellón 2 ", StartDate: springStart, EndDate: springEnd}))
	edited, _, _ := test.events.Find("e1")
	if edited.Status != Closed || !edited.CreatedAt.Equal(first.CreatedAt) || edited.Name != "Otro nombre" || edited.Location != "Pabellón 2" {
		t.Errorf("edited %+v", edited)
	}
	if edited.WhenAndWhere() != "18–19 abr 2026 · Pabellón 2" {
		t.Errorf("when and where %q", edited.WhenAndWhere())
	}
}

func TestStockMovesBetweenTheCatalogueAndTheStand(t *testing.T) {
	test := newBench(t)
	test.must(t)(test.events.MoveStock("e1", "v1", 6))
	if test.catalogueStock(t) != 4 || test.line(t).Taken != 6 {
		t.Fatalf("catalogue %d, taken %d", test.catalogueStock(t), test.line(t).Taken)
	}
	refusedWith(t, Refusal{Code: CodeNotEnoughStock, Available: 4})(test.events.MoveStock("e1", "v1", 5))
	refusedWith(t, Refusal{Code: CodeInvalidQuantity})(test.events.MoveStock("e1", "v1", 0))
	test.must(t)(test.events.ReturnStock("e1", "v1", 2))
	if test.catalogueStock(t) != 6 || test.line(t).Remaining() != 4 {
		t.Errorf("after returning: catalogue %d, remaining %d", test.catalogueStock(t), test.line(t).Remaining())
	}
	refusedWith(t, Refusal{Code: CodeExceedsRemaining, Available: 4})(test.events.ReturnStock("e1", "v1", 5))
}

func TestSellingAtAFairLowersTheStandButNotTheCatalogue(t *testing.T) {
	test := newBench(t)
	test.must(t)(test.events.MoveStock("e1", "v1", 6))
	test.must(t)(test.events.Sell("e1", "v1", 2, 900))
	if line := test.line(t); line.Remaining() != 4 || test.catalogueStock(t) != 4 {
		t.Errorf("remaining %d, catalogue %d", line.Remaining(), test.catalogueStock(t))
	}
	detail, _ := test.events.Detail("e1")
	sale := detail.Sales[0]
	if sale.Channel != sales.AtEvent || sale.Lines[0].UnitPrice != 900 || sale.Lines[0].UnitCost != 300 || sale.Lines[0].ProductName != "Lámina gato" {
		t.Errorf("sale %+v", sale)
	}
	refusedWith(t, Refusal{Code: CodeExceedsRemaining, Available: 4})(test.events.Sell("e1", "v1", 5, 900))
	refused, _ := test.events.Sell("e1", "v1", 0, -1)
	if !slices.Equal(refused, refusals(CodeInvalidQuantity, CodeNegativePrice)) {
		t.Errorf("a zero quantity at a negative price: %v", refused)
	}
	test.must(t)(test.events.Undo(sale.ID))
	if test.line(t).Remaining() != 6 {
		t.Errorf("after undoing: remaining %d", test.line(t).Remaining())
	}
}

func TestClosingBringsUnsoldUnitsBackAndMakesTheFairReadOnly(t *testing.T) {
	test := newBench(t)
	test.must(t)(test.events.MoveStock("e1", "v1", 6))
	test.must(t)(test.events.Sell("e1", "v1", 2, 1000))
	test.must(t)(test.events.Close("e1"))
	if test.catalogueStock(t) != 8 {
		t.Errorf("catalogue %d after closing", test.catalogueStock(t))
	}
	closed := Refusal{Code: CodeEventClosed}
	refusedWith(t, closed)(test.events.MoveStock("e1", "v1", 1))
	refusedWith(t, closed)(test.events.Sell("e1", "v1", 1, 1000))
	refusedWith(t, closed)(test.events.AddExpense("e1", "Stand", 100))
	detail, _ := test.events.Detail("e1")
	refusedWith(t, closed)(test.events.Undo(detail.Sales[0].ID))
	if summary := Summarise(detail); summary.UnitsSold != 2 || summary.Gross != 2000 {
		t.Errorf("a closed fair lost its sales: %+v", summary)
	}
}

func TestOnlyAFairWithoutSalesIsDeleted(t *testing.T) {
	test := newBench(t)
	test.must(t)(test.events.MoveStock("e1", "v1", 3))
	test.must(t)(test.events.Delete("e1"))
	if test.catalogueStock(t) != 10 {
		t.Errorf("catalogue %d after deleting", test.catalogueStock(t))
	}
	if _, found, _ := test.events.Find("e1"); found {
		t.Error("a deleted fair is read")
	}
	test.must(t)(test.events.Save(Draft{ID: "e2", Name: "Otra", StartDate: springStart, EndDate: springEnd}))
	test.must(t)(test.events.MoveStock("e2", "v1", 3))
	test.must(t)(test.events.Sell("e2", "v1", 1, 1000))
	refusedWith(t, Refusal{Code: CodeHasSales})(test.events.Delete("e2"))
}

func TestExpensesAddUpIntoTheProfit(t *testing.T) {
	test := newBench(t)
	test.must(t)(test.events.MoveStock("e1", "v1", 5))
	test.must(t)(test.events.Sell("e1", "v1", 3, 1000))
	test.must(t)(test.events.AddExpense("e1", " Stand ", 1500))
	test.must(t)(test.events.AddExpense("e1", "Gasolina", 500))
	refused, _ := test.events.AddExpense("e1", " ", -1)
	if !slices.Equal(refused, refusals(CodeBlankConcept, CodeNegativeAmount)) {
		t.Errorf("a blank negative expense: %v", refused)
	}
	detail, _ := test.events.Detail("e1")
	summary := Summarise(detail)
	if summary != (Summary{UnitsSold: 3, Gross: 3000, CostOfGoods: 900, Expenses: 2000}) || summary.Profit() != 100 {
		t.Errorf("summary %+v profit %d", summary, summary.Profit())
	}
	if detail.Expenses[0].Concept != "Stand" {
		t.Errorf("concept %q", detail.Expenses[0].Concept)
	}
	test.must(t)(test.events.DeleteExpense("e1", detail.Expenses[0].ID))
	detail, _ = test.events.Detail("e1")
	if len(detail.Expenses) != 1 {
		t.Errorf("%d expenses after deleting one", len(detail.Expenses))
	}
	if Summarise(Detail{Expenses: []Expense{{Amount: 100}}}).Profit() != -100 {
		t.Error("a fair with only expenses does not lose money")
	}
}

func TestEachErrorConstructorCarriesItsOwnCode(t *testing.T) {
	cases := map[Code]error{
		CodeEventNotFound: eventNotFound("x"),
		CodeNotTaken:      notTaken("x"),
		CodeSaleNotFound:  saleNotFound("x"),
		CodeStorageFailed: storageFailed("x", errors.New("cause")),
	}
	for code, err := range cases {
		var failure *Error
		if !errors.As(err, &failure) || failure.Code != code {
			t.Errorf("%v does not carry %s", err, code)
		}
	}
}

func TestACustomersBasketIsOneSaleThatSaysHowItWasPaid(t *testing.T) {
	test := newBench(t)
	test.must(t)(test.events.MoveStock("e1", "v1", 6))
	basket := []sales.LineDraft{{VariantID: "v1", Quantity: 2, UnitPrice: 900}, {VariantID: "v1", Quantity: 1, UnitPrice: 0}}
	test.must(t)(test.events.SellLines("e1", basket, sales.Cash))
	detail, _ := test.events.Detail("e1")
	if len(detail.Sales) != 1 || len(detail.Sales[0].Lines) != 2 || detail.Sales[0].PaidWith != sales.Cash {
		t.Fatalf("sales %+v", detail.Sales)
	}
	if test.line(t).Remaining() != 3 {
		t.Errorf("remaining %d after selling 3 of 6", test.line(t).Remaining())
	}
	tooMany := []sales.LineDraft{{VariantID: "v1", Quantity: 2, UnitPrice: 900}, {VariantID: "v1", Quantity: 2, UnitPrice: 900}}
	refusedWith(t, Refusal{Code: CodeExceedsRemaining, Available: 3})(test.events.SellLines("e1", tooMany, sales.Card))
	refusedWith(t, Refusal{Code: CodeNothingToSell})(test.events.SellLines("e1", nil, sales.Card))
}
