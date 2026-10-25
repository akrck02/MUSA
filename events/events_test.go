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

func newBench(test *testing.T) bench {
	test.Helper()
	database, err := storage.Open(storage.InMemory)
	if err != nil {
		test.Fatal(err)
	}
	test.Cleanup(func() { database.Close() })
	counter, moment := 0, time.UnixMilli(1_700_000_000_000)
	ids := func() string { counter++; return fmt.Sprintf("id-%d", counter) }
	clock := func() time.Time { moment = moment.Add(time.Second); return moment }
	products := catalogue.NewStore(database, clock, ids)
	sold := sales.NewStore(database, products, clock, ids)
	fixture := bench{events: NewStore(database, sold, clock, ids), products: products, sales: sold}
	_, problems, err := products.Save(catalogue.ProductDraft{Product: catalogue.Product{ID: "p1", Kind: catalogue.Physical,
		Name: "Lámina gato", Status: catalogue.Active, Variants: []catalogue.Variant{{ID: "v1", Price: 1000, Cost: 300, Stock: 10}}}})
	if err != nil || problems != nil {
		test.Fatal(problems, err)
	}
	fixture.must(test)(fixture.events.Save(Draft{ID: "e1", Name: "Mercado de primavera", StartDate: springStart, EndDate: springEnd}))
	return fixture
}

// must fails the fixture when an action is refused or fails.
func (bench) must(test *testing.T) func([]Refusal, error) {
	return func(refused []Refusal, err error) {
		test.Helper()
		if err != nil || refused != nil {
			test.Fatalf("refused %v, %v", refused, err)
		}
	}
}

func (fixture bench) catalogueStock(test *testing.T) int {
	stock, _, err := fixture.products.Stock("v1")
	if err != nil {
		test.Fatal(err)
	}
	return stock
}

func (fixture bench) line(test *testing.T) StockLine {
	detail, err := fixture.events.Detail("e1")
	if err != nil {
		test.Fatal(err)
	}
	line, _ := detail.LineOf("v1")
	return line
}

func refusedWith(test *testing.T, want Refusal) func([]Refusal, error) {
	return func(refused []Refusal, err error) {
		test.Helper()
		if err != nil || !slices.Equal(refused, []Refusal{want}) {
			test.Errorf("refused %v, %v; want %v", refused, err, want)
		}
	}
}

func TestADraftNeedsANameAndAnEndAfterItsStart(test *testing.T) {
	draft := Draft{Name: "Feria", StartDate: springStart, EndDate: springEnd}
	if Validate(draft) != nil {
		test.Error("a valid draft is refused")
	}
	draft.EndDate = springStart
	if Validate(draft) != nil {
		test.Error("a one-day fair is refused")
	}
	if got := Validate(Draft{Name: "  ", StartDate: springEnd, EndDate: springStart}); !slices.Equal(got, []Code{CodeBlankName, CodeEndBeforeStart}) {
		test.Errorf("problems %v", got)
	}
}

func TestSavingKeepsTheStatusAndCreationTimeOfAnEditedFair(test *testing.T) {
	fixture := newBench(test)
	first, _, _ := fixture.events.Find("e1")
	if first.Status != Open {
		test.Fatalf("a new fair is %s", first.Status)
	}
	fixture.must(test)(fixture.events.Close("e1"))
	fixture.must(test)(fixture.events.Save(Draft{ID: "e1", Name: "  Otro nombre ", Location: " Pabellón 2 ", StartDate: springStart, EndDate: springEnd}))
	edited, _, _ := fixture.events.Find("e1")
	if edited.Status != Closed || !edited.CreatedAt.Equal(first.CreatedAt) || edited.Name != "Otro nombre" || edited.Location != "Pabellón 2" {
		test.Errorf("edited %+v", edited)
	}
	if edited.WhenAndWhere() != "18–19 abr 2026 · Pabellón 2" {
		test.Errorf("when and where %q", edited.WhenAndWhere())
	}
}

func TestStockMovesBetweenTheCatalogueAndTheStand(test *testing.T) {
	fixture := newBench(test)
	fixture.must(test)(fixture.events.MoveStock("e1", "v1", 6))
	if fixture.catalogueStock(test) != 4 || fixture.line(test).Taken != 6 {
		test.Fatalf("catalogue %d, taken %d", fixture.catalogueStock(test), fixture.line(test).Taken)
	}
	refusedWith(test, Refusal{Code: CodeNotEnoughStock, Available: 4})(fixture.events.MoveStock("e1", "v1", 5))
	refusedWith(test, Refusal{Code: CodeInvalidQuantity})(fixture.events.MoveStock("e1", "v1", 0))
	fixture.must(test)(fixture.events.ReturnStock("e1", "v1", 2))
	if fixture.catalogueStock(test) != 6 || fixture.line(test).Remaining() != 4 {
		test.Errorf("after returning: catalogue %d, remaining %d", fixture.catalogueStock(test), fixture.line(test).Remaining())
	}
	refusedWith(test, Refusal{Code: CodeExceedsRemaining, Available: 4})(fixture.events.ReturnStock("e1", "v1", 5))
}

func TestSellingAtAFairLowersTheStandButNotTheCatalogue(test *testing.T) {
	fixture := newBench(test)
	fixture.must(test)(fixture.events.MoveStock("e1", "v1", 6))
	fixture.must(test)(fixture.events.Sell("e1", "v1", 2, 900))
	if line := fixture.line(test); line.Remaining() != 4 || fixture.catalogueStock(test) != 4 {
		test.Errorf("remaining %d, catalogue %d", line.Remaining(), fixture.catalogueStock(test))
	}
	detail, _ := fixture.events.Detail("e1")
	sale := detail.Sales[0]
	if sale.Channel != sales.AtEvent || sale.Lines[0].UnitPrice != 900 || sale.Lines[0].UnitCost != 300 || sale.Lines[0].ProductName != "Lámina gato" {
		test.Errorf("sale %+v", sale)
	}
	refusedWith(test, Refusal{Code: CodeExceedsRemaining, Available: 4})(fixture.events.Sell("e1", "v1", 5, 900))
	refused, _ := fixture.events.Sell("e1", "v1", 0, -1)
	if !slices.Equal(refused, refusals(CodeInvalidQuantity, CodeNegativePrice)) {
		test.Errorf("a zero quantity at a negative price: %v", refused)
	}
	fixture.must(test)(fixture.events.Undo(sale.ID))
	if fixture.line(test).Remaining() != 6 {
		test.Errorf("after undoing: remaining %d", fixture.line(test).Remaining())
	}
}

func TestClosingBringsUnsoldUnitsBackAndMakesTheFairReadOnly(test *testing.T) {
	fixture := newBench(test)
	fixture.must(test)(fixture.events.MoveStock("e1", "v1", 6))
	fixture.must(test)(fixture.events.Sell("e1", "v1", 2, 1000))
	fixture.must(test)(fixture.events.Close("e1"))
	if fixture.catalogueStock(test) != 8 {
		test.Errorf("catalogue %d after closing", fixture.catalogueStock(test))
	}
	closed := Refusal{Code: CodeEventClosed}
	refusedWith(test, closed)(fixture.events.MoveStock("e1", "v1", 1))
	refusedWith(test, closed)(fixture.events.Sell("e1", "v1", 1, 1000))
	refusedWith(test, closed)(fixture.events.AddExpense("e1", "Stand", 100))
	detail, _ := fixture.events.Detail("e1")
	refusedWith(test, closed)(fixture.events.Undo(detail.Sales[0].ID))
	if summary := Summarise(detail); summary.UnitsSold != 2 || summary.Gross != 2000 {
		test.Errorf("a closed fair lost its sales: %+v", summary)
	}
}

func TestOnlyAFairWithoutSalesIsDeleted(test *testing.T) {
	fixture := newBench(test)
	fixture.must(test)(fixture.events.MoveStock("e1", "v1", 3))
	fixture.must(test)(fixture.events.Delete("e1"))
	if fixture.catalogueStock(test) != 10 {
		test.Errorf("catalogue %d after deleting", fixture.catalogueStock(test))
	}
	if _, found, _ := fixture.events.Find("e1"); found {
		test.Error("a deleted fair is read")
	}
	fixture.must(test)(fixture.events.Save(Draft{ID: "e2", Name: "Otra", StartDate: springStart, EndDate: springEnd}))
	fixture.must(test)(fixture.events.MoveStock("e2", "v1", 3))
	fixture.must(test)(fixture.events.Sell("e2", "v1", 1, 1000))
	refusedWith(test, Refusal{Code: CodeHasSales})(fixture.events.Delete("e2"))
}

func TestExpensesAddUpIntoTheProfit(test *testing.T) {
	fixture := newBench(test)
	fixture.must(test)(fixture.events.MoveStock("e1", "v1", 5))
	fixture.must(test)(fixture.events.Sell("e1", "v1", 3, 1000))
	fixture.must(test)(fixture.events.AddExpense("e1", " Stand ", 1500))
	fixture.must(test)(fixture.events.AddExpense("e1", "Gasolina", 500))
	refused, _ := fixture.events.AddExpense("e1", " ", -1)
	if !slices.Equal(refused, refusals(CodeBlankConcept, CodeNegativeAmount)) {
		test.Errorf("a blank negative expense: %v", refused)
	}
	detail, _ := fixture.events.Detail("e1")
	summary := Summarise(detail)
	if summary != (Summary{UnitsSold: 3, Gross: 3000, CostOfGoods: 900, Expenses: 2000}) || summary.Profit() != 100 {
		test.Errorf("summary %+v profit %d", summary, summary.Profit())
	}
	if detail.Expenses[0].Concept != "Stand" {
		test.Errorf("concept %q", detail.Expenses[0].Concept)
	}
	fixture.must(test)(fixture.events.DeleteExpense("e1", detail.Expenses[0].ID))
	detail, _ = fixture.events.Detail("e1")
	if len(detail.Expenses) != 1 {
		test.Errorf("%d expenses after deleting one", len(detail.Expenses))
	}
	if Summarise(Detail{Expenses: []Expense{{Amount: 100}}}).Profit() != -100 {
		test.Error("a fair with only expenses does not lose money")
	}
}

func TestEachErrorConstructorCarriesItsOwnCode(test *testing.T) {
	cases := map[Code]error{
		CodeEventNotFound: eventNotFound("x"),
		CodeNotTaken:      notTaken("x"),
		CodeSaleNotFound:  saleNotFound("x"),
		CodeStorageFailed: storageFailed("x", errors.New("cause")),
	}
	for code, err := range cases {
		var failure *Error
		if !errors.As(err, &failure) || failure.Code != code {
			test.Errorf("%v does not carry %s", err, code)
		}
	}
}

func TestACustomersBasketIsOneSaleThatSaysHowItWasPaid(test *testing.T) {
	fixture := newBench(test)
	fixture.must(test)(fixture.events.MoveStock("e1", "v1", 6))
	basket := []sales.LineDraft{{VariantID: "v1", Quantity: 2, UnitPrice: 900}, {VariantID: "v1", Quantity: 1, UnitPrice: 0}}
	fixture.must(test)(fixture.events.SellLines("e1", basket, sales.Cash))
	detail, _ := fixture.events.Detail("e1")
	if len(detail.Sales) != 1 || len(detail.Sales[0].Lines) != 2 || detail.Sales[0].PaidWith != sales.Cash {
		test.Fatalf("sales %+v", detail.Sales)
	}
	if fixture.line(test).Remaining() != 3 {
		test.Errorf("remaining %d after selling 3 of 6", fixture.line(test).Remaining())
	}
	tooMany := []sales.LineDraft{{VariantID: "v1", Quantity: 2, UnitPrice: 900}, {VariantID: "v1", Quantity: 2, UnitPrice: 900}}
	refusedWith(test, Refusal{Code: CodeExceedsRemaining, Available: 3})(fixture.events.SellLines("e1", tooMany, sales.Card))
	refusedWith(test, Refusal{Code: CodeNothingToSell})(fixture.events.SellLines("e1", nil, sales.Card))
}

func TestAStoreWithItsOwnClockDatesTheSalesItRecords(test *testing.T) {
	fixture := newBench(test)
	fixture.must(test)(fixture.events.MoveStock("e1", "v1", 3))
	fairDay := springStart.At(12, time.UTC)
	fixture.must(test)(fixture.events.WithClock(func() time.Time { return fairDay }).Sell("e1", "v1", 1, 1000))
	detail, _ := fixture.events.Detail("e1")
	if !detail.Sales[0].SoldAt.Equal(fairDay) {
		test.Errorf("sold at %v, want the fair's day %v", detail.Sales[0].SoldAt, fairDay)
	}
}
