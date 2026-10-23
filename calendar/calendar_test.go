package calendar

import (
	"testing"
	"time"
)

func TestDatesAreWrittenInSpanishRepeatingOnlyWhatChanges(t *testing.T) {
	day := func(year int, month time.Month, number int) Date { return Date{year, month, number} }
	cases := map[string]string{
		RangeWords(day(2026, 1, 8), day(2026, 1, 8)):   "8 ene 2026",
		RangeWords(day(2026, 4, 18), day(2026, 4, 19)): "18–19 abr 2026",
		RangeWords(day(2026, 4, 30), day(2026, 5, 2)):  "30 abr – 2 may 2026",
		RangeWords(day(2025, 12, 30), day(2026, 1, 2)): "30 dic 2025 – 2 ene 2026",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("%q; want %q", got, want)
		}
	}
}

func TestADateIsStoredAsISOTextAndReadBack(t *testing.T) {
	date := Date{2026, time.April, 18}
	if date.String() != "2026-04-18" {
		t.Errorf("String = %q", date.String())
	}
	if parsed, err := Parse("2026-04-18"); err != nil || parsed != date {
		t.Errorf("Parse = %v %v", parsed, err)
	}
	if _, err := Parse("18/04/2026"); err == nil {
		t.Error("a date in another layout was read")
	}
	if got := (Date{2026, time.January, 31}).AddDays(1); got != (Date{2026, time.February, 1}) {
		t.Errorf("AddDays = %v", got)
	}
}
