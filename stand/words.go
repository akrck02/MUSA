package stand

import (
	"fmt"
	"strings"

	"github.com/Itros97/MUSA/cash"
	"github.com/Itros97/MUSA/money"
)

// numberWords are the Spanish words for the few notes or coins of one kind change ever takes.
var numberWords = []string{"", "un", "dos", "tres", "cuatro"}

// changeWords says which notes and coins make the change: "un billete de 10 € y una moneda de 1 €".
func changeWords(change money.Money) string {
	pieces := cash.Breakdown(change)
	var parts []string
	for index := 0; index < len(pieces); {
		count := 1
		for index+count < len(pieces) && pieces[index+count] == pieces[index] {
			count++
		}
		parts = append(parts, pieceWords(pieces[index], count))
		index += count
	}
	if len(parts) < 2 {
		return strings.Join(parts, "")
	}
	return strings.Join(parts[:len(parts)-1], ", ") + " y " + parts[len(parts)-1]
}

func pieceWords(piece money.Money, count int) string {
	kind, plural, article := "billete", "billetes", "un"
	if !cash.IsNote(piece) {
		kind, plural, article = "moneda", "monedas", "una"
	}
	if count == 1 {
		return fmt.Sprintf("%s %s de %s", article, kind, money.Format(piece))
	}
	number := fmt.Sprint(count)
	if count < len(numberWords) {
		number = numberWords[count]
	}
	return fmt.Sprintf("%s %s de %s", number, plural, money.Format(piece))
}
