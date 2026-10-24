package app

import (
	"slices"
	"strings"
	"testing"

	"github.com/akrck02/go-ui/uitest"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/customers"
	"github.com/Itros97/MUSA/demo"
	"github.com/Itros97/MUSA/look"
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
	program, err := New(database, Options{DevelopmentTools: true, OpenLink: func(string) bool { return true }}) // as if a browser opened it
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
	steps(t, window, clickIn(window, before.VariantID, "Vender 1 · "+money.Format(before.Price)))
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
	steps(t, window, click(window, "Nuevo pedido"), typeInto(window, "Opcional", "Nora"), click(window, "Guardar"))
	after, _ := program.stores.Sales.All()
	if len(after) != len(before)+1 {
		t.Fatalf("%d sales, then %d", len(before), len(after))
	}
	var recorded sales.Sale
	for _, sale := range after {
		if sale.Customer == "Nora" {
			recorded = sale
		}
	}
	if recorded.Channel != sales.OnlineShop || len(recorded.Lines) != 1 || recorded.Lines[0].ProductName == "" {
		t.Errorf("recorded %+v", recorded)
	}
}

func TestAnOrderMovesAlongTheBoardAndThenToTheHistory(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Sales)
	window.Frame()
	before, _ := program.stores.Sales.All()
	onTheWay := sales.OnTheWay(before)
	index := slices.IndexFunc(onTheWay, func(sale sales.Sale) bool { return sale.Customer != "" && sale.Status == sales.ToPrepare })
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
	if _, err := window.Element("Seguir vendiendo"); err != nil {
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
	steps(t, window, click(window, "Nuevo proyecto"), click(window, "Editar redes"),
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

func TestUndoingACashSaleAtTheStandGivesItsPiecesBack(t *testing.T) {
	program, window := openApp(t, true)
	program.Show("stand")
	window.Frame()
	before, _ := program.stores.Sales.All()
	items, err := stand.WholeStock{Products: program.stores.Products}.Items()
	if err != nil || len(items) == 0 {
		t.Fatalf("no product on the stand (%v)", err)
	}
	item := items[0]
	steps(t, window, clickIn(window, item.VariantID, item.Label+" · "+money.Format(item.Price)), click(window, "Cobrar"),
		click(window, "Pagar en efectivo"), click(window, "Justo"), click(window, "Confirmar"), click(window, "Deshacer el cobro"))
	after, _ := program.stores.Sales.All()
	if len(after) != len(before) {
		t.Errorf("%d sales, then %d after undoing", len(before), len(after))
	}
	if stock, _, _ := program.stores.Products.Stock(item.VariantID); stock != item.Left {
		t.Errorf("stock %d, not the %d it had before the sale", stock, item.Left)
	}
}

func TestTheHistoryShowsAnotherPeriodChosenFromItsFigure(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Sales)
	program.sales.ShowHistory()
	window.Frame()
	lastMonth := func() bool {
		return slices.ContainsFunc(window.Texts(), func(text string) bool { return strings.Contains(text, "Mercadillo de verano") })
	}
	if lastMonth() {
		t.Fatal("this month's history already shows a fair from last month")
	}
	steps(t, window, click(window, "Periodo"), clickIn(window, "periods", "Todo"))
	if !lastMonth() {
		t.Errorf("choosing «Todo» does not show last month's fair: %q", window.Texts())
	}
}

func TestPagesCopiedFromOutsideServicesSayTheyAreNotSyncedYet(t *testing.T) {
	program, window := openApp(t, true)
	for _, page := range []Place{Projects, Sales} {
		program.Open(page)
		window.Frame()
		if !slices.Contains(window.Texts(), look.Unsynced) {
			t.Errorf("page %v does not say %q: %q", page, look.Unsynced, window.Texts())
		}
	}
}

func TestTheHomePageSwitchesBetweenTheLightAndTheDarkTheme(t *testing.T) {
	program, window := openApp(t, true)
	window.Frame()
	steps(t, window, click(window, "Modo oscuro"))
	if !program.Dark() {
		t.Fatal("«Modo oscuro» leaves the light theme on")
	}
	steps(t, window, click(window, "Modo claro"))
	if program.Dark() {
		t.Error("«Modo claro» leaves the dark theme on")
	}
}

func TestTheChosenThemeIsKeptForTheNextTimeMUSAOpens(t *testing.T) {
	program, window := openApp(t, false)
	window.Frame()
	steps(t, window, click(window, "Modo oscuro"))
	reopened, err := New(program.stores.Database, Options{Dark: false})
	if err != nil {
		t.Fatal(err)
	}
	if !reopened.Dark() {
		t.Error("MUSA opens light again after the artist chose the dark theme")
	}
}

func TestAProductCardOpensItsEditorWhenTapped(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Catalogue)
	window.Frame()
	steps(t, window, click(window, "Luna"))
	if !program.catalogue.Editing() {
		t.Error("tapping the card of «Luna» does not open its editor")
	}
}

func TestTheEditorsBackButtonIsNamedAndClosesIt(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Catalogue)
	window.Frame()
	steps(t, window, click(window, "Luna"), click(window, "Volver"))
	if program.catalogue.Editing() {
		t.Error("«Volver» leaves the editor open")
	}
}

func TestTabsAreNamedByTheirWordsWithoutTheirCount(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Catalogue)
	window.Frame()
	steps(t, window, click(window, "Digital"))
	if !slices.Contains(window.Texts(), "Procreate") { // the brush pack's format
		t.Errorf("the «Digital» tab does not show the digital products: %q", window.Texts())
	}
}

func TestOpeningAProjectsEditorLeavesThePageUnderItStill(t *testing.T) {
	program, window := openApp(t, false)
	program.Open(Projects)
	window.Frame()
	before := program.pageName()
	steps(t, window, click(window, "Nuevo proyecto"))
	if after := program.pageName(); after != before {
		t.Errorf("the page became %q from %q, so it would rise in again under the dialog", after, before)
	}
}

func TestDraggingAProjectToAnotherColumnMovesIt(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Projects)
	window.Frames(2)
	all, err := program.stores.Projects.All()
	if err != nil {
		t.Fatal(err)
	}
	id := ""
	for _, project := range all {
		if project.Title == "Retrato de familia" {
			id = project.ID
		}
	}
	card, err := window.ElementIn(id, "Retrato de familia")
	if err != nil {
		t.Fatal(err)
	}
	centreX, centreY := card.Rect.X+card.Rect.Width/2, card.Rect.Y+card.Rect.Height/2
	window.Drag(centreX, centreY, centreX+card.Rect.Width+40, centreY)
	window.Frames(2)
	all, err = program.stores.Projects.All()
	if err != nil {
		t.Fatal(err)
	}
	for _, project := range all {
		if project.Title == "Retrato de familia" && project.Stage != projects.Drawing {
			t.Errorf("the project is in %s after being dropped on «Dibujando»", project.Stage)
		}
	}
}

func TestDraggingAnOrderToTheNextColumnSendsIt(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Sales)
	window.Frames(2)
	row, err := window.Element("Marcos")
	if err != nil {
		t.Fatal(err)
	}
	window.Drag(row.Rect.X, row.Rect.Y+row.Rect.Height/2, row.Rect.X+640, row.Rect.Y+row.Rect.Height/2)
	window.Frames(2)
	all, err := program.stores.Sales.All()
	if err != nil {
		t.Fatal(err)
	}
	for _, sale := range all {
		if sale.Customer == "Marcos" && sale.Status != sales.Sent {
			t.Errorf("Marcos's order is %s after being dropped on «Enviado»", sale.Status)
		}
	}
}

func TestTappingAProjectOnTheCalendarOpensItsEditor(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Home)
	program.home.tab = 1
	window.Frames(2)
	steps(t, window, click(window, "Logo para tienda"))
	window.Frames(2)
	if program.place != Projects || program.pageName() == "" || window.Click("Pasar a entregado") != nil {
		t.Errorf("tapping the project on the calendar did not open its editor (place %v)", program.place)
	}
}

func TestTappingADayOnTheCalendarShowsWhatItHolds(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Home)
	program.home.tab = 1
	window.Frames(2)
	day := "Ver el " + calendar.LongWords(program.today.date)
	cell, err := window.Element(day)
	if err != nil {
		t.Fatal(err)
	}
	window.ClickAt(cell.Rect.X+cell.Rect.Width-4, cell.Rect.Y+4) // beside its number, away from its entries
	window.Frames(2)
	if program.home.dayShown != program.today.date {
		t.Fatalf("the day's detail is not open")
	}
	if _, err := window.ElementIn("day detail", "Nueva feria"); err != nil {
		t.Errorf("the day's detail has no «Nueva feria»: %v", err)
	}
}
