package catalogue

import "testing"

func TestATagsTabNamesItInThePlural(test *testing.T) {
	for name, want := range map[string]string{
		"Lámina": "Láminas", "Llavero": "Llaveros", "Pegatina": "Pegatinas", "Imán": "Imanes",
		"Lápiz": "Lápices", "Póster": "Pósteres", "Café": "Cafés", "Chapas": "Chapas", "Taza grande": "Tazas grande",
	} {
		if got := Plural(name); got != want {
			test.Errorf("the plural of %q is %q, not %q", name, got, want)
		}
	}
}

func TestATagOfACommonKindOfWorkHasItsOwnIcon(test *testing.T) {
	for name, want := range map[string]string{"Pegatina": "sticky-note", "Láminas": "printer", "llavero": "key", "Acuarela": "tag"} {
		if got := TagIcon(name); got != want {
			test.Errorf("the icon of %q is %q, not %q", name, got, want)
		}
	}
}

func TestAProductSaysWhatItIsWithItsFirstTagAndAKindThatAgrees(test *testing.T) {
	cases := []struct {
		kind Kind
		tags []string
		want string
	}{
		{Physical, []string{"Llavero", "ranas"}, "Llavero físico"},
		{Physical, []string{"Pegatina"}, "Pegatina física"},
		{Digital, []string{"Fondo"}, "Fondo digital"},
		{Physical, nil, "Físico"},
	}
	for _, each := range cases {
		if got := KindWords(each.kind, each.tags); got != each.want {
			test.Errorf("KindWords(%v, %v) = %q, want %q", each.kind, each.tags, got, each.want)
		}
	}
}
