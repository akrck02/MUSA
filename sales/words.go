package sales

import "fmt"

// What the screens call each channel and period, in Spanish, the language of the app.
var (
	channelWords = map[Channel]string{AtEvent: "Evento", OnlineShop: "Tienda online", Social: "Redes y mensajes",
		Marketplace: "Marketplace", InPerson: "Venta local", Other: "Otro", Commission: "Proyecto"}
	periodWords = map[Period]string{ThisMonth: "Este mes", LastMonth: "Mes pasado", ThisYear: "Este año", AllTime: "Todo"}
)

var statusWords = map[Status]string{ToPrepare: "Por preparar", Sent: "Enviado", Delivered: "Entregado"}

// StatusWords is a status's name, in Spanish.
func StatusWords(status Status) string { return statusWords[status] }

// ChannelWords is a channel's name, in Spanish.
func ChannelWords(channel Channel) string { return channelWords[channel] }

var problemWords = map[Code]string{
	CodeNoLines:          "Añade al menos un producto",
	CodeEventChannel:     "Las ventas de evento se apuntan desde su evento",
	CodeNegativeShipping: "El envío no puede ser negativo",
	CodeManagedByEvent:   "Esta venta es de un evento: deshazla desde el evento",
	CodeAlreadyDelivered: "Este pedido ya está entregado",
	CodeInvalidQuantity:  "La cantidad tiene que ser al menos 1",
	CodeNegativePrice:    "El precio no puede ser negativo",
}

// ProblemWords is what the artist reads about a problem of a sale.
func ProblemWords(problem Problem) string {
	if problem.Code == CodeNotEnoughStock {
		return fmt.Sprintf("En el catálogo solo quedan %d", problem.Available)
	}
	if words, found := problemWords[problem.Code]; found {
		return words
	}
	return string(problem.Code)
}

// linesShown is how many lines a sale's summary names before it only counts the pieces.
const linesShown = 2

// SummaryWords is what a sale was and how it ended, as the history writes it: "2 × Gato A5,
// 1 × Rana · en mano", "3 piezas · entregado", or a project's name alone.
func SummaryWords(sale Sale) string {
	words := linesWords(sale)
	switch {
	case sale.Channel == InPerson:
		words += " · en mano"
	case sale.Channel != AtEvent && sale.Status == Delivered:
		words += " · entregado"
	}
	return words
}

func linesWords(sale Sale) string {
	switch {
	case len(sale.Lines) == 0:
		return "Venta"
	case sale.Channel == Commission:
		return sale.Lines[0].ProductName
	case len(sale.Lines) > linesShown:
		return piecesWords(sale)
	}
	words := ""
	for index, line := range sale.Lines {
		if index > 0 {
			words += ", "
		}
		words += fmt.Sprintf("%d × %s", line.Quantity, line.ProductName)
		if line.VariantName != "" {
			words += " " + line.VariantName
		}
	}
	return words
}
