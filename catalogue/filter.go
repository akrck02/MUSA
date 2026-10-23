package catalogue

import (
	"slices"
	"strings"
)

// Query is the search and the filters the artist applies to the catalogue. Its zero value shows
// every product that is not archived.
type Query struct {
	Text         string
	Kind         Kind // empty for both kinds
	TagIDs       []string
	LowStockOnly bool
	Archived     bool
}

// Filtering tells whether the query narrows the catalogue at all.
func (query Query) Filtering() bool {
	return query.Text != "" || query.Kind != "" || len(query.TagIDs) > 0 || query.LowStockOnly || query.Archived
}

// Filter keeps the products that match every part of the query, in the order they come: the
// store gives them oldest first, so a new work goes to the end, where the artist expects it.
func Filter(products []Product, query Query) []Product {
	text := Fold(strings.TrimSpace(query.Text))
	var matches []Product
	for _, product := range products {
		if matchesQuery(product, query, text) {
			matches = append(matches, product)
		}
	}
	return matches
}

func matchesQuery(product Product, query Query, foldedText string) bool {
	if (product.Status == Archived) != query.Archived {
		return false
	}
	if query.Kind != "" && product.Kind != query.Kind {
		return false
	}
	if query.LowStockOnly && !product.LowStock() {
		return false
	}
	return hasEveryTag(product, query.TagIDs) && matchesText(product, foldedText)
}

func hasEveryTag(product Product, tagIDs []string) bool {
	for _, wanted := range tagIDs {
		found := slices.ContainsFunc(product.Tags, func(tag Tag) bool { return tag.ID == wanted })
		if !found {
			return false
		}
	}
	return true
}

func matchesText(product Product, foldedText string) bool {
	if foldedText == "" {
		return true
	}
	words := []string{product.Name, product.Description}
	for _, variant := range product.Variants {
		words = append(words, variant.SKU)
	}
	for _, word := range words {
		if containsText(Fold(word), foldedText) {
			return true
		}
	}
	return false
}

func containsText(text, part string) bool { return part == "" || strings.Contains(text, part) }
