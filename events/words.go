package events

import "fmt"

// What the screens say, in Spanish, the language of the app.
var (
	statusWords  = map[Status]string{Open: "Abierto", Closed: "Cerrado"}
	refusalWords = map[Code]string{
		CodeBlankName:       "Ponle un nombre",
		CodeEndBeforeStart:  "No puede terminar antes de empezar",
		CodeInvalidQuantity: "La cantidad tiene que ser al menos 1",
		CodeNegativePrice:   "El precio no puede ser negativo",
		CodeBlankConcept:    "Escribe en qué fue el gasto",
		CodeNegativeAmount:  "El importe no puede ser negativo",
		CodeNothingToSell:   "Elige al menos un producto",
		CodeEventClosed:     "El evento está cerrado",
		CodeHasSales:        "Tiene ventas: ciérralo en vez de eliminarlo",
	}
)

// RefusalWords is what the artist reads when an action on a fair is turned down.
func RefusalWords(refusal Refusal) string {
	switch refusal.Code {
	case CodeNotEnoughStock:
		return fmt.Sprintf("En el catálogo solo quedan %d", refusal.Available)
	case CodeExceedsRemaining:
		return fmt.Sprintf("En el evento solo quedan %d", refusal.Available)
	}
	if words, found := refusalWords[refusal.Code]; found {
		return words
	}
	return string(refusal.Code)
}
