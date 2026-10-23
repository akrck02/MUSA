package money

import "testing"

func TestTypedPricesAreReadAsCents(t *testing.T) {
	cases := map[string]Money{"12": 1200, "12,5": 1250, "12.05": 1205, " 8 € ": 800}
	for text, want := range cases {
		got, found := Parse(text)
		if !found || got != want {
			t.Errorf("Parse(%q) = %d, %v; want %d", text, got, found, want)
		}
	}
}

func TestTextThatIsNotAPriceIsRejected(t *testing.T) {
	for _, text := range []string{"", "abc", "-3", "1,234"} {
		if _, found := Parse(text); found {
			t.Errorf("Parse(%q) was accepted", text)
		}
	}
}

func TestAmountsAreWrittenTheSpanishWay(t *testing.T) {
	cases := map[Money]string{1200: "12 €", 1205: "12,05 €", 123450: "1.234,50 €", -350: "-3,50 €", 0: "0 €"}
	for amount, want := range cases {
		if got := Format(amount); got != want {
			t.Errorf("Format(%d) = %q; want %q", amount, got, want)
		}
	}
	if got := FormatForInput(990); got != "9,90" {
		t.Errorf("FormatForInput(990) = %q; want 9,90", got)
	}
}

func TestAMarginIsProfitAndItsShareOfThePrice(t *testing.T) {
	if got := MarginOf(1000, 350); got != (Margin{Amount: 650, Percent: 65, HasPercent: true}) {
		t.Errorf("MarginOf(1000, 350) = %+v", got)
	}
	if !MarginOf(500, 800).Negative() {
		t.Error("selling under cost is not negative")
	}
	if got := MarginOf(4000, 0); got.Percent != 100 {
		t.Errorf("a free cost gives %d %%", got.Percent)
	}
	if MarginOf(0, 100).HasPercent {
		t.Error("a zero price has a percent")
	}
}

func TestSumAddsWhatEachItemGives(t *testing.T) {
	identity := func(amount Money) Money { return amount }
	if got := Sum([]Money{100, 250}, identity); got != 350 {
		t.Errorf("Sum = %d", got)
	}
	if got := Sum(nil, identity); got != 0 {
		t.Errorf("Sum of nothing = %d", got)
	}
	if got := Money(1200).Times(3); got != 3600 {
		t.Errorf("Times = %d", got)
	}
}
