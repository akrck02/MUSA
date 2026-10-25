package app

import (
	"slices"
	"testing"
	"time"

	"github.com/Itros97/MUSA/calendar"
	"github.com/akrck02/go-ui/uitest"
)

func TestAHistoryKindKeepsOnlyItsSales(test *testing.T) {
	program, window := openApp(test, true)
	program.Show("history")
	settle(window)
	steps(test, window, click(window, "Ventas locales"))
	settle(window)
	for _, text := range window.Texts() {
		if text == "Sara" { // Sara's order was sent, not sold on the spot
			test.Fatalf("a sent order shows among the local sales")
		}
	}
	steps(test, window, click(window, "Envíos"))
	settle(window)
	if !slices.Contains(window.Texts(), "Sara") {
		test.Errorf("Sara's sent order is not among the sent ones: %q", window.Texts())
	}
}

func TestASaleOpenedFromTheHistoryCanBeDeleted(test *testing.T) {
	program, window := openApp(test, true)
	program.Show("history")
	settle(window)
	before := salesCount(test, program)
	// the demo sold to Sara yesterday, whatever day the test runs
	yesterday := calendar.DayWords(calendar.Of(time.Now()).AddDays(-1))
	steps(test, window, click(window, yesterday+" · Sara · 10 €"))
	settle(window)
	steps(test, window, click(window, "Eliminar venta"))
	settle(window)
	if after := salesCount(test, program); after != before-1 {
		test.Errorf("%d sales after deleting one of %d", after, before)
	}
}

func TestASentOrderDroppedOnDeliveredLeavesTheBoard(test *testing.T) {
	program, window := openApp(test, true)
	program.Open(Sales)
	settle(window)
	dragOnto(test, window, "Ana", "Entregado")
	settle(window)
	if sale := saleOf(test, program, "Ana"); sale.Status != "DELIVERED" {
		test.Errorf("Ana's order is %s after being dropped on «Entregado»", sale.Status)
	}
}

func TestAnOrderToPrepareCannotSkipToDelivered(test *testing.T) {
	program, window := openApp(test, true)
	program.Open(Sales)
	settle(window)
	dragOnto(test, window, "Marcos", "Entregado")
	settle(window)
	if sale := saleOf(test, program, "Marcos"); sale.Status != "TO_PREPARE" {
		test.Errorf("Marcos's order is %s after being dropped on «Entregado» before it was sent", sale.Status)
	}
}

// dragOnto drags an element by its middle and lets it go over the middle of another.
func dragOnto(test *testing.T, window *uitest.Window, from, to string) {
	test.Helper()
	source, err := window.Element(from)
	if err != nil {
		test.Fatal(err)
	}
	target, err := window.Element(to)
	if err != nil {
		test.Fatal(err)
	}
	window.Drag(source.Rect.X+source.Rect.Width/2, source.Rect.Y+source.Rect.Height/2,
		target.Rect.X+target.Rect.Width/2, target.Rect.Y+target.Rect.Height/2)
	window.Frames(2)
}
