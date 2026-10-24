package app

import (
	"slices"
	"testing"

	"github.com/akrck02/go-ui/uitest"
)

func TestAHistoryKindKeepsOnlyItsSales(t *testing.T) {
	program, window := openApp(t, true)
	program.Show("history")
	settle(window)
	steps(t, window, click(window, "Ventas locales"))
	settle(window)
	for _, text := range window.Texts() {
		if text == "Sara" { // Sara's order was sent, not sold on the spot
			t.Fatalf("a sent order shows among the local sales")
		}
	}
	steps(t, window, click(window, "Envíos"))
	settle(window)
	if !slices.Contains(window.Texts(), "Sara") {
		t.Errorf("Sara's sent order is not among the sent ones: %q", window.Texts())
	}
}

func TestASaleOpenedFromTheHistoryCanBeDeleted(t *testing.T) {
	program, window := openApp(t, true)
	program.Show("history")
	settle(window)
	before := salesCount(t, program)
	steps(t, window, click(window, "23 oct · Sara · 10 €"))
	settle(window)
	steps(t, window, click(window, "Eliminar venta"))
	settle(window)
	if after := salesCount(t, program); after != before-1 {
		t.Errorf("%d sales after deleting one of %d", after, before)
	}
}

func TestASentOrderDroppedOnDeliveredLeavesTheBoard(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Sales)
	settle(window)
	dragOnto(t, window, "Ana", "Entregado")
	settle(window)
	if sale := saleOf(t, program, "Ana"); sale.Status != "DELIVERED" {
		t.Errorf("Ana's order is %s after being dropped on «Entregado»", sale.Status)
	}
}

func TestAnOrderToPrepareCannotSkipToDelivered(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Sales)
	settle(window)
	dragOnto(t, window, "Marcos", "Entregado")
	settle(window)
	if sale := saleOf(t, program, "Marcos"); sale.Status != "TO_PREPARE" {
		t.Errorf("Marcos's order is %s after being dropped on «Entregado» before it was sent", sale.Status)
	}
}

// dragOnto drags an element by its middle and lets it go over the middle of another.
func dragOnto(t *testing.T, window *uitest.Window, from, to string) {
	t.Helper()
	source, err := window.Element(from)
	if err != nil {
		t.Fatal(err)
	}
	target, err := window.Element(to)
	if err != nil {
		t.Fatal(err)
	}
	window.Drag(source.Rect.X+source.Rect.Width/2, source.Rect.Y+source.Rect.Height/2,
		target.Rect.X+target.Rect.Width/2, target.Rect.Y+target.Rect.Height/2)
	window.Frames(2)
}
