package app

import (
	"testing"

	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/sales"
	"github.com/Itros97/MUSA/stand"
	"github.com/akrck02/go-ui/uitest"
)

// fairNamed is the stored fair with a name, with what it holds.
func fairNamed(t *testing.T, program *App, name string) (events.Detail, bool) {
	t.Helper()
	all, err := program.stores.Events.All()
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range all {
		if event.Name == name {
			detail, err := program.stores.Events.Detail(event.ID)
			if err != nil {
				t.Fatal(err)
			}
			return detail, true
		}
	}
	return events.Detail{}, false
}

// openFair opens the fair happening today, with its stand, sales and expenses.
func openFair(t *testing.T) (*App, *uitest.Window, string) {
	t.Helper()
	program, window := openApp(t, true)
	program.Show("fair")
	settle(window)
	return program, window, "Mercado de Otoño"
}

func TestANewFairIsSaved(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Fairs)
	settle(window)
	steps(t, window, click(window, "Nuevo evento"))
	settle(window)
	steps(t, window, typeInto(window, "Ej.: Feria del cómic", "Feria del Faro"), typeInto(window, "Ej.: Pabellón 3", "Puerto"),
		click(window, "Guardar"))
	settle(window)
	if _, found := fairNamed(t, program, "Feria del Faro"); !found {
		t.Errorf("the new fair was not saved; the page shows %q", window.Texts())
	}
}

func TestASaleIsUndoneFromTheEventsPage(t *testing.T) {
	program, window, name := openFair(t)
	before, _ := fairNamed(t, program, name)
	steps(t, window, clickIn(window, before.Sales[0].ID, "##Deshacer venta"))
	settle(window)
	after, _ := fairNamed(t, program, name)
	if len(after.Sales) != len(before.Sales)-1 {
		t.Errorf("%d sales after undoing one of %d", len(after.Sales), len(before.Sales))
	}
}

func TestAnExpenseIsAddedToAFair(t *testing.T) {
	program, window, name := openFair(t)
	before, _ := fairNamed(t, program, name)
	steps(t, window, click(window, "Añadir gasto"))
	settle(window)
	steps(t, window, typeInto(window, "Stand, gasolina, comida…", "Mesa"), typeInto(window, "En euros", "15"), click(window, "Guardar"))
	settle(window)
	after, _ := fairNamed(t, program, name)
	if len(after.Expenses) != len(before.Expenses)+1 {
		t.Errorf("%d expenses after adding one to %d", len(after.Expenses), len(before.Expenses))
	}
}

// packOne opens the choice of what to take to the open fair and puts one piece in the case.
func packOne(t *testing.T, program *App, window *uitest.Window) {
	t.Helper()
	steps(t, window, click(window, "Llevar"))
	settle(window)
	if !program.stand.Packing() || program.place != Stand {
		t.Fatalf("«Llevar» does not open the stand's grid to pack")
	}
	items, err := stand.WholeStock{Products: program.stores.Products}.Items()
	if err != nil {
		t.Fatal(err)
	}
	pick(t, window, items, items[0])
}

func TestPackedPiecesMoveToTheEvent(t *testing.T) {
	program, window, name := openFair(t)
	before, _ := fairNamed(t, program, name)
	packOne(t, program, window)
	steps(t, window, click(window, "Llevar al evento"))
	settle(window)
	after, _ := fairNamed(t, program, name)
	if taken(after) != taken(before)+1 {
		t.Errorf("%d pieces taken after packing one more to %d", taken(after), taken(before))
	}
	if program.place != Fairs {
		t.Errorf("packing leaves the app on %v, not on the event", program.place)
	}
}

func TestLeavingThePackingMovesNothing(t *testing.T) {
	program, window, name := openFair(t)
	before, _ := fairNamed(t, program, name)
	packOne(t, program, window)
	steps(t, window, click(window, "Volver"))
	settle(window)
	after, _ := fairNamed(t, program, name)
	if taken(after) != taken(before) || program.place != Fairs {
		t.Errorf("leaving took %d pieces and stayed on %v", taken(after)-taken(before), program.place)
	}
}

func TestWhatIsLeftOfAProductReturnsFromItsRow(t *testing.T) {
	program, window, name := openFair(t)
	before, _ := fairNamed(t, program, name)
	line := before.Stock[0]
	steps(t, window, click(window, sales.ItemLabel(line.ProductName, line.VariantName)))
	settle(window)
	steps(t, window, click(window, "Devolver"))
	settle(window)
	steps(t, window, click(window, "Devolver"))
	settle(window)
	after, _ := fairNamed(t, program, name)
	if taken(after) != taken(before)-line.Remaining() {
		t.Errorf("%d pieces taken after returning %d of %d", taken(after), line.Remaining(), taken(before))
	}
}

func taken(detail events.Detail) int {
	total := 0
	for _, line := range detail.Stock {
		total += line.Taken
	}
	return total
}

func TestClosingAFairMakesItReadOnly(t *testing.T) {
	program, window, name := openFair(t)
	steps(t, window, click(window, "Más acciones"))
	settle(window)
	steps(t, window, click(window, "Cerrar evento"))
	settle(window)
	steps(t, window, click(window, "Cerrar evento"))
	settle(window)
	closed, _ := fairNamed(t, program, name)
	if closed.Event.Status != events.Closed {
		t.Fatalf("the fair is %s after closing it", closed.Event.Status)
	}
	for _, action := range []string{"Llevar", "Abrir el puesto", "Añadir gasto", "Deshacer venta"} {
		if _, err := window.Element(action); err == nil {
			t.Errorf("a closed event still offers «%s»", action)
		}
	}
}

func TestAFairWithSalesCannotBeDeleted(t *testing.T) {
	program, window, name := openFair(t)
	steps(t, window, click(window, "Más acciones"))
	settle(window)
	steps(t, window, click(window, "Eliminar"))
	settle(window)
	steps(t, window, click(window, "Eliminar"))
	settle(window)
	if _, found := fairNamed(t, program, name); !found {
		t.Errorf("a fair with sales was deleted")
	}
}

func TestTheNextEventComesFirstAmongThoseToCome(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Fairs)
	window.Frames(2)
	soon, err := window.Element("Feria del Puerto")
	if err != nil {
		t.Fatal(err)
	}
	later, err := window.Element("Mercadillo de Invierno")
	if err != nil {
		t.Fatal(err)
	}
	if soon.Rect.Y > later.Rect.Y {
		t.Errorf("the event in three days is listed after the one in December")
	}
}

func TestANarrowWindowStacksTheEventsCardsInOneColumn(t *testing.T) {
	_, window, _ := openFair(t)
	const phoneWidth, phoneHeight = 400, 2400
	if err := window.Resize(phoneWidth, phoneHeight); err != nil {
		t.Fatal(err)
	}
	settle(window)
	var tops []float32
	for _, title := range []string{"Lo que llevas", "Ventas", "Añadir gasto"} {
		element, err := window.Element(title)
		if err != nil {
			t.Fatal(err)
		}
		tops = append(tops, element.Rect.Y)
	}
	if tops[0] >= tops[1] || tops[1] >= tops[2] {
		t.Errorf("the cards are not one under another, in order: tops %v", tops)
	}
}
