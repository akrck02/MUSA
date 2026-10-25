package catalogue

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Itros97/MUSA/money"
)

// What the screens call each kind and status, in Spanish, the language of the app.
var (
	kindWords       = map[Kind]string{Physical: "Físico", Digital: "Digital"}
	pluralKindWords = map[Kind]string{Physical: "Físicos", Digital: "Digitales"}
	statusWords     = map[Status]string{Draft: "Borrador", Active: "Activo", Archived: "Archivado"}
	statuses        = []Status{Draft, Active, Archived}
)

// PriceWords is "8 €", or "desde 8 €" when the variants have different prices.
func PriceWords(product Product) string {
	price := money.Format(product.LowestPrice())
	if product.HasPriceRange() {
		return "desde " + price
	}
	return price
}

// PriceRangeWords is "8 €", or "8 – 15 €" from the cheapest variant to the dearest.
func PriceRangeWords(product Product) string {
	lowest := product.LowestPrice()
	highest := lowest
	for _, variant := range product.Variants {
		highest = max(highest, variant.Price)
	}
	if highest == lowest {
		return money.Format(lowest)
	}
	return money.FormatForInput(lowest) + " – " + money.Format(highest)
}

// SummaryWords is the kind, with the number of variants when there is more than one.
func SummaryWords(product Product) string {
	if len(product.Variants) > 1 {
		return fmt.Sprintf("%s · %d versiones", kindWords[product.Kind], len(product.Variants))
	}
	return kindWords[product.Kind]
}

// StockWords is a physical product's stock in a few words.
func StockWords(product Product) string {
	total := product.TotalStock()
	switch {
	case total == 0:
		return "Agotado"
	case product.LowStock():
		return fmt.Sprintf("Stock bajo · %d", total)
	default:
		return fmt.Sprintf("%d en stock", total)
	}
}

// problemWords is what the artist reads beside a field a problem is about.
var problemWords = map[Code]string{
	CodeBlankName:            "Ponle un nombre",
	CodeNoVariants:           "Añade al menos una versión",
	CodeNegativeTurnaround:   "El plazo no puede ser negativo",
	CodeBlankVariantName:     "Con varias versiones, cada una necesita nombre",
	CodeDuplicateVariantName: "Ya hay otra versión con este nombre",
	CodeNegativePrice:        "El precio no puede ser negativo",
	CodeNegativeCost:         "El coste no puede ser negativo",
	CodeNegativeStock:        "El stock no puede ser negativo",
	CodeNegativeMinimumStock: "El mínimo no puede ser negativo",
	CodeInvalidMeasure:       "Las medidas tienen que ser mayores que cero",
	CodeDuplicateSKU:         "Este SKU se repite en el producto",
}

// ProblemWords is what the artist reads about a problem of a draft.
func ProblemWords(problem Problem) string {
	if problem.Code == CodeDuplicateSKU && problem.OwnerName != "" {
		return "Este SKU ya lo usa «" + problem.OwnerName + "»"
	}
	if words, found := problemWords[problem.Code]; found {
		return words
	}
	return string(problem.Code)
}

// DeliveryWords is how long a digital product takes to reach the buyer: «Al momento» when it is
// ready, or «7 días».
func DeliveryWords(days int) string {
	switch days {
	case 0:
		return "Al momento"
	case 1:
		return "1 día"
	}
	return fmt.Sprintf("%d días", days)
}

// SoldWords says how many times a product sold this month: «Vendido 8 veces este mes».
func SoldWords(units int) string {
	switch units {
	case 0:
		return "Sin ventas este mes"
	case 1:
		return "Vendido 1 vez este mes"
	}
	return fmt.Sprintf("Vendido %d veces este mes", units)
}

// KindWords is what a product is, as its editor says it: its first tag and the kind, which agrees
// with it ("Llavero físico", "Pegatina física"), or the kind alone ("Físico", "Digital").
func KindWords(kind Kind, tags []string) string {
	if len(tags) == 0 {
		return kindWords[kind]
	}
	adjective := map[Kind]string{Physical: "físico", Digital: "digital"}[kind]
	if kind == Physical && strings.HasSuffix(tags[0], "a") {
		adjective = "física" // the feminine, as «Pegatina física»
	}
	return tags[0] + " " + adjective
}

// archiveWords is the button that archives a product, or restores an archived one.
func archiveWords(archived bool) string {
	if archived {
		return "Restaurar"
	}
	return "Archivar"
}

// archivedWords is the note after that button: what the product became.
func archivedWords(wasArchived bool) string {
	if wasArchived {
		return "Restaurado"
	}
	return "Archivado"
}

// restockTitle is "1 producto por reponer", or "3 productos por reponer".
func restockTitle(count int) string {
	if count == 1 {
		return "1 producto por reponer"
	}
	return fmt.Sprintf("%d productos por reponer", count)
}

// restockLinkWords is the link that shows only what needs restocking, or everything again.
func restockLinkWords(onlyLowStock bool) string {
	if onlyLowStock {
		return "Ver todos"
	}
	return "Ver cuáles"
}

// priceHeading is the heading of a product with one version: physical ones have stock too.
func priceHeading(physical bool) string {
	if physical {
		return "Precio y stock"
	}
	return "Precio"
}

// versionsHint is what each version of a product holds.
func versionsHint(physical bool) string {
	if physical {
		return "Cada versión tiene su precio y su stock."
	}
	return "Cada versión tiene su precio."
}

// lowStockWords is the warning on a version with few left, or none.
func lowStockWords(left int) string {
	if left == 0 {
		return "Agotado"
	}
	return "Pocos"
}

// marginWords says what each piece earns or loses at the typed price and cost, and whether it
// loses; it says nothing until both read as amounts.
func marginWords(price, cost string) (words string, losing bool) {
	priceAmount, priceRead := money.Parse(price)
	costAmount, costRead := money.Parse(cost)
	if !priceRead || !costRead {
		return "", false
	}
	margin := money.MarginOf(priceAmount, costAmount)
	if margin.Negative() {
		return "Pierdes " + money.Format(-margin.Amount) + " con cada pieza: te cuesta más de lo que cobras", true
	}
	words = "Ganas " + money.Format(margin.Amount) + " por pieza"
	if margin.HasPercent {
		words += " · " + strconv.Itoa(margin.Percent) + " %"
	}
	return words, false
}
