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
