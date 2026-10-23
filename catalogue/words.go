package catalogue

import (
	"fmt"

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

// PriceRangeWords is "8 €", or "8 € – 15 €" from the cheapest variant to the dearest.
func PriceRangeWords(product Product) string {
	lowest := product.LowestPrice()
	highest := lowest
	for _, variant := range product.Variants {
		highest = max(highest, variant.Price)
	}
	if highest == lowest {
		return money.Format(lowest)
	}
	return money.Format(lowest) + " – " + money.Format(highest)
}

// SummaryWords is the kind, with the number of variants when there is more than one.
func SummaryWords(product Product) string {
	if len(product.Variants) > 1 {
		return fmt.Sprintf("%s · %d variantes", kindWords[product.Kind], len(product.Variants))
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
	CodeNoVariants:           "Añade al menos una variante",
	CodeNegativeTurnaround:   "El plazo no puede ser negativo",
	CodeBlankVariantName:     "Con varias variantes, cada una necesita nombre",
	CodeDuplicateVariantName: "Ya hay otra variante con este nombre",
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
