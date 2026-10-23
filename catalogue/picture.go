package catalogue

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Initial is the first letter of a product's name, drawn on its picture until it has a photo: the
// quoted title when there is one, so «Lámina «Gato»» gives G, not the L every print shares.
func Initial(name string) string {
	if _, title, quoted := strings.Cut(name, "«"); quoted {
		name = title
	}
	for _, letter := range name {
		if unicode.IsLetter(letter) || unicode.IsDigit(letter) {
			return string(unicode.ToUpper(letter))
		}
	}
	first, _ := utf8.DecodeRuneInString(name)
	return string(first)
}

// PictureColor is the colour of a product's picture until it has a photo: one of the tags'
// palette, picked by its name, so works side by side look different and each keeps its colour.
func PictureColor(product Product) TagColor { return TagColorFor(TagKey(product.Name)) }

// Restock lists what needs restocking, as the artist says it: "A3 de Gato astronauta" for a
// version, "Luna" for a product with a single one. It reads physical products that are not
// archived.
func Restock(products []Product) []string {
	var names []string
	for _, product := range products {
		if product.Kind != Physical || product.Status == Archived {
			continue
		}
		for _, variant := range product.Variants {
			if !variant.LowStock() {
				continue
			}
			if len(product.Variants) == 1 || variant.Name == "" {
				names = append(names, product.Name)
				continue
			}
			names = append(names, variant.Name+" de "+product.Name)
		}
	}
	return names
}
