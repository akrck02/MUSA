package app

import (
	"slices"
	"strings"
	"testing"

	"github.com/akrck02/go-ui/uitest"

	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/customers"
	"github.com/Itros97/MUSA/demo"
	"github.com/Itros97/MUSA/money"
	"github.com/Itros97/MUSA/projects"
	"github.com/Itros97/MUSA/sales"
	"github.com/Itros97/MUSA/stand"
	"github.com/Itros97/MUSA/storage"
)

// openApp opens MUSA over a database in memory, filled with the demo when withDemo, in a test
// window the size of a laptop's.
func openApp(t *testing.T, withDemo bool) (*App, *uitest.Window) {
	t.Helper()
	database, err := storage.Open(storage.InMemory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	program, err := New(database, Options{DevelopmentTools: true})
	if err != nil {
		t.Fatal(err)
	}
	if withDemo {
		if err := demo.Fill(program.Stores()); err != nil {
			t.Fatal(err)
		}
	}
	window, err := uitest.New(program.Build, uitest.Size(1280, 1600), uitest.Test(t))
	if err != nil {
		t.Fatal(err)
	}
	window.Frame()
	return program, window
}

func steps(t *testing.T, window *uitest.Window, script ...func() error) {
	t.Helper()
	for index, step := range script {
		if err := step(); err != nil {
			t.Fatalf("step %d: %v", index+1, err)
		}
	}
}

func typeInto(window *uitest.Window, placeholder, text string) func() error {
	return func() error {
		if err := window.Click(placeholder); err != nil {
			return err
		}
		window.Type(text)
		return nil
	}
}

func click(window *uitest.Window, label string) func() error {
	return func() error { return window.Click(label) }
}

func TestCreatingAProductPutsItInTheCatalogue(t *testing.T) {
	program, window := openApp(t, false)
	program.Open(Catalogue)
	window.Frame()
	steps(t, window,
		click(window, "Crear tu primer producto"),
		click(window, "Producto físico"),
		typeInto(window, "Gato astronauta", "Lámina de prueba"),
		typeInto(window, "12,50", "15"),
		click(window, "Guardar"),
	)
	products, err := program.stores.Products.All()
	if err != nil || len(products) != 1 {
		t.Fatalf("%d products, %v", len(products), err)
	}
	if product := products[0]; product.Name != "Lámina de prueba" || product.Kind != catalogue.Physical || product.Variants[0].Price != 1500 {
		t.Errorf("saved %+v", product)
	}
	if _, err := window.Element("Lámina de prueba"); err != nil {
		t.Error("the new product is not in the list:", err)
	}
}

func TestAProductWithoutANameIsNotSaved(t *testing.T) {
	program, window := openApp(t, false)
	program.Open(Catalogue)
	window.Frame()
	steps(t, window, click(window, "Crear tu primer producto"),
		click(window, "Producto digital"), click(window, "Guardar"))
	if _, err := window.Element("Ponle un nombre"); err != nil {
		t.Error("the editor does not say what is missing:", err)
	}
	if products, _ := program.stores.Products.All(); len(products) != 0 {
		t.Errorf("%d products saved", len(products))
	}
}

func TestSellingAtAFairLeavesFewerOnTheStand(t *testing.T) {
	program, window := openApp(t, true)
	program.Show("fair")
	window.Frame()
	before := program.fairs.Detail().Stock[0]
	// One touch sells one at the stand's price, with no sheet to confirm: undo is offered instead.
	steps(t, window, click(window, "Vender 1 · "+money.Format(before.Price)))
	after := program.fairs.Detail().Stock[0]
	if after.Remaining() != before.Remaining()-1 {
		t.Errorf("remaining %d, then %d", before.Remaining(), after.Remaining())
	}
}

func TestRecordingAnOrderTakesItsStockFromTheCatalogue(t *testing.T) {
	program, window := openApp(t, true)
	before, _ := program.stores.Sales.All()
	program.Open(Sales)
	window.Frame()
	steps(t, window, click(window, "Nuevo pedido"), typeInto(window, "Opcional", "Ana"), click(window, "Guardar"))
	after, _ := program.stores.Sales.All()
	if len(after) != len(before)+1 {
		t.Fatalf("%d sales, then %d", len(before), len(after))
	}
	var recorded sales.Sale
	for _, sale := range after {
		if sale.Customer == "Ana" {
			recorded = sale
		}
	}
	if recorded.Channel != sales.OnlineShop || len(recorded.Lines) != 1 || !strings.Contains(recorded.Lines[0].ProductName, "Bosque") {
		t.Errorf("recorded %+v", recorded)
	}
}

func TestAnOrderMovesAlongTheBoardAndThenToTheHistory(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Sales)
	window.Frame()
	before, _ := program.stores.Sales.All()
	onTheWay := sales.OnTheWay(before)
	index := slices.IndexFunc(onTheWay, func(sale sales.Sale) bool { return sale.Customer != "" })
	if index < 0 {
		t.Fatal("the demo has no order with a customer on the way")
	}
	order := onTheWay[index]
	steps(t, window, clickIn(window, order.ID, "Preparado"), clickIn(window, order.ID, "Entregado"))
	after, found, err := program.stores.Sales.Find(order.ID)
	if err != nil || !found || after.Status != sales.Delivered {
		t.Fatalf("the order is %q (%v)", after.Status, err)
	}
	steps(t, window, click(window, "Historial de ventas"))
	if !slices.ContainsFunc(window.Texts(), func(text string) bool { return strings.Contains(text, order.Customer) }) {
		t.Errorf("the history does not show %s's order: %q", order.Customer, window.Texts())
	}
}

// clickIn clicks the control with label inside the scope pushed with name, for labels a list repeats.
func clickIn(window *uitest.Window, scope, label string) func() error {
	return func() error {
		element, err := window.ElementIn(scope, label)
		if err != nil {
			return err
		}
		window.ClickAt(element.Rect.X+element.Rect.Width/2, element.Rect.Y+element.Rect.Height/2)
		window.Frame()
		return nil
	}
}

func TestPayingInCashAtTheStandRecordsALocalSaleAndTakesTheStock(t *testing.T) {
	program, window := openApp(t, true)
	program.Show("stand")
	window.Frame()
	before, _ := program.stores.Sales.All()
	items, err := stand.WholeStock{Products: program.stores.Products}.Items()
	if err != nil || len(items) == 0 {
		t.Fatalf("no product on the stand (%v)", err)
	}
	item := items[0]
	steps(t, window, clickIn(window, item.VariantID, item.Label+" · "+money.Format(item.Price)), click(window, "Cobrar"), click(window, "Pagar en efectivo"),
		click(window, money.Format(5000)), click(window, "Confirmar"))
	if _, err := window.Element("Siguiente cliente"); err != nil {
		t.Fatal("the sale did not close:", err)
	}
	after, _ := program.stores.Sales.All()
	if len(after) != len(before)+1 {
		t.Fatalf("%d sales, then %d", len(before), len(after))
	}
	sold := after[0]
	if sold.Channel != sales.InPerson || sold.Status != sales.Delivered || sold.PaidWith != sales.Cash || sold.Units() != 1 {
		t.Errorf("recorded %+v", sold)
	}
	if stock, _, _ := program.stores.Products.Stock(item.VariantID); stock != item.Left-1 {
		t.Errorf("stock %d after selling 1 of %d", stock, item.Left)
	}
}

func TestGoingToTheStandFromTheHomePageOpensTodaysFair(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Home) // read the day again, now that the demo is in
	window.Frame()
	if !program.today.hasFair {
		t.Fatal("the demo has no fair happening today")
	}
	steps(t, window, click(window, "Ir al puesto"))
	if program.place != Stand {
		t.Fatalf("the place is %v, not the stand", program.place)
	}
	if _, err := window.Element(program.today.fair.Name); err != nil {
		t.Error("the stand does not show the fair:", err)
	}
}

func TestANewProjectIsSavedWithItsCustomerAndTheirNetwork(t *testing.T) {
	program, window := openApp(t, false)
	program.Open(Projects)
	window.Frame()
	steps(t, window, click(window, "Nuevo proyecto"),
		typeInto(window, "Retrato de familia", "Retrato de Toby"),
		typeInto(window, "Lucía", "Marta"),
		typeInto(window, "@instagram", "@marta"),
		typeInto(window, "40", "35"),
		click(window, "Guardar"))
	saved, err := program.stores.Projects.All()
	if err != nil || len(saved) != 1 || saved[0].Title != "Retrato de Toby" || saved[0].Price != 3500 || saved[0].Stage != projects.ToStart {
		t.Fatalf("saved %+v, %v", saved, err)
	}
	person, _, _ := program.stores.Customers.Find(saved[0].CustomerID)
	if handle, found := person.Handle("instagram"); person.Name != "Marta" || !found || handle != "@marta" {
		t.Errorf("customer %+v", person)
	}
}

func TestTheArtistCanAddASocialNetwork(t *testing.T) {
	program, window := openApp(t, false)
	program.Open(Projects)
	window.Frame()
	steps(t, window, click(window, "Redes"),
		typeInto(window, "Nueva red", "Mastodon"),
		typeInto(window, "https://ejemplo.com/{user}", "https://mastodon.social/@{user}"),
		click(window, "Guardar redes"))
	networks, err := program.stores.Customers.Networks()
	if err != nil {
		t.Fatal(err)
	}
	found := slices.ContainsFunc(networks, func(network customers.Network) bool {
		return network.Name == "Mastodon" && network.Link("@ana") == "https://mastodon.social/@ana"
	})
	if !found {
		t.Errorf("networks %+v", networks)
	}
}
