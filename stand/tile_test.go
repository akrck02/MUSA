package stand

import "testing"

func TestLeftWordsWarnWhenFewAreLeft(test *testing.T) {
	cases := []struct {
		left  int
		words string
	}{{0, "Agotado"}, {2, "Quedan 2"}, {9, "9"}}
	for _, each := range cases {
		if words := leftWords(each.left); words != each.words {
			test.Errorf("%d left: %q, want %q", each.left, words, each.words)
		}
	}
}
