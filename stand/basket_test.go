package stand

import (
	"fmt"
	"testing"
)

var (
	cat  = Item{VariantID: "cat", Label: "Gato", Price: 300, Left: 2}
	frog = Item{VariantID: "frog", Label: "Rana", Price: 500, Left: 1}
)

func TestABasketAddsUpWhatTheCustomerTakes(t *testing.T) {
	var basket Basket
	basket.Add(cat)
	basket.Add(cat)
	basket.Add(frog)
	if basket.Count() != 3 || basket.Total() != 1100 || basket.Quantity("cat") != 2 {
		t.Errorf("count %d, total %d, cats %d", basket.Count(), basket.Total(), basket.Quantity("cat"))
	}
}

func TestABasketTakesNoMoreThanIsLeft(t *testing.T) {
	var basket Basket
	basket.Add(frog)
	if basket.Add(frog) {
		t.Error("took a second frog when only one was left")
	}
	if basket.Add(Item{VariantID: "gone", Left: 0}) {
		t.Error("took an item sold out")
	}
}

func TestRemovingTheLastOfAnItemTakesItOutOfTheBasket(t *testing.T) {
	var basket Basket
	basket.Add(cat)
	basket.Add(frog)
	basket.Remove("cat")
	basket.Remove("missing") // not in the basket: nothing happens
	counter := 0
	lines := basket.Lines(func() string { counter++; return fmt.Sprint(counter) })
	if len(lines) != 1 || lines[0].VariantID != "frog" || lines[0].Quantity != 1 || lines[0].UnitPrice != 500 || lines[0].ID != "1" {
		t.Errorf("lines %+v", lines)
	}
	basket.Clear()
	if basket.Count() != 0 {
		t.Error("a cleared basket still holds something")
	}
}
