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

func TestADraftIsValidatedBeforeSaving(test *testing.T) {
	valid := physicalProduct("p1", "Lámina")
	if problems := Validate(valid); len(problems) != 0 {
		test.Errorf("a single unnamed variant is rejected: %v", problems)
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
			test.Errorf("%s: %v; want %s", name, got, each.want)
		}
	}
	commission := digitalProduct("c1", "Comisión")
	commission.TurnaroundDays = -2
	if !slices.Contains(codes(Validate(commission)), CodeNegativeTurnaround) {
		test.Error("a negative turnaround is accepted")
	}
}

func names(products []Product) []string {
	var all []string
	for _, product := range products {
		all = append(all, product.Name)
	}
	return all
}

func TestTheListIsFilteredBySearchKindTagsAndStock(test *testing.T) {
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
			test.Errorf("%s: %v; want %v", name, got, each.want)
		}
	}
}

func TestTagsAreSuggestedWithoutTheChosenOnes(test *testing.T) {
	all := []Tag{tag("Pokémon"), tag("Navidad"), tag("Gatos")}
	suggested := func(chosen []string, typed string) []string {
		var result []string
		for _, each := range SuggestTags(all, chosen, typed) {
			result = append(result, each.Name)
		}
		return result
	}
	if got := suggested([]string{"pokemon"}, ""); !slices.Equal(got, []string{"Navidad", "Gatos"}) {
		test.Errorf("with nothing typed: %v", got)
	}
	if got := suggested(nil, "poke"); !slices.Equal(got, []string{"Pokémon"}) {
		test.Errorf("typing poke: %v", got)
	}
}

func TestSpellingsOfOneTagAreTheSameTag(test *testing.T) {
	if got := Fold("Lámina NIÑO"); got != "lamina nino" {
		test.Errorf("Fold = %q", got)
	}
	if got := TagDisplayName("  Fan   Art "); got != "Fan Art" {
		test.Errorf("TagDisplayName = %q", got)
	}
	if got := DistinctTagNames([]string{"Pokémon", "pokemon ", "POKEMON", " ", ""}); !slices.Equal(got, []string{"Pokémon"}) {
		test.Errorf("DistinctTagNames = %v", got)
	}
	if TagColorFor("navidad") != TagColorFor(TagKey("Navidad ")) {
		test.Error("one tag gets two colours")
	}
}

func TestTagColorsSpreadOverThePalette(test *testing.T) {
	used := map[TagColor]bool{}
	for _, name := range []string{"gatos", "espacio", "paisaje", "pegatinas", "hongos", "chapas", "merch", "fanzines", "comisiones"} {
		used[TagColorFor(name)] = true
	}
	if len(used) < 4 {
		test.Errorf("nine tags share %d colours", len(used))
	}
}

func TestCountsAndCentimetresAreReadAsTyped(test *testing.T) {
	if count, found := ParseCount(" 12 "); !found || count != 12 {
		test.Errorf("ParseCount(12) = %d %v", count, found)
	}
	for _, text := range []string{"-1", "1,5"} {
		if _, found := ParseCount(text); found {
			test.Errorf("ParseCount(%q) accepted", text)
		}
	}
	if millimetres, found := ParseCentimetres("29,7"); !found || millimetres != 297 {
		test.Errorf("ParseCentimetres = %d", millimetres)
	}
	if _, found := ParseCentimetres("29,75"); found {
		test.Error("two decimals accepted")
	}
	if FormatCentimetres(297) != "29,7" || FormatCentimetres(300) != "30" {
		test.Error("FormatCentimetres")
	}
}

func TestADigitalProductSaysWhenItArrivesAndHowOftenItSold(test *testing.T) {
	for days, want := range map[int]string{0: "Al momento", 1: "1 día", 7: "7 días"} {
		if got := DeliveryWords(days); got != want {
			test.Errorf("%d days: %q, not %q", days, got, want)
		}
	}
	for units, want := range map[int]string{0: "Sin ventas este mes", 1: "Vendido 1 vez este mes", 23: "Vendido 23 veces este mes"} {
		if got := SoldWords(units); got != want {
			test.Errorf("%d units: %q, not %q", units, got, want)
		}
	}
}

func TestTheEditorSaysWhatEachPieceEarnsOrLoses(test *testing.T) {
	cases := []struct {
		price, cost string
		words       string
		losing      bool
	}{
		{"12,50", "3,50", "Ganas 9 € por pieza · 72 %", false},
		{"10", "12", "Pierdes 2 € con cada pieza: te cuesta más de lo que cobras", true},
		{"", "3", "", false},
		{"abc", "3", "", false},
	}
	for _, each := range cases {
		words, losing := marginWords(each.price, each.cost)
		if words != each.words || losing != each.losing {
			test.Errorf("%s and %s: %q (losing %v), want %q", each.price, each.cost, words, losing, each.words)
		}
	}
}

func TestAProductWithNoNameHasNoInitial(test *testing.T) {
	if initial := Initial(""); initial != "" {
		test.Errorf("Initial of an empty name: %q, want nothing", initial)
	}
}
