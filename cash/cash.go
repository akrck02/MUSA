// Package cash is paying with notes and coins: what the customer handed over, and the change the
// artist gives back.
package cash

import "github.com/Itros97/MUSA/money"

// Denominations are the euro notes and coins the stand offers to tap, largest first. The 500 €
// and 200 € notes and the 1 and 2 cent coins are left out: nobody pays for a print with them.
var Denominations = []money.Money{5000, 2000, 1000, 500, 200, 100, 50, 20, 10, 5}

// Received is the notes and coins a customer hands over, tapped one by one.
type Received struct {
	pieces []money.Money
}

// Add counts one more note or coin.
func (received *Received) Add(piece money.Money) { received.pieces = append(received.pieces, piece) }

// Undo takes back the last note or coin tapped, for a mistaken tap; nothing happens when none was.
func (received *Received) Undo() {
	if len(received.pieces) > 0 {
		received.pieces = received.pieces[:len(received.pieces)-1]
	}
}

// Clear forgets everything tapped.
func (received *Received) Clear() { received.pieces = nil }

// Count is how many of one note or coin were tapped.
func (received Received) Count(piece money.Money) int {
	count := 0
	for _, each := range received.pieces {
		if each == piece {
			count++
		}
	}
	return count
}

// Total is what the customer handed over.
func (received Received) Total() money.Money {
	return money.Sum(received.pieces, func(piece money.Money) money.Money { return piece })
}

// Change is what to give back for a total; enough is false while the customer has handed over
// less than the total, and change is then what is still missing, as a positive amount.
func Change(total money.Money, received Received) (change money.Money, enough bool) {
	given := received.Total()
	if given < total {
		return total - given, false
	}
	return given - total, true
}
