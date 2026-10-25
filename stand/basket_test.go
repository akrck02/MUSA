package stand

import (
	"fmt"
	"testing"

	"github.com/Itros97/MUSA/money"
)

var (
	cat  = Item{VariantID: "cat", Label: "Gato", Price: 300, Left: 2}
	frog = Item{VariantID: "frog", Label: "Rana", Price: 500, Left: 1}
)

func TestABasketAddsUpWhatTheCustomerTakes(test *testing.T) {
	var basket Basket
	basket.Add(cat)
	basket.Add(cat)
	basket.Add(frog)
	if basket.Count() != 3 || basket.Total() != 1100 || basket.Quantity("cat") != 2 {
		test.Errorf("count %d, total %d, cats %d", basket.Count(), basket.Total(), basket.Quantity("cat"))
	}
}

func TestABasketTakesNoMoreThanIsLeft(test *testing.T) {
	var basket Basket
	basket.Add(frog)
	if basket.Add(frog) {
		test.Error("took a second frog when only one was left")
	}
	if basket.Add(Item{VariantID: "gone", Left: 0}) {
		test.Error("took an item sold out")
	}
}

func TestRemovingTheLastOfAnItemTakesItOutOfTheBasket(test *testing.T) {
	var basket Basket
	basket.Add(cat)
	basket.Add(frog)
	basket.Remove("cat")
	basket.Remove("missing") // not in the basket: nothing happens
	counter := 0
	lines := basket.Lines(func() string { counter++; return fmt.Sprint(counter) })
	if len(lines) != 1 || lines[0].VariantID != "frog" || lines[0].Quantity != 1 || lines[0].UnitPrice != 500 || lines[0].ID != "1" {
		test.Errorf("lines %+v", lines)
	}
	basket.Clear()
	if basket.Count() != 0 {
		test.Error("a cleared basket still holds something")
	}
}

func TestTheBasketSaysWhatItHolds(test *testing.T) {
	var basket Basket
	basket.Add(Item{VariantID: "cat", Label: "Gato", Left: 5})
	basket.Add(Item{VariantID: "cat", Label: "Gato", Left: 5})
	basket.Add(Item{VariantID: "frog", Label: "Rana", Left: 5})
	if got := basket.Words(); got != "2 × Gato, 1 × Rana" {
		test.Errorf("Words = %q", got)
	}
}

func TestTheChangeIsSaidInNotesAndCoins(test *testing.T) {
	for change, want := range map[money.Money]string{
		1100: "un billete de 10 € y una moneda de 1 €",
		400:  "dos monedas de 2 €",
		0:    "",
	} {
		if got := changeWords(change); got != want {
			test.Errorf("changeWords(%d) = %q, want %q", change, got, want)
		}
	}
}

func TestTheReceiptListsEachItemWithWhatItCost(test *testing.T) {
	var basket Basket
	moon := Item{VariantID: "moon", Label: "Luna · Mate", Price: 300, Left: 5}
	basket.Add(moon)
	basket.Add(moon)
	taken := basket.Taken()
	if len(taken) != 1 || taken[0].Words != "2 × Luna · Mate" || taken[0].Amount != 600 {
		test.Errorf("taken %+v", taken)
	}
}
