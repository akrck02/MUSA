package app

import (
	"fmt"
	"slices"
	"testing"

	"github.com/Itros97/MUSA/money"
	"github.com/Itros97/MUSA/sales"
	"github.com/Itros97/MUSA/stand"
	"github.com/akrck02/go-ui/uitest"
)

// openStand opens the stand on the whole stock with one piece of its first item in the basket.
func openStand(test *testing.T) (*App, *uitest.Window, stand.Item) {
	test.Helper()
	program, window := openApp(test, true)
	program.Show("stand")
	settle(window)
	items, err := stand.WholeStock{Products: program.stores.Products}.Items()
	if err != nil || len(items) == 0 {
		test.Fatalf("no product on the stand (%v)", err)
	}
	item := items[0]
	pick(test, window, items, item)
	return program, window, item
}

// pick puts one of an item in the basket the way the artist does: tapping its product's tile and,
// when the product has several versions, choosing the item's version.
func pick(test *testing.T, window *uitest.Window, items []stand.Item, item stand.Item) {
	test.Helper()
	versions := 0
	for _, each := range items {
		if each.ProductID == item.ProductID {
			versions++
		}
	}
	if versions == 1 {
		steps(test, window, clickIn(window, item.ProductID, item.Label+" · "+money.Format(item.Price)))
		settle(window)
		return
	}
	steps(test, window, clickIn(window, item.ProductID, fmt.Sprintf("%s · %d versiones", item.Name, versions)))
	settle(window)
	steps(test, window, click(window, item.Version))
	settle(window)
}

func salesCount(test *testing.T, program *App) int {
	test.Helper()
	all, err := program.stores.Sales.All()
	if err != nil {
		test.Fatal(err)
	}
	return len(all)
}

func TestPayingByCardRecordsACardSale(test *testing.T) {
	program, window, _ := openStand(test)
	before := salesCount(test, program)
	steps(test, window, click(window, "Cobrar"), click(window, "Pagar con tarjeta"), click(window, "Cobrado con tarjeta"))
	settle(window)
	all, _ := program.stores.Sales.All()
	if len(all) != before+1 || all[0].PaidWith != sales.Card {
		test.Errorf("after a card payment: %d sales from %d, the last paid with %s", len(all), before, all[0].PaidWith)
	}
}

func TestACardThatFailsRecordsNothing(test *testing.T) {
	program, window, _ := openStand(test)
	before := salesCount(test, program)
	steps(test, window, click(window, "Cobrar"), click(window, "Pagar con tarjeta"), click(window, "No ha salido"))
	settle(window)
	if after := salesCount(test, program); after != before {
		test.Errorf("a failed card payment recorded a sale")
	}
}

func TestExactCashClosesTheSale(test *testing.T) {
	program, window, _ := openStand(test)
	before := salesCount(test, program)
	steps(test, window, click(window, "Cobrar"), click(window, "Pagar en efectivo"), click(window, "Justo"), click(window, "Confirmar"))
	settle(window)
	if after := salesCount(test, program); after != before+1 {
		test.Errorf("paying the exact amount did not record the sale")
	}
}

func TestTooLittleCashDoesNotCloseTheSale(test *testing.T) {
	program, window, item := openStand(test)
	if item.Price <= 50 {
		test.Skip("the first item costs too little to pay short")
	}
	before := salesCount(test, program)
	steps(test, window, click(window, "Cobrar"), click(window, "Pagar en efectivo"), click(window, money.Format(50)), click(window, "Confirmar"))
	settle(window)
	if after := salesCount(test, program); after != before {
		test.Errorf("a sale was recorded with too little money")
	}
	if !slices.Contains(window.Texts(), "Falta") {
		test.Errorf("the dialog does not say money is missing")
	}
}

func TestClearingTheCashTypedStartsAgain(test *testing.T) {
	_, window, _ := openStand(test)
	steps(test, window, click(window, "Cobrar"), click(window, "Pagar en efectivo"), click(window, money.Format(5000)), click(window, "Borrar"))
	settle(window)
	if slices.Contains(window.Texts(), "Cambio") && slices.Contains(window.Texts(), money.Format(5000)) {
		test.Errorf("the money given is still there after Borrar")
	}
}

func TestTakingOneOutOfTheBasketEmptiesIt(test *testing.T) {
	_, window, item := openStand(test)
	steps(test, window, clickIn(window, item.ProductID, item.Name+" · 3 versiones"))
	settle(window)
	steps(test, window, click(window, "Quitar uno"))
	settle(window)
	if _, err := window.Element("Cobrar"); err == nil {
		test.Errorf("the basket bar still shows after taking its only piece out")
	}
}

func TestSearchingTheStandKeepsOnlyWhatMatches(test *testing.T) {
	program, window := openApp(test, true)
	program.Show("stand")
	settle(window)
	steps(test, window, click(window, "Buscar"))
	settle(window)
	window.Type("zorro")
	settle(window)
	for _, text := range window.Texts() {
		if text == "Estrella" {
			test.Fatalf("«Estrella» still shows when searching «zorro»")
		}
	}
	if !slices.Contains(window.Texts(), "Zorro") {
		test.Errorf("«Zorro» is not found by «zorro»: %q", window.Texts())
	}
}

func TestCambiarGoesBackToTheChoiceOfStands(test *testing.T) {
	_, window, _ := openStand(test)
	steps(test, window, click(window, "Cambiar"))
	settle(window)
	if _, err := window.Element("Todo el stock Todos tus productos, para vender fuera de un evento"); err != nil {
		test.Errorf("Cambiar did not go back to the choice of stands: %v", err)
	}
}

func TestAProductOfThreeVersionsIsOneTileOnTheStand(test *testing.T) {
	_, window, item := openStand(test)
	if _, err := window.Element(item.Name + " · 3 versiones"); err != nil {
		test.Fatal(err)
	}
	if slices.Contains(window.Texts(), item.Label+" · "+money.Format(item.Price)) {
		test.Errorf("a version of «%s» still has a tile of its own", item.Name)
	}
}

func TestChoosingAVersionPutsThatVersionInTheBasket(test *testing.T) {
	program, _, item := openStand(test)
	if quantity := program.stand.BasketQuantity(item.VariantID); quantity != 1 {
		test.Errorf("%d of «%s» in the basket after choosing it", quantity, item.Label)
	}
}

func TestAFairWithNothingTakenOffersToSetUpItsStock(test *testing.T) {
	program, window := openApp(test, true)
	empty, _ := fairNamed(test, program, "Mercadillo de Invierno")
	program.Open(Stand)
	program.stand.Choose(stand.Fair{Event: empty.Event, Events: program.stores.Events, Products: program.stores.Products})
	settle(window)
	steps(test, window, click(window, "Configurar stock"))
	settle(window)
	if !program.stand.Preparing() {
		test.Errorf("«Configurar stock» does not open the choice of what to take")
	}
}
