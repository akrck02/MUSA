package style

import (
	"slices"
	"testing"
)

func TestARunOfCapitalsIsOneWord(test *testing.T) {
	cases := map[string][]string{
		"basketQty": {"basket", "qty"},
		"HTTPReq":   {"http", "req"},
		"VariantID": {"variant", "id"},
		"max_val":   {"max", "val"},
	}
	for name, want := range cases {
		if words := camelCaseWords(name); !slices.Equal(words, want) {
			test.Errorf("%s: %q, want %q", name, words, want)
		}
	}
}
