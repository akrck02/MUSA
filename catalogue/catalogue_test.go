package catalogue

import (
	"slices"
	"testing"
)

func codes(problems []Problem) []Code {
	var all []Code
	for _, problem := range problems {
		all = append(all, problem.Code)
	}
	return all
}

func TestADraftIsValidatedBeforeSaving(t *testing.T) {
	valid := physicalProduct("p1", "Lámina")
	if problems := Validate(valid); len(problems) != 0 {
		t.Errorf("a single unnamed variant is rejected: %v", problems)
	}
	cases := map[string]struct {
		change func(*Product)
		want   Code
	}{
		"blank name":       {func(product *Product) { product.Name = "  " }, CodeBlankName},
		"no variants":      {func(product *Product) { product.Variants = nil }, CodeNoVariants},
		"negative stock":   {func(product *Product) { product.Variants[0].Stock = -1 }, CodeNegativeStock},
		"negative measure": {func(product *Product) { product.Variants[0].WeightGrams = -1 }, CodeInvalidMeasure},
		"unnamed among several": {func(product *Product) {
			product.Variants = []Variant{{ID: "a", Name: "A4"}, {ID: "b", Name: " "}}
		}, CodeBlankVariantName},
		"repeated name ignoring case": {func(product *Product) {
			product.Variants = []Variant{{ID: "a", Name: "A4"}, {ID: "b", Name: "a4"}}
		}, CodeDuplicateVariantName},
		"repeated sku inside the product": {func(product *Product) {
			product.Variants = []Variant{{ID: "a", Name: "A4", SKU: "LAM"}, {ID: "b", Name: "A3", SKU: "lam"}}
		}, CodeDuplicateSKU},
	}
	for name, each := range cases {
		product := physicalProduct("p1", "Lámina")
		each.change(&product)
		if got := codes(Validate(product)); !slices.Equal(got, []Code{each.want}) {
			t.Errorf("%s: %v; want %s", name, got, each.want)
		}
	}
	commission := digitalProduct("c1", "Comisión")
	commission.TurnaroundDays = -2
	if !slices.Contains(codes(Validate(commission)), CodeNegativeTurnaround) {
		t.Error("a negative turnaround is accepted")
	}
}

func names(products []Product) []string {
	var all []string
	for _, product := range products {
		all = append(all, product.Name)
	}
	return all
}

func TestTheListIsFilteredBySearchKindTagsAndStock(t *testing.T) {
	pokemon, christmas := tag("Pokémon"), tag("Navidad")
	print := physicalProduct("print", "Lámina Pikachu")
	print.Tags = []Tag{pokemon, christmas}
	sticker := physicalProduct("sticker", "Pegatina")
	sticker.Tags = []Tag{pokemon}
	sticker.Variants = []Variant{{ID: "s1", SKU: "PEG-01", Stock: 1, MinimumStock: 2}}
	archived := physicalProduct("old", "Antigua")
	archived.Status = Archived
	all := []Product{sticker, digitalProduct("commission", "Comisión busto"), print, archived}

	cases := map[string]struct {
		query Query
		want  []string
	}{
		"in the given order, archived hidden": {Query{}, []string{"Pegatina", "Comisión busto", "Lámina Pikachu"}},
		"only archived":                       {Query{Archived: true}, []string{"Antigua"}},
		"without case or accents":             {Query{Text: "LAMINA"}, []string{"Lámina Pikachu"}},
		"by sku":                              {Query{Text: "peg-01"}, []string{"Pegatina"}},
		"by kind":                             {Query{Kind: Digital}, []string{"Comisión busto"}},
		"every tag chosen":                    {Query{TagIDs: []string{pokemon.ID, christmas.ID}}, []string{"Lámina Pikachu"}},
		"low stock":                           {Query{LowStockOnly: true}, []string{"Pegatina"}},
	}
	for name, each := range cases {
		if got := names(Filter(all, each.query)); !slices.Equal(got, each.want) {
			t.Errorf("%s: %v; want %v", name, got, each.want)
		}
	}
}

func TestTagsAreSuggestedWithoutTheChosenOnes(t *testing.T) {
	all := []Tag{tag("Pokémon"), tag("Navidad"), tag("Gatos")}
	suggested := func(chosen []string, typed string) []string {
		var result []string
		for _, each := range SuggestTags(all, chosen, typed) {
			result = append(result, each.Name)
		}
		return result
	}
	if got := suggested([]string{"pokemon"}, ""); !slices.Equal(got, []string{"Navidad", "Gatos"}) {
		t.Errorf("with nothing typed: %v", got)
	}
	if got := suggested(nil, "poke"); !slices.Equal(got, []string{"Pokémon"}) {
		t.Errorf("typing poke: %v", got)
	}
}

func TestSpellingsOfOneTagAreTheSameTag(t *testing.T) {
	if got := Fold("Lámina NIÑO"); got != "lamina nino" {
		t.Errorf("Fold = %q", got)
	}
	if got := TagDisplayName("  Fan   Art "); got != "Fan Art" {
		t.Errorf("TagDisplayName = %q", got)
	}
	if got := DistinctTagNames([]string{"Pokémon", "pokemon ", "POKEMON", " ", ""}); !slices.Equal(got, []string{"Pokémon"}) {
		t.Errorf("DistinctTagNames = %v", got)
	}
	if TagColorFor("navidad") != TagColorFor(TagKey("Navidad ")) {
		t.Error("one tag gets two colours")
	}
}

func TestTagColorsSpreadOverThePalette(t *testing.T) {
	used := map[TagColor]bool{}
	for _, name := range []string{"gatos", "espacio", "paisaje", "pegatinas", "hongos", "chapas", "merch", "fanzines", "comisiones"} {
		used[TagColorFor(name)] = true
	}
	if len(used) < 4 {
		t.Errorf("nine tags share %d colours", len(used))
	}
}

func TestCountsAndCentimetresAreReadAsTyped(t *testing.T) {
	if count, found := ParseCount(" 12 "); !found || count != 12 {
		t.Errorf("ParseCount(12) = %d %v", count, found)
	}
	for _, text := range []string{"-1", "1,5"} {
		if _, found := ParseCount(text); found {
			t.Errorf("ParseCount(%q) accepted", text)
		}
	}
	if millimetres, found := ParseCentimetres("29,7"); !found || millimetres != 297 {
		t.Errorf("ParseCentimetres = %d", millimetres)
	}
	if _, found := ParseCentimetres("29,75"); found {
		t.Error("two decimals accepted")
	}
	if FormatCentimetres(297) != "29,7" || FormatCentimetres(300) != "30" {
		t.Error("FormatCentimetres")
	}
}
