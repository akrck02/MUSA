package sales

import (
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/storage"
)

var madrid, _ = time.LoadLocation("Europe/Madrid")

func TestAnOrderAddsLinesAndShipping(t *testing.T) {
	order := Sale{
		Lines: []Line{
			{Quantity: 2, UnitPrice: 1200, UnitCost: 350},
			{Quantity: 1, UnitPrice: 300, UnitCost: 60},
		},
		ShippingCharged: 450, ShippingCost: 380,
	}
	if order.Total() != 2700+450 || order.Cost() != 760+380 || order.Margin() != 3150-1140 || order.Units() != 3 {
		t.Errorf("total %d cost %d margin %d units %d", order.Total(), order.Cost(), order.Margin(), order.Units())
	}
	loss := Sale{Lines: []Line{{Quantity: 1, UnitPrice: 100, UnitCost: 300}}}
	if loss.Margin() >= 0 {
		t.Error("a sale under cost has no loss")
	}
}

func TestADraftIsValidatedBeforeRecording(t *testing.T) {
	valid := Draft{Channel: OnlineShop, Lines: []LineDraft{{ID: "l1", VariantID: "v1", Quantity: 1, UnitPrice: 100}}}
	if problems := Validate(valid); len(problems) != 0 {
		t.Errorf("valid draft: %v", problems)
	}
	cases := map[string]struct {
		change func(*Draft)
		want   Problem
	}{
		"no lines":          {func(draft *Draft) { draft.Lines = nil }, Problem{Code: CodeNoLines}},
		"event channel":     {func(draft *Draft) { draft.Channel = AtEvent }, Problem{Code: CodeEventChannel}},
		"negative shipping": {func(draft *Draft) { draft.ShippingCost = -1 }, Problem{Code: CodeNegativeShipping}},
		"zero quantity":     {func(draft *Draft) { draft.Lines[0].Quantity = 0 }, Problem{Code: CodeInvalidQuantity, LineID: "l1"}},
		"negative price":    {func(draft *Draft) { draft.Lines[0].UnitPrice = -1 }, Problem{Code: CodeNegativePrice, LineID: "l1"}},
	}
	for name, each := range cases {
		draft := valid
		draft.Lines = slices.Clone(valid.Lines)
		each.change(&draft)
		if got := Validate(draft); !slices.Equal(got, []Problem{each.want}) {
			t.Errorf("%s: %v", name, got)
		}
	}
}

func TestTheListShowsThePeriodAndChannelNewestFirst(t *testing.T) {
	on := func(id string, day calendar.Date, hour int, channel Channel) Sale {
		return Sale{ID: id, Channel: channel, SoldAt: day.At(hour, madrid),
			Lines: []Line{{Quantity: 1, UnitPrice: 1000, UnitCost: 300}}}
	}
	all := []Sale{
		on("dec", calendar.Date{Year: 2025, Month: 12, Day: 31}, 23, OnlineShop),
		on("jan1", calendar.Date{Year: 2026, Month: 1, Day: 1}, 0, OnlineShop),
		on("jan15", calendar.Date{Year: 2026, Month: 1, Day: 15}, 12, Social),
	}
	today := calendar.Date{Year: 2026, Month: 1, Day: 15}
	ids := func(query Query) []string {
		var result []string
		for _, sale := range Filter(all, query, today, madrid) {
			result = append(result, sale.ID)
		}
		return result
	}
	cases := map[string]struct {
		query Query
		want  []string
	}{
		"this month from its first local day": {Query{Period: ThisMonth}, []string{"jan15", "jan1"}},
		"last month crosses into last year":   {Query{Period: LastMonth}, []string{"dec"}},
		"this year":                           {Query{Period: ThisYear}, []string{"jan15", "jan1"}},
		"all time":                            {Query{Period: AllTime}, []string{"jan15", "jan1", "dec"}},
		"by channel":                          {Query{Period: AllTime, Channel: Social}, []string{"jan15"}},
	}
	for name, each := range cases {
		if got := ids(each.query); !slices.Equal(got, each.want) {
			t.Errorf("%s: %v; want %v", name, got, each.want)
		}
	}
	summary := Summarise(Filter(all, Query{Period: ThisMonth}, today, madrid))
	if summary != (Summary{Count: 2, Units: 2, Revenue: 2000, Cost: 600}) || summary.Margin() != 1400 {
		t.Errorf("summary %+v", summary)
	}
	if Summarise(nil) != (Summary{}) {
		t.Error("the summary of nothing is not zero")
	}
}

func TestASaleOnAnotherDayIsStoredAtItsNoon(t *testing.T) {
	now := time.Date(2026, 1, 15, 18, 30, 0, 0, madrid)
	if got := SoldAt(calendar.Of(now), now); !got.Equal(now) {
		t.Errorf("today = %v", got)
	}
	if got := SoldAt(calendar.Date{Year: 2026, Month: 1, Day: 10}, now); !got.Equal(time.Date(2026, 1, 10, 12, 0, 0, 0, madrid)) {
		t.Errorf("another day = %v", got)
	}
}

// testStores opens a catalogue with one physical product (variant "v1", 5 in stock, cost 300) and
// one digital one (variant "d1"), and the sales beside it.
func testStores(t *testing.T) (*Store, *catalogue.Store) {
	t.Helper()
	database, err := storage.Open(storage.InMemory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	counter := 0
	ids := func() string { counter++; return fmt.Sprintf("id-%d", counter) }
	clock := func() time.Time { return time.UnixMilli(1_700_000_000_000) }
	products := catalogue.NewStore(database, clock, ids)
	for _, product := range []catalogue.Product{
		{ID: "p1", Kind: catalogue.Physical, Name: "Lámina gato", Status: catalogue.Active,
			Variants: []catalogue.Variant{{ID: "v1", Name: "A4", Price: 1000, Cost: 300, Stock: 5}}},
		{ID: "p2", Kind: catalogue.Digital, Name: "Retrato", Status: catalogue.Active,
			Variants: []catalogue.Variant{{ID: "d1", Price: 4000}}},
	} {
		if _, problems, err := products.Save(catalogue.ProductDraft{Product: product}); err != nil || problems != nil {
			t.Fatal(problems, err)
		}
	}
	return NewStore(database, products, clock, ids), products
}

func stockOf(t *testing.T, products *catalogue.Store) int {
	stock, _, err := products.Stock("v1")
	if err != nil {
		t.Fatal(err)
	}
	return stock
}

func TestRecordingASaleTakesItsUnitsAndCopiesNamesAndCost(t *testing.T) {
	store, products := testStores(t)
	sale, problems, err := store.Record(Draft{Channel: OnlineShop, Customer: " Ana ", Lines: []LineDraft{
		{ID: "l1", VariantID: "v1", Quantity: 2, UnitPrice: 900},
		{ID: "l2", VariantID: "d1", Quantity: 3, UnitPrice: 4000},
	}})
	if err != nil || problems != nil {
		t.Fatal(problems, err)
	}
	if sale.Customer != "Ana" || sale.Lines[0].ProductName != "Lámina gato" || sale.Lines[0].VariantName != "A4" || sale.Lines[0].UnitCost != 300 {
		t.Errorf("sale %+v", sale)
	}
	if got := stockOf(t, products); got != 3 {
		t.Errorf("stock %d after selling 2 of 5", got)
	}
	stored, found, _ := store.Find(sale.ID)
	if !found || len(stored.Lines) != 2 || stored.Total() != sale.Total() {
		t.Errorf("stored %+v", stored)
	}
}

func TestLinesOfOneVariantAddUpAgainstItsStock(t *testing.T) {
	store, products := testStores(t)
	_, problems, err := store.Record(Draft{Channel: OnlineShop, Lines: []LineDraft{
		{ID: "l1", VariantID: "v1", Quantity: 3, UnitPrice: 900},
		{ID: "l2", VariantID: "v1", Quantity: 3, UnitPrice: 900},
	}})
	if err != nil || !slices.Equal(problems, []Problem{{Code: CodeNotEnoughStock, VariantID: "v1", Available: 5}}) {
		t.Fatalf("problems %v %v", problems, err)
	}
	if got := stockOf(t, products); got != 5 {
		t.Errorf("a rejected sale took stock: %d", got)
	}
}

func TestDeletingASaleGivesItsUnitsBackButNotAFairsSale(t *testing.T) {
	store, products := testStores(t)
	sale, _, _ := store.Record(Draft{Channel: InPerson, Lines: []LineDraft{{ID: "l1", VariantID: "v1", Quantity: 2, UnitPrice: 900}}})
	if problems, err := store.Delete(sale.ID); err != nil || problems != nil {
		t.Fatal(problems, err)
	}
	if got := stockOf(t, products); got != 5 {
		t.Errorf("stock %d after deleting", got)
	}
	if all, _ := store.All(); len(all) != 0 {
		t.Error("a deleted sale is listed")
	}
	fair := Sale{ID: "fair", Channel: AtEvent, EventID: "e1", Lines: []Line{{ID: "fl", VariantID: "v1", Quantity: 1}}}
	if err := store.database.InTransaction(func(rows storage.Querier) error { return InsertIn(rows, fair) }); err != nil {
		t.Fatal(err)
	}
	if got := stockOf(t, products); got != 5 {
		t.Errorf("a fair's sale took catalogue stock: %d", got)
	}
	if problems, _ := store.Delete("fair"); !slices.Equal(problems, []Problem{{Code: CodeManagedByEvent}}) {
		t.Errorf("deleting a fair's sale: %v", problems)
	}
	if _, err := store.Delete("missing"); !hasCode(err, CodeSaleNotFound) {
		t.Errorf("deleting a missing sale: %v", err)
	}
}

func hasCode(err error, code Code) bool {
	var failure *Error
	return errors.As(err, &failure) && failure.Code == code
}

func TestEachErrorConstructorCarriesItsOwnCode(t *testing.T) {
	cases := map[Code]error{
		CodeSaleNotFound:    saleNotFound("x"),
		CodeVariantNotFound: variantNotFound("x"),
		CodeStorageFailed:   storageFailed("x", errors.New("cause")),
	}
	for code, err := range cases {
		if !hasCode(err, code) {
			t.Errorf("%v does not carry %s", err, code)
		}
	}
}

func TestAnOrderToSendStartsToPrepareAndALocalSaleIsDeliveredAtOnce(t *testing.T) {
	store, _ := testStores(t)
	online, _, _ := store.Record(Draft{Channel: OnlineShop, Lines: []LineDraft{{ID: "l1", VariantID: "v1", Quantity: 1, UnitPrice: 900}}})
	local, _, _ := store.Record(Draft{Channel: InPerson, Lines: []LineDraft{{ID: "l2", VariantID: "v1", Quantity: 1, UnitPrice: 900}}})
	if online.Status != ToPrepare || local.Status != Delivered {
		t.Errorf("online %q, local %q", online.Status, local.Status)
	}
	all, _ := store.All()
	if onTheWay := OnTheWay(all); len(onTheWay) != 1 || onTheWay[0].ID != online.ID {
		t.Errorf("on the way: %+v", onTheWay)
	}
	if done := Done(all); len(done) != 1 || done[0].ID != local.ID {
		t.Errorf("in the history: %+v", done)
	}
}

func TestADeliveredOrderLeavesTheBoardAndCannotMoveFurther(t *testing.T) {
	store, _ := testStores(t)
	order, _, _ := store.Record(Draft{Channel: OnlineShop, Lines: []LineDraft{{ID: "l1", VariantID: "v1", Quantity: 1, UnitPrice: 900}}})
	for _, want := range []Status{Sent, Delivered} {
		if problems, err := store.Advance(order.ID); err != nil || problems != nil {
			t.Fatal(problems, err)
		}
		if stored, _, _ := store.Find(order.ID); stored.Status != want {
			t.Fatalf("status %q, want %q", stored.Status, want)
		}
	}
	if problems, _ := store.Advance(order.ID); !slices.Equal(problems, []Problem{{Code: CodeAlreadyDelivered}}) {
		t.Errorf("advancing a delivered order: %v", problems)
	}
	if all, _ := store.All(); len(OnTheWay(all)) != 0 || len(Done(all)) != 1 {
		t.Error("a delivered order is still on the board")
	}
}

func TestAFairsSaleIsDelivered(t *testing.T) {
	store, _ := testStores(t)
	fair := Sale{ID: "fair", Channel: AtEvent, EventID: "e1", Lines: []Line{{ID: "fl", VariantID: "v1", Quantity: 1}}}
	if err := store.database.InTransaction(func(rows storage.Querier) error { return InsertIn(rows, fair) }); err != nil {
		t.Fatal(err)
	}
	if stored, _, _ := store.Find("fair"); stored.Status != Delivered {
		t.Errorf("a fair's sale is %q", stored.Status)
	}
}

func TestEveryProblemCodeHasWordsForTheArtist(t *testing.T) {
	problems := []Code{CodeNoLines, CodeEventChannel, CodeNegativeShipping, CodeManagedByEvent, CodeInvalidQuantity,
		CodeNegativePrice, CodeAlreadyDelivered}
	for _, code := range problems {
		if ProblemWords(Problem{Code: code}) == string(code) {
			t.Errorf("%s has no words", code)
		}
	}
}

func TestTheHistorySeparatesLocalSalesFromOrdersThatTravelled(t *testing.T) {
	all := []Sale{{ID: "fair", Channel: AtEvent}, {ID: "hand", Channel: InPerson}, {ID: "shop", Channel: OnlineShop}}
	local := OfKind(all, IsLocal)
	sent := OfKind(all, func(channel Channel) bool { return !IsLocal(channel) })
	if len(local) != 2 || local[0].ID != "fair" || local[1].ID != "hand" || len(sent) != 1 || sent[0].ID != "shop" {
		t.Errorf("local %v, sent %v", local, sent)
	}
}
