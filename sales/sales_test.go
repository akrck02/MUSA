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

func TestAnOrderAddsLinesAndShipping(test *testing.T) {
	order := Sale{
		Lines: []Line{
			{Quantity: 2, UnitPrice: 1200, UnitCost: 350},
			{Quantity: 1, UnitPrice: 300, UnitCost: 60},
		},
		ShippingCharged: 450, ShippingCost: 380,
	}
	if order.Total() != 2700+450 || order.Cost() != 760+380 || order.Margin() != 3150-1140 || order.Units() != 3 {
		test.Errorf("total %d cost %d margin %d units %d", order.Total(), order.Cost(), order.Margin(), order.Units())
	}
	loss := Sale{Lines: []Line{{Quantity: 1, UnitPrice: 100, UnitCost: 300}}}
	if loss.Margin() >= 0 {
		test.Error("a sale under cost has no loss")
	}
}

func TestADraftIsValidatedBeforeRecording(test *testing.T) {
	valid := Draft{Channel: OnlineShop, Lines: []LineDraft{{ID: "l1", VariantID: "v1", Quantity: 1, UnitPrice: 100}}}
	if problems := Validate(valid); len(problems) != 0 {
		test.Errorf("valid draft: %v", problems)
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
			test.Errorf("%s: %v", name, got)
		}
	}
}

func TestTheListShowsThePeriodAndChannelNewestFirst(test *testing.T) {
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
			test.Errorf("%s: %v; want %v", name, got, each.want)
		}
	}
	summary := Summarise(Filter(all, Query{Period: ThisMonth}, today, madrid))
	if summary != (Summary{Count: 2, Units: 2, Revenue: 2000, Cost: 600}) || summary.Margin() != 1400 {
		test.Errorf("summary %+v", summary)
	}
	if Summarise(nil) != (Summary{}) {
		test.Error("the summary of nothing is not zero")
	}
}

func TestASaleOnAnotherDayIsStoredAtItsNoon(test *testing.T) {
	now := time.Date(2026, 1, 15, 18, 30, 0, 0, madrid)
	if got := SoldAt(calendar.Of(now), now); !got.Equal(now) {
		test.Errorf("today = %v", got)
	}
	if got := SoldAt(calendar.Date{Year: 2026, Month: 1, Day: 10}, now); !got.Equal(time.Date(2026, 1, 10, 12, 0, 0, 0, madrid)) {
		test.Errorf("another day = %v", got)
	}
}

// testStores opens a catalogue with one physical product (variant "v1", 5 in stock, cost 300) and
// one digital one (variant "d1"), and the sales beside it.
func testStores(test *testing.T) (*Store, *catalogue.Store) {
	test.Helper()
	database, err := storage.Open(storage.InMemory)
	if err != nil {
		test.Fatal(err)
	}
	test.Cleanup(func() { database.Close() })
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
			test.Fatal(problems, err)
		}
	}
	return NewStore(database, products, clock, ids), products
}

func stockOf(test *testing.T, products *catalogue.Store) int {
	stock, _, err := products.Stock("v1")
	if err != nil {
		test.Fatal(err)
	}
	return stock
}

func TestRecordingASaleTakesItsUnitsAndCopiesNamesAndCost(test *testing.T) {
	store, products := testStores(test)
	sale, problems, err := store.Record(Draft{Channel: OnlineShop, Customer: " Ana ", Lines: []LineDraft{
		{ID: "l1", VariantID: "v1", Quantity: 2, UnitPrice: 900},
		{ID: "l2", VariantID: "d1", Quantity: 3, UnitPrice: 4000},
	}})
	if err != nil || problems != nil {
		test.Fatal(problems, err)
	}
	if sale.Customer != "Ana" || sale.Lines[0].ProductName != "Lámina gato" || sale.Lines[0].VariantName != "A4" || sale.Lines[0].UnitCost != 300 {
		test.Errorf("sale %+v", sale)
	}
	if got := stockOf(test, products); got != 3 {
		test.Errorf("stock %d after selling 2 of 5", got)
	}
	stored, found, _ := store.Find(sale.ID)
	if !found || len(stored.Lines) != 2 || stored.Total() != sale.Total() {
		test.Errorf("stored %+v", stored)
	}
}

func TestLinesOfOneVariantAddUpAgainstItsStock(test *testing.T) {
	store, products := testStores(test)
	_, problems, err := store.Record(Draft{Channel: OnlineShop, Lines: []LineDraft{
		{ID: "l1", VariantID: "v1", Quantity: 3, UnitPrice: 900},
		{ID: "l2", VariantID: "v1", Quantity: 3, UnitPrice: 900},
	}})
	if err != nil || !slices.Equal(problems, []Problem{{Code: CodeNotEnoughStock, VariantID: "v1", Available: 5}}) {
		test.Fatalf("problems %v %v", problems, err)
	}
	if got := stockOf(test, products); got != 5 {
		test.Errorf("a rejected sale took stock: %d", got)
	}
}

func TestDeletingASaleGivesItsUnitsBackButNotAFairsSale(test *testing.T) {
	store, products := testStores(test)
	sale, _, _ := store.Record(Draft{Channel: InPerson, Lines: []LineDraft{{ID: "l1", VariantID: "v1", Quantity: 2, UnitPrice: 900}}})
	if problems, err := store.Delete(sale.ID); err != nil || problems != nil {
		test.Fatal(problems, err)
	}
	if got := stockOf(test, products); got != 5 {
		test.Errorf("stock %d after deleting", got)
	}
	if all, _ := store.All(); len(all) != 0 {
		test.Error("a deleted sale is listed")
	}
	fair := Sale{ID: "fair", Channel: AtEvent, EventID: "e1", Lines: []Line{{ID: "fl", VariantID: "v1", Quantity: 1}}}
	if err := store.database.InTransaction(func(rows storage.Querier) error { return InsertIn(rows, fair) }); err != nil {
		test.Fatal(err)
	}
	if got := stockOf(test, products); got != 5 {
		test.Errorf("a fair's sale took catalogue stock: %d", got)
	}
	if problems, _ := store.Delete("fair"); !slices.Equal(problems, []Problem{{Code: CodeManagedByEvent}}) {
		test.Errorf("deleting a fair's sale: %v", problems)
	}
	if _, err := store.Delete("missing"); !hasCode(err, CodeSaleNotFound) {
		test.Errorf("deleting a missing sale: %v", err)
	}
}

func hasCode(err error, code Code) bool {
	var failure *Error
	return errors.As(err, &failure) && failure.Code == code
}

func TestEachErrorConstructorCarriesItsOwnCode(test *testing.T) {
	cases := map[Code]error{
		CodeSaleNotFound:    saleNotFound("x"),
		CodeVariantNotFound: variantNotFound("x"),
		CodeStorageFailed:   storageFailed("x", errors.New("cause")),
	}
	for code, err := range cases {
		if !hasCode(err, code) {
			test.Errorf("%v does not carry %s", err, code)
		}
	}
}

func TestAnOrderToSendStartsToPrepareAndALocalSaleIsDeliveredAtOnce(test *testing.T) {
	store, _ := testStores(test)
	online, _, _ := store.Record(Draft{Channel: OnlineShop, Lines: []LineDraft{{ID: "l1", VariantID: "v1", Quantity: 1, UnitPrice: 900}}})
	local, _, _ := store.Record(Draft{Channel: InPerson, Lines: []LineDraft{{ID: "l2", VariantID: "v1", Quantity: 1, UnitPrice: 900}}})
	if online.Status != ToPrepare || local.Status != Delivered {
		test.Errorf("online %q, local %q", online.Status, local.Status)
	}
	all, _ := store.All()
	if onTheWay := OnTheWay(all); len(onTheWay) != 1 || onTheWay[0].ID != online.ID {
		test.Errorf("on the way: %+v", onTheWay)
	}
	if done := Done(all); len(done) != 1 || done[0].ID != local.ID {
		test.Errorf("in the history: %+v", done)
	}
}

func TestADeliveredOrderLeavesTheBoardAndCannotMoveFurther(test *testing.T) {
	store, _ := testStores(test)
	order, _, _ := store.Record(Draft{Channel: OnlineShop, Lines: []LineDraft{{ID: "l1", VariantID: "v1", Quantity: 1, UnitPrice: 900}}})
	for _, want := range []Status{Sent, Delivered} {
		if problems, err := store.Advance(order.ID); err != nil || problems != nil {
			test.Fatal(problems, err)
		}
		if stored, _, _ := store.Find(order.ID); stored.Status != want {
			test.Fatalf("status %q, want %q", stored.Status, want)
		}
	}
	if problems, _ := store.Advance(order.ID); !slices.Equal(problems, []Problem{{Code: CodeAlreadyDelivered}}) {
		test.Errorf("advancing a delivered order: %v", problems)
	}
	if all, _ := store.All(); len(OnTheWay(all)) != 0 || len(Done(all)) != 1 {
		test.Error("a delivered order is still on the board")
	}
}

func TestAFairsSaleIsDelivered(test *testing.T) {
	store, _ := testStores(test)
	fair := Sale{ID: "fair", Channel: AtEvent, EventID: "e1", Lines: []Line{{ID: "fl", VariantID: "v1", Quantity: 1}}}
	if err := store.database.InTransaction(func(rows storage.Querier) error { return InsertIn(rows, fair) }); err != nil {
		test.Fatal(err)
	}
	if stored, _, _ := store.Find("fair"); stored.Status != Delivered {
		test.Errorf("a fair's sale is %q", stored.Status)
	}
}

func TestEveryProblemCodeHasWordsForTheArtist(test *testing.T) {
	problems := []Code{CodeNoLines, CodeEventChannel, CodeNegativeShipping, CodeManagedByEvent, CodeInvalidQuantity,
		CodeNegativePrice, CodeAlreadyDelivered}
	for _, code := range problems {
		if ProblemWords(Problem{Code: code}) == string(code) {
			test.Errorf("%s has no words", code)
		}
	}
}

func TestTheHistorySeparatesLocalSalesFromOrdersThatTravelled(test *testing.T) {
	all := []Sale{{ID: "fair", Channel: AtEvent}, {ID: "hand", Channel: InPerson}, {ID: "shop", Channel: OnlineShop}}
	local := OfKind(all, IsLocal)
	sent := OfKind(all, func(channel Channel) bool { return !IsLocal(channel) })
	if len(local) != 2 || local[0].ID != "fair" || local[1].ID != "hand" || len(sent) != 1 || sent[0].ID != "shop" {
		test.Errorf("local %v, sent %v", local, sent)
	}
}

func TestUnitsSinceCountsOnlyTheSalesFromThatMomentOn(test *testing.T) {
	start := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	made := []Sale{
		{SoldAt: start.Add(time.Hour), Lines: []Line{{VariantID: "brushes", Quantity: 2}, {VariantID: "wallpapers", Quantity: 1}}},
		{SoldAt: start, Lines: []Line{{VariantID: "brushes", Quantity: 3}}},
		{SoldAt: start.Add(-time.Hour), Lines: []Line{{VariantID: "brushes", Quantity: 10}}},
	}
	units := UnitsSince(made, start)
	if units["brushes"] != 5 || units["wallpapers"] != 1 {
		test.Errorf("units %v", units)
	}
}

func TestTheHistorySaysWhatEachSaleWasAndHowItEnded(test *testing.T) {
	cat := Line{ProductName: "Gato", VariantName: "A5", Quantity: 2}
	frog := Line{ProductName: "Rana", Quantity: 1}
	cases := []struct {
		sale Sale
		want string
	}{
		{Sale{Channel: AtEvent, Lines: []Line{cat, frog}}, "2 × Gato A5, 1 × Rana"},
		{Sale{Channel: InPerson, Lines: []Line{frog}}, "1 × Rana · en mano"},
		{Sale{Channel: Marketplace, Status: Delivered, Lines: []Line{cat, frog, frog}}, "4 piezas · entregado"},
		{Sale{Channel: Commission, Status: Delivered, Lines: []Line{{ProductName: "Retrato de un perro", Quantity: 1}}}, "Retrato de un perro · entregado"},
	}
	for _, each := range cases {
		if got := SummaryWords(each.sale); got != each.want {
			test.Errorf("got %q, want %q", got, each.want)
		}
	}
}
