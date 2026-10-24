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
func openStand(t *testing.T) (*App, *uitest.Window, stand.Item) {
	t.Helper()
	program, window := openApp(t, true)
	program.Show("stand")
	settle(window)
	items, err := stand.WholeStock{Products: program.stores.Products}.Items()
	if err != nil || len(items) == 0 {
		t.Fatalf("no product on the stand (%v)", err)
	}
	item := items[0]
	pick(t, window, items, item)
	return program, window, item
}

// pick puts one of an item in the basket the way the artist does: tapping its product's tile and,
// when the product has several versions, choosing the item's version.
func pick(t *testing.T, window *uitest.Window, items []stand.Item, item stand.Item) {
	t.Helper()
	versions := 0
	for _, each := range items {
		if each.ProductID == item.ProductID {
			versions++
		}
	}
	if versions == 1 {
		steps(t, window, clickIn(window, item.ProductID, item.Label+" · "+money.Format(item.Price)))
		settle(window)
		return
	}
	steps(t, window, clickIn(window, item.ProductID, fmt.Sprintf("%s · %d versiones", item.Name, versions)))
	settle(window)
	steps(t, window, click(window, item.Version))
	settle(window)
}

func salesCount(t *testing.T, program *App) int {
	t.Helper()
	all, err := program.stores.Sales.All()
	if err != nil {
		t.Fatal(err)
	}
	return len(all)
}

func TestPayingByCardRecordsACardSale(t *testing.T) {
	program, window, _ := openStand(t)
	before := salesCount(t, program)
	steps(t, window, click(window, "Cobrar"), click(window, "Pagar con tarjeta"), click(window, "Cobrado con tarjeta"))
	settle(window)
	all, _ := program.stores.Sales.All()
	if len(all) != before+1 || all[0].PaidWith != sales.Card {
		t.Errorf("after a card payment: %d sales from %d, the last paid with %s", len(all), before, all[0].PaidWith)
	}
}

func TestACardThatFailsRecordsNothing(t *testing.T) {
	program, window, _ := openStand(t)
	before := salesCount(t, program)
	steps(t, window, click(window, "Cobrar"), click(window, "Pagar con tarjeta"), click(window, "No ha salido"))
	settle(window)
	if after := salesCount(t, program); after != before {
		t.Errorf("a failed card payment recorded a sale")
	}
}

func TestExactCashClosesTheSale(t *testing.T) {
	program, window, _ := openStand(t)
	before := salesCount(t, program)
	steps(t, window, click(window, "Cobrar"), click(window, "Pagar en efectivo"), click(window, "Justo"), click(window, "Confirmar"))
	settle(window)
	if after := salesCount(t, program); after != before+1 {
		t.Errorf("paying the exact amount did not record the sale")
	}
}

func TestTooLittleCashDoesNotCloseTheSale(t *testing.T) {
	program, window, item := openStand(t)
	if item.Price <= 50 {
		t.Skip("the first item costs too little to pay short")
	}
	before := salesCount(t, program)
	steps(t, window, click(window, "Cobrar"), click(window, "Pagar en efectivo"), click(window, money.Format(50)), click(window, "Confirmar"))
	settle(window)
	if after := salesCount(t, program); after != before {
		t.Errorf("a sale was recorded with too little money")
	}
	if !slices.Contains(window.Texts(), "Falta") {
		t.Errorf("the dialog does not say money is missing")
	}
}

func TestClearingTheCashTypedStartsAgain(t *testing.T) {
	_, window, _ := openStand(t)
	steps(t, window, click(window, "Cobrar"), click(window, "Pagar en efectivo"), click(window, money.Format(5000)), click(window, "Borrar"))
	settle(window)
	if slices.Contains(window.Texts(), "Cambio") && slices.Contains(window.Texts(), money.Format(5000)) {
		t.Errorf("the money given is still there after Borrar")
	}
}

func TestTakingOneOutOfTheBasketEmptiesIt(t *testing.T) {
	_, window, item := openStand(t)
	steps(t, window, clickIn(window, item.ProductID, item.Name+" · 3 versiones"))
	settle(window)
	steps(t, window, click(window, "Quitar uno"))
	settle(window)
	if _, err := window.Element("Cobrar"); err == nil {
		t.Errorf("the basket bar still shows after taking its only piece out")
	}
}

func TestSearchingTheStandKeepsOnlyWhatMatches(t *testing.T) {
	program, window := openApp(t, true)
	program.Show("stand")
	settle(window)
	steps(t, window, click(window, "Buscar"))
	settle(window)
	window.Type("zorro")
	settle(window)
	for _, text := range window.Texts() {
		if text == "Estrella" {
			t.Fatalf("«Estrella» still shows when searching «zorro»")
		}
	}
	if !slices.Contains(window.Texts(), "Zorro") {
		t.Errorf("«Zorro» is not found by «zorro»: %q", window.Texts())
	}
}

func TestCambiarGoesBackToTheChoiceOfStands(t *testing.T) {
	_, window, _ := openStand(t)
	steps(t, window, click(window, "Cambiar"))
	settle(window)
	if _, err := window.Element("Todo el stock Todos tus productos, para vender fuera de un evento"); err != nil {
		t.Errorf("Cambiar did not go back to the choice of stands: %v", err)
	}
}

func TestAProductOfThreeVersionsIsOneTileOnTheStand(t *testing.T) {
	_, window, item := openStand(t)
	if _, err := window.Element(item.Name + " · 3 versiones"); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(window.Texts(), item.Label+" · "+money.Format(item.Price)) {
		t.Errorf("a version of «%s» still has a tile of its own", item.Name)
	}
}

func TestChoosingAVersionPutsThatVersionInTheBasket(t *testing.T) {
	program, _, item := openStand(t)
	if quantity := program.stand.BasketQuantity(item.VariantID); quantity != 1 {
		t.Errorf("%d of «%s» in the basket after choosing it", quantity, item.Label)
	}
}

func TestAFairWithNothingTakenOffersToSetUpItsStock(t *testing.T) {
	program, window := openApp(t, true)
	empty, _ := fairNamed(t, program, "Mercadillo de Invierno")
	program.Open(Stand)
	program.stand.Choose(stand.Fair{Event: empty.Event, Events: program.stores.Events, Products: program.stores.Products})
	settle(window)
	steps(t, window, click(window, "Configurar stock"))
	settle(window)
	if !program.stand.Preparing() {
		t.Errorf("«Configurar stock» does not open the choice of what to take")
	}
}
