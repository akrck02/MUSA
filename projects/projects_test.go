package projects

import (
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/customers"
	"github.com/Itros97/MUSA/sales"
	"github.com/Itros97/MUSA/storage"
)

type bench struct {
	projects *Store
	sales    *sales.Store
	customer customers.Customer
	digital  string // a digital product's variant
	physical string // a physical product's variant
}

func newBench(test *testing.T) bench {
	test.Helper()
	database, err := storage.Open(storage.InMemory)
	if err != nil {
		test.Fatal(err)
	}
	test.Cleanup(func() { database.Close() })
	counter := 0
	ids := func() string { counter++; return fmt.Sprint("id", counter) }
	now := func() time.Time { return time.UnixMilli(1_700_000_000_000) }
	products := catalogue.NewStore(database, now, ids)
	people := customers.NewStore(database, now, ids)
	fixture := bench{projects: NewStore(database, products, people, now, ids), sales: sales.NewStore(database, products, now, ids)}
	fixture.customer, _, _ = people.Save(customers.Customer{Name: "Ana"})
	portrait, problems, err := products.Save(catalogue.ProductDraft{Product: catalogue.Product{ID: "portrait", Kind: catalogue.Digital, Name: "Portrait",
		Status: catalogue.Active, DeliveryFormat: "PNG", Variants: []catalogue.Variant{{ID: "bust", Name: "Bust", Price: 4000, Cost: 500}}}})
	if err != nil || problems != nil {
		test.Fatal(problems, err)
	}
	print, problems, err := products.Save(catalogue.ProductDraft{Product: catalogue.Product{ID: "print", Kind: catalogue.Physical, Name: "Print",
		Status: catalogue.Active, Variants: []catalogue.Variant{{ID: "print", Price: 1000, Stock: 3}}}})
	if err != nil || problems != nil {
		test.Fatal(problems, err)
	}
	fixture.digital, fixture.physical = portrait.Variants[0].ID, print.Variants[0].ID
	return fixture
}

func TestANewProjectStartsAtTheBeginningOfTheBoard(test *testing.T) {
	fixture := newBench(test)
	project, problems, err := fixture.projects.Save(Project{Title: " Family portrait ", CustomerID: fixture.customer.ID, Price: 4000, VariantID: fixture.digital})
	if err != nil || problems != nil {
		test.Fatal(problems, err)
	}
	if project.Stage != ToStart || project.Title != "Family portrait" {
		test.Errorf("saved %+v", project)
	}
}

func TestAProjectNeedsATitleACustomerAndADigitalProduct(test *testing.T) {
	fixture := newBench(test)
	_, problems, _ := fixture.projects.Save(Project{Price: -1, VariantID: fixture.physical})
	want := []Code{CodeBlankTitle, CodeNoCustomer, CodeNegativePrice, CodeProductNotDigital}
	if !slices.Equal(problems, want) {
		test.Errorf("problems %v, want %v", problems, want)
	}
}

func TestDeliveringAProjectRecordsItsSaleOnce(test *testing.T) {
	fixture := newBench(test)
	project, _, _ := fixture.projects.Save(Project{Title: "Portrait of Ana", CustomerID: fixture.customer.ID, Price: 4000, VariantID: fixture.digital})
	for _, stage := range []Stage{Drawing, ToReview, Delivered} {
		if problems, err := fixture.projects.Move(project.ID, stage); err != nil || problems != nil {
			test.Fatal(stage, problems, err)
		}
	}
	all, _ := fixture.sales.All()
	if len(all) != 1 || all[0].Channel != sales.Commission || all[0].Total() != 4000 || all[0].Cost() != 500 || all[0].Customer != "Ana" {
		test.Fatalf("sales %+v", all)
	}
	if problems, _ := fixture.projects.Delete(project.ID); !slices.Equal(problems, []Code{CodeAlreadyDelivered}) {
		test.Errorf("deleting a delivered project: %v", problems)
	}
}

// A project marked delivered by mistake goes back to its column, and its sale leaves the history,
// so delivering it again records it once, not twice.
func TestMovingADeliveredProjectBackTakesItsSaleBack(test *testing.T) {
	fixture := newBench(test)
	project, _, _ := fixture.projects.Save(Project{Title: "Portrait of Ana", CustomerID: fixture.customer.ID, Price: 4000, VariantID: fixture.digital})
	for _, stage := range []Stage{Delivered, ToReview, Delivered} {
		if problems, err := fixture.projects.Move(project.ID, stage); err != nil || problems != nil {
			test.Fatal(stage, problems, err)
		}
	}
	if all, _ := fixture.sales.All(); len(all) != 1 {
		test.Fatalf("%d sales after delivering, taking back and delivering again", len(all))
	}
	if _, err := fixture.projects.Move(project.ID, Drawing); err != nil {
		test.Fatal(err)
	}
	moved, _, _ := fixture.projects.Find(project.ID)
	if all, _ := fixture.sales.All(); len(all) != 0 || moved.Stage != Drawing || moved.SaleID != "" {
		test.Errorf("after moving back: %d sales, stage %s, sale %q", len(all), moved.Stage, moved.SaleID)
	}
}

func TestAProjectWithoutAProductIsSoldByItsTitle(test *testing.T) {
	fixture := newBench(test)
	project, _, _ := fixture.projects.Save(Project{Title: "Shop logo", CustomerID: fixture.customer.ID, Price: 6000})
	if _, err := fixture.projects.Move(project.ID, Delivered); err != nil {
		test.Fatal(err)
	}
	all, _ := fixture.sales.All()
	if len(all) != 1 || all[0].Lines[0].ProductName != "Shop logo" || all[0].Cost() != 0 {
		test.Errorf("sales %+v", all)
	}
}

func TestADeadlineSaysHowFarAwayItIs(test *testing.T) {
	today := calendar.Date{Year: 2026, Month: 5, Day: 10}
	cases := map[int]string{0: "Vence hoy", 1: "Vence mañana", 6: "Vence en 6 días", -1: "Venció ayer", -3: "Venció hace 3 días"}
	for days, want := range cases {
		if got := DeadlineWords(today.AddDays(days), today); got != want {
			test.Errorf("%d days: %q, want %q", days, got, want)
		}
	}
	if DeadlineWords(calendar.Date{}, today) != "" {
		test.Error("no deadline has words")
	}
	late := Project{Stage: Drawing, Deadline: today.AddDays(-1)}
	if !late.Late(today) || (Project{Stage: Delivered, Deadline: today.AddDays(-1)}).Late(today) {
		test.Error("lateness is wrong")
	}
}

func TestEveryProblemHasWordsAndEachFailureItsCode(test *testing.T) {
	for _, code := range []Code{CodeBlankTitle, CodeNoCustomer, CodeNegativePrice, CodeProductNotDigital, CodeAlreadyDelivered} {
		if ProblemWords(code) == string(code) {
			test.Errorf("%s has no words", code)
		}
	}
	for code, err := range map[Code]error{CodeProjectNotFound: projectNotFound("x"), CodeStorageFailed: storageFailed("x", errors.New("cause"))} {
		var failure *Error
		if !errors.As(err, &failure) || failure.Code != code {
			test.Errorf("%v does not carry %s", err, code)
		}
	}
}

func TestADeliveredProjectSaysTheDayItWasDelivered(fixture *testing.T) {
	delivered := time.Date(2026, time.October, 18, 17, 0, 0, 0, time.UTC)
	if words := DeliveredWords(delivered); words != "Entregado el 18" {
		fixture.Errorf("got %q", words)
	}
	if words := DeliveredWords(time.Time{}); words != "Entregado" {
		fixture.Errorf("without a day, got %q", words)
	}
}

func TestTheBoardCountsItsProjectsInWords(fixture *testing.T) {
	if projectCount(1) != "1 proyecto" || projectCount(5) != "5 proyectos" {
		fixture.Errorf("got %q and %q", projectCount(1), projectCount(5))
	}
}

func TestASearchFindsProjectsByTitleOrCustomerWhateverTheAccents(test *testing.T) {
	lighthouse := Project{Title: "Retrato del faro"}
	cases := []struct {
		query string
		want  bool
	}{{"", true}, {"FARO", true}, {"ines", true}, {"barco", false}}
	for _, each := range cases {
		if got := Matches(lighthouse, "Inés", each.query); got != each.want {
			test.Errorf("Matches(%q) = %v, want %v", each.query, got, each.want)
		}
	}
}
