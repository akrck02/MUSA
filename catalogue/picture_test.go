package catalogue

import (
	"slices"
	"testing"
)

func TestAPicturesLetterSkipsQuotes(test *testing.T) {
	for name, want := range map[string]string{"Lámina «Gato»": "G", "Luna": "L", "«luna»": "L", "3D": "3", "": "�"} {
		if got := Initial(name); got != want {
			test.Errorf("Initial(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestRestockNamesTheVersionThatRunsLow(test *testing.T) {
	products := []Product{
		{Name: "Gato", Kind: Physical, Variants: []Variant{{Name: "A4", Stock: 9}, {Name: "A3", Stock: 0}}},
		{Name: "Luna", Kind: Physical, Variants: []Variant{{Stock: 1, MinimumStock: 2}}},
		{Name: "Retrato", Kind: Digital, Variants: []Variant{{Stock: 0}}},
		{Name: "Viejo", Kind: Physical, Status: Archived, Variants: []Variant{{Stock: 0}}},
	}
	want := []string{"A3 de Gato", "Luna"}
	if got := Restock(products); !slices.Equal(got, want) {
		test.Errorf("Restock = %v, want %v", got, want)
	}
}

func TestAPriceRangeGoesFromCheapestToDearest(test *testing.T) {
	product := Product{Variants: []Variant{{Price: 1500}, {Price: 800}}}
	if got := PriceRangeWords(product); got != "8 € – 15 €" {
		test.Errorf("PriceRangeWords = %q", got)
	}
}
