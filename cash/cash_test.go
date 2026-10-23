package cash

import (
	"testing"

	"github.com/Itros97/MUSA/money"
)

func TestTheChangeIsWhatWasHandedOverMinusTheTotal(t *testing.T) {
	var received Received
	received.Add(2000)
	received.Add(200)
	if change, enough := Change(1850, received); !enough || change != 350 {
		t.Errorf("for 18,50 € paid with 22 €: change %d, enough %v", change, enough)
	}
}

func TestTooLittleMoneySaysHowMuchIsMissing(t *testing.T) {
	var received Received
	received.Add(1000)
	if missing, enough := Change(1850, received); enough || missing != 850 {
		t.Errorf("for 18,50 € paid with 10 €: missing %d, enough %v", missing, enough)
	}
}

func TestAMistakenTapIsUndone(t *testing.T) {
	var received Received
	received.Add(500)
	received.Add(5000)
	received.Undo()
	if received.Total() != 500 || received.Count(5000) != 0 || received.Count(500) != 1 {
		t.Errorf("after undoing: total %d", received.Total())
	}
	received.Clear()
	received.Undo() // nothing left: must not panic
	if received.Total() != 0 {
		t.Errorf("after clearing: total %d", received.Total())
	}
}

func TestTheDenominationsGoFromLargestToSmallest(t *testing.T) {
	for index := 1; index < len(Denominations); index++ {
		if Denominations[index] >= Denominations[index-1] {
			t.Errorf("%v comes after %v", Denominations[index], money.Money(Denominations[index-1]))
		}
	}
}
