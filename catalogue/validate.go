package catalogue

import "strings"

// Validate lists every problem of a draft; none means it can be saved. Checks that need the
// database, such as a SKU used by another product, are made when saving.
func Validate(product Product) []Problem {
	var problems []Problem
	if strings.TrimSpace(product.Name) == "" {
		problems = append(problems, Problem{Code: CodeBlankName})
	}
	if len(product.Variants) == 0 {
		problems = append(problems, Problem{Code: CodeNoVariants})
	}
	problems = append(problems, variantNameProblems(product.Variants)...)
	for _, variant := range product.Variants {
		problems = append(problems, variantProblems(product.Kind, variant)...)
	}
	if product.Kind == Physical {
		problems = append(problems, repeatedSKUProblems(product.Variants)...)
	}
	if product.Kind == Digital && product.TurnaroundDays < 0 {
		problems = append(problems, Problem{Code: CodeNegativeTurnaround})
	}
	return problems
}

// variantNameProblems asks for names only when there are several variants: a single one is shown
// as "the product", so it may stay unnamed.
func variantNameProblems(variants []Variant) []Problem {
	if len(variants) < 2 {
		return nil
	}
	seen := map[string]bool{}
	var problems []Problem
	for _, variant := range variants {
		key := Fold(strings.TrimSpace(variant.Name))
		switch {
		case key == "":
			problems = append(problems, Problem{Code: CodeBlankVariantName, VariantID: variant.ID})
		case seen[key]:
			problems = append(problems, Problem{Code: CodeDuplicateVariantName, VariantID: variant.ID})
		}
		seen[key] = true
	}
	return problems
}

func variantProblems(kind Kind, variant Variant) []Problem {
	var codes []Code
	if variant.Price < 0 {
		codes = append(codes, CodeNegativePrice)
	}
	if variant.Cost < 0 {
		codes = append(codes, CodeNegativeCost)
	}
	if kind == Physical {
		codes = append(codes, stockProblems(variant)...)
	}
	problems := make([]Problem, 0, len(codes))
	for _, code := range codes {
		problems = append(problems, Problem{Code: code, VariantID: variant.ID})
	}
	return problems
}

func stockProblems(variant Variant) []Code {
	var codes []Code
	if variant.Stock < 0 {
		codes = append(codes, CodeNegativeStock)
	}
	if variant.MinimumStock < 0 {
		codes = append(codes, CodeNegativeMinimumStock)
	}
	measures := []int{variant.WeightGrams, variant.WidthMillimetres, variant.HeightMillimetres, variant.DepthMillimetres}
	for _, measure := range measures {
		if measure < 0 {
			return append(codes, CodeInvalidMeasure)
		}
	}
	return codes
}

func repeatedSKUProblems(variants []Variant) []Problem {
	seen := map[string]bool{}
	var problems []Problem
	for _, variant := range variants {
		key := strings.ToLower(strings.TrimSpace(variant.SKU))
		if key == "" {
			continue
		}
		if seen[key] {
			problems = append(problems, Problem{Code: CodeDuplicateSKU, VariantID: variant.ID})
		}
		seen[key] = true
	}
	return problems
}
