package projects

import "fmt"

// stageWords are the board's columns as the artist reads them.
var stageWords = map[Stage]string{ToStart: "Por empezar", Drawing: "Dibujando", ToReview: "Para revisar", Delivered: "Entregado"}

// StageWords is a stage's name, in Spanish.
func StageWords(stage Stage) string { return stageWords[stage] }

var problemWords = map[Code]string{
	CodeBlankTitle:        "Ponle un título",
	CodeNoCustomer:        "¿Para quién es?",
	CodeNegativePrice:     "El precio no puede ser negativo",
	CodeProductNotDigital: "Un proyecto solo lleva productos digitales",
	CodeAlreadyDelivered:  "Ya está entregado: su venta está en el historial",
}

// ProblemWords is what the artist reads about a problem, in Spanish; the code when it has no words.
func ProblemWords(code Code) string {
	if words, found := problemWords[code]; found {
		return words
	}
	return string(code)
}

// projectCount is "1 proyecto" or "5 proyectos".
func projectCount(count int) string {
	if count == 1 {
		return "1 proyecto"
	}
	return fmt.Sprintf("%d proyectos", count)
}
