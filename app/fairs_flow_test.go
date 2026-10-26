package app

import (
	"testing"

	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/sales"
	"github.com/Itros97/MUSA/stand"
	"github.com/akrck02/bubble/goui"
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-ui/uitest"
)

// fairNamed is the stored fair with a name, with what it holds.
func fairNamed(test *testing.T, program *App, name string) (events.Detail, bool) {
	test.Helper()
	all, err := program.stores.Events.All()
	if err != nil {
		test.Fatal(err)
	}
	for _, event := range all {
		if event.Name == name {
			detail, err := program.stores.Events.Detail(event.ID)
			if err != nil {
				test.Fatal(err)
			}
			return detail, true
		}
	}
	return events.Detail{}, false
}

// openFair opens the fair happening today, with its stand, sales and expenses.
func openFair(test *testing.T) (*App, *uitest.Window, string) {
	test.Helper()
	program, window := openApp(test, true)
	program.Show("fair")
	settle(window)
	return program, window, "Mercado de Otoño"
}

func TestANewFairIsSaved(test *testing.T) {
	program, window := openApp(test, true)
	program.Open(Fairs)
	settle(window)
	steps(test, window, click(window, "Nuevo evento"))
	settle(window)
	steps(test, window, typeInto(window, "Ej.: Feria del cómic", "Feria del Faro"), typeInto(window, "Ej.: Pabellón 3", "Puerto"),
		click(window, "Guardar"))
	settle(window)
	if _, found := fairNamed(test, program, "Feria del Faro"); !found {
		test.Errorf("the new fair was not saved; the page shows %q", window.Texts())
	}
}

func TestASaleIsUndoneFromTheEventsPage(test *testing.T) {
	program, window, name := openFair(test)
	before, _ := fairNamed(test, program, name)
	steps(test, window, clickIn(window, before.Sales[0].ID, "##Deshacer venta"))
	settle(window)
	after, _ := fairNamed(test, program, name)
	if len(after.Sales) != len(before.Sales)-1 {
		test.Errorf("%d sales after undoing one of %d", len(after.Sales), len(before.Sales))
	}
}

func TestAnExpenseIsAddedToAFair(test *testing.T) {
	program, window, name := openFair(test)
	before, _ := fairNamed(test, program, name)
	steps(test, window, click(window, "Añadir gasto"))
	settle(window)
	steps(test, window, typeInto(window, "Stand, gasolina, comida…", "Mesa"), typeInto(window, "En euros", "15"), click(window, "Guardar"))
	settle(window)
	after, _ := fairNamed(test, program, name)
	if len(after.Expenses) != len(before.Expenses)+1 {
		test.Errorf("%d expenses after adding one to %d", len(after.Expenses), len(before.Expenses))
	}
}

// prepareOne opens the choice of what to take to the open fair and puts one piece in the case.
func prepareOne(test *testing.T, program *App, window *uitest.Window) {
	test.Helper()
	steps(test, window, click(window, "Llevar"))
	settle(window)
	if !program.stand.Preparing() || program.place != Stand {
		test.Fatalf("«Llevar» does not open the stand's grid to prepare")
	}
	items, err := stand.WholeStock{Products: program.stores.Products}.Items()
	if err != nil {
		test.Fatal(err)
	}
	pick(test, window, items, items[0])
}

func TestPackedPiecesMoveToTheEvent(test *testing.T) {
	program, window, name := openFair(test)
	before, _ := fairNamed(test, program, name)
	prepareOne(test, program, window)
	steps(test, window, click(window, "Llevar al evento"))
	settle(window)
	after, _ := fairNamed(test, program, name)
	if taken(after) != taken(before)+1 {
		test.Errorf("%d pieces taken after packing one more to %d", taken(after), taken(before))
	}
	if program.place != Fairs {
		test.Errorf("packing leaves the app on %v, not on the event", program.place)
	}
}

func TestLeavingThePreparingMovesNothing(test *testing.T) {
	program, window, name := openFair(test)
	before, _ := fairNamed(test, program, name)
	prepareOne(test, program, window)
	steps(test, window, click(window, "Volver"))
	settle(window)
	after, _ := fairNamed(test, program, name)
	if taken(after) != taken(before) || program.place != Fairs {
		test.Errorf("leaving took %d pieces and stayed on %v", taken(after)-taken(before), program.place)
	}
}

func TestWhatIsLeftOfAProductReturnsFromItsRow(test *testing.T) {
	program, window, name := openFair(test)
	before, _ := fairNamed(test, program, name)
	line := before.Stock[0]
	steps(test, window, click(window, sales.ItemLabel(line.ProductName, line.VariantName)))
	settle(window)
	steps(test, window, click(window, "Devolver"))
	settle(window)
	steps(test, window, click(window, "Devolver"))
	settle(window)
	after, _ := fairNamed(test, program, name)
	if taken(after) != taken(before)-line.Remaining() {
		test.Errorf("%d pieces taken after returning %d of %d", taken(after), line.Remaining(), taken(before))
	}
}

func taken(detail events.Detail) int {
	total := 0
	for _, line := range detail.Stock {
		total += line.Taken
	}
	return total
}

func TestClosingAFairMakesItReadOnly(test *testing.T) {
	program, window, name := openFair(test)
	steps(test, window, click(window, "Más acciones"))
	settle(window)
	steps(test, window, click(window, "Cerrar evento"))
	settle(window)
	steps(test, window, click(window, "Cerrar evento"))
	settle(window)
	closed, _ := fairNamed(test, program, name)
	if closed.Event.Status != events.Closed {
		test.Fatalf("the fair is %s after closing it", closed.Event.Status)
	}
	for _, action := range []string{"Llevar", "Abrir el puesto", "Añadir gasto", "Deshacer venta"} {
		if _, err := window.Element(action); err == nil {
			test.Errorf("a closed event still offers «%s»", action)
		}
	}
}

func TestAFairWithSalesCannotBeDeleted(test *testing.T) {
	program, window, name := openFair(test)
	steps(test, window, click(window, "Más acciones"))
	settle(window)
	steps(test, window, click(window, "Eliminar"))
	settle(window)
	steps(test, window, click(window, "Eliminar"))
	settle(window)
	if _, found := fairNamed(test, program, name); !found {
		test.Errorf("a fair with sales was deleted")
	}
}

func TestTheNextEventComesFirstAmongThoseToCome(test *testing.T) {
	program, window := openApp(test, true)
	program.Open(Fairs)
	window.Frames(2)
	soon, err := window.Element("Feria del Puerto")
	if err != nil {
		test.Fatal(err)
	}
	later, err := window.Element("Mercadillo de Invierno")
	if err != nil {
		test.Fatal(err)
	}
	if soon.Rect.Y > later.Rect.Y {
		test.Errorf("the event in three days is listed after the one in December")
	}
}

func TestANarrowWindowStacksTheEventsCardsInOneColumn(test *testing.T) {
	_, window, _ := openFair(test)
	const phoneWidth, phoneHeight = 400, 2400
	if err := window.Resize(phoneWidth, phoneHeight); err != nil {
		test.Fatal(err)
	}
	settle(window)
	var tops []float32
	for _, title := range []string{"Lo que llevas", "Ventas", "Añadir gasto"} {
		element, err := window.Element(title)
		if err != nil {
			test.Fatal(err)
		}
		tops = append(tops, element.Rect.Y)
	}
	if tops[0] >= tops[1] || tops[1] >= tops[2] {
		test.Errorf("the cards are not one under another, in order: tops %v", tops)
	}
}

func TestTheActionsOfAStockLineAreTonal(test *testing.T) {
	program, window, name := openFair(test)
	fair, _ := fairNamed(test, program, name)
	line := fair.Stock[0]
	steps(test, window, click(window, sales.ItemLabel(line.ProductName, line.VariantName)))
	settle(window)
	tonal := goui.TonalLook(program.design())
	for _, label := range []string{"Llevar más", "Devolver"} {
		button, err := window.Element(label)
		if err != nil {
			test.Fatal(err)
		}
		if !filledAt(window, button.Rect, tonal.Fill) {
			test.Errorf("«%s» is not drawn on the accent's tint", label)
		}
	}
}

// filledAt tells whether the frame drew a rectangle of this fill under the middle of the area.
func filledAt(window *uitest.Window, area draw2d.Rectangle, fill draw2d.Color) bool {
	x, y := area.X+area.Width/2, area.Y+area.Height/2
	for _, command := range window.DrawList().Commands {
		under := command.Area
		inside := x >= under.X && x <= under.X+under.Width && y >= under.Y && y <= under.Y+under.Height
		if command.Kind == draw2d.RectangleKind && command.Fill == fill && inside {
			return true
		}
	}
	return false
}
