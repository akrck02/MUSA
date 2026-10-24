package app

import (
	"slices"
	"testing"
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

func TestTheOrderButtonMovesAnOrderOn(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Sales)
	settle(window)
	steps(t, window, clickIn(window, orderID(t, program, "Marcos"), "Preparado"))
	settle(window)
	if sale := saleOf(t, program, "Marcos"); sale.Status != "SENT" {
		t.Errorf("Marcos's order is %s after «Preparado»", sale.Status)
	}
}
