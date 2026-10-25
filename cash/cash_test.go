package cash

import (
	"slices"
	"testing"

	"github.com/Itros97/MUSA/money"
)

func TestTheChangeIsWhatWasHandedOverMinusTheTotal(test *testing.T) {
	var received Received
	received.Add(2000)
	received.Add(200)
	if change, enough := Change(1850, received); !enough || change != 350 {
		test.Errorf("for 18,50 € paid with 22 €: change %d, enough %v", change, enough)
	}
}

func TestTooLittleMoneySaysHowMuchIsMissing(test *testing.T) {
	var received Received
	received.Add(1000)
	if missing, enough := Change(1850, received); enough || missing != 850 {
		test.Errorf("for 18,50 € paid with 10 €: missing %d, enough %v", missing, enough)
	}
}

func TestAMistakenTapIsUndone(test *testing.T) {
	var received Received
	received.Add(500)
	received.Add(5000)
	received.Undo()
	if received.Total() != 500 || received.Count(5000) != 0 || received.Count(500) != 1 {
		test.Errorf("after undoing: total %d", received.Total())
	}
	received.Clear()
	received.Undo() // nothing left: must not panic
	if received.Total() != 0 {
		test.Errorf("after clearing: total %d", received.Total())
	}
}

func TestTheDenominationsGoFromLargestToSmallest(test *testing.T) {
	for index := 1; index < len(Denominations); index++ {
		if Denominations[index] >= Denominations[index-1] {
			test.Errorf("%v comes after %v", Denominations[index], money.Money(Denominations[index-1]))
		}
	}
}

func TestTheChangeIsGivenWithTheFewestPieces(test *testing.T) {
	got := Breakdown(1150)
	want := []money.Money{1000, 100, 50}
	if !slices.Equal(got, want) {
		test.Errorf("Breakdown(11,50 €) = %v, want %v", got, want)
	}
}
