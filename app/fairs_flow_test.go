package app

import (
	"testing"

	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/money"
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
	steps(t, window, typeInto(window, "Feria del cómic", "Feria del Faro"), typeInto(window, "Pabellón 3", "Puerto"),
		click(window, "Guardar"))
	settle(window)
	if _, found := fairNamed(t, program, "Feria del Faro"); !found {
		t.Errorf("the new fair was not saved; the page shows %q", window.Texts())
	}
}

func TestSellingOneAtAFairAndUndoingIt(t *testing.T) {
	program, window, name := openFair(t)
	before, _ := fairNamed(t, program, name)
	line := before.Stock[0]
	steps(t, window, clickIn(window, line.VariantID, "Vender 1 · "+money.Format(line.Price)))
	settle(window)
	sold, _ := fairNamed(t, program, name)
	if len(sold.Sales) != len(before.Sales)+1 || sold.Stock[0].Sold != line.Sold+1 {
		t.Fatalf("selling one left %d sales and %d sold, from %d and %d", len(sold.Sales), sold.Stock[0].Sold, len(before.Sales), line.Sold)
	}
	steps(t, window, click(window, "Deshacer"))
	settle(window)
	undone, _ := fairNamed(t, program, name)
	if len(undone.Sales) != len(before.Sales) || undone.Stock[0].Sold != line.Sold {
		t.Errorf("undoing left %d sales and %d sold, not %d and %d", len(undone.Sales), undone.Stock[0].Sold, len(before.Sales), line.Sold)
	}
}

func TestAnExpenseIsAddedToAFair(t *testing.T) {
	program, window, name := openFair(t)
	before, _ := fairNamed(t, program, name)
	steps(t, window, click(window, "Gastos (1)"))
	settle(window)
	steps(t, window, typeInto(window, "Stand, gasolina, comida…", "Mesa"), typeInto(window, "40", "15"), click(window, "Añadir gasto"))
	settle(window)
	after, _ := fairNamed(t, program, name)
	if len(after.Expenses) != len(before.Expenses)+1 {
		t.Errorf("%d expenses after adding one to %d", len(after.Expenses), len(before.Expenses))
	}
}

func TestMoreStockIsTakenToAFair(t *testing.T) {
	program, window, name := openFair(t)
	before, _ := fairNamed(t, program, name)
	steps(t, window, click(window, "Llevar más stock"))
	settle(window)
	steps(t, window, click(window, "Llevar"))
	settle(window)
	after, _ := fairNamed(t, program, name)
	if taken(after) != taken(before)+1 {
		t.Errorf("%d pieces taken after taking one more to %d", taken(after), taken(before))
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
	if _, err := window.Element("Llevar más stock"); err == nil {
		t.Errorf("a closed fair still takes stock")
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
