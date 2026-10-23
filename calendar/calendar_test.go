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

func TestADateIsTypedAsDayMonthAndYear(t *testing.T) {
	date := Date{2026, time.April, 18}
	for _, text := range []string{"18/4/2026", " 18-04-2026 "} {
		if got, found := ParseTyped(text); !found || got != date {
			t.Errorf("ParseTyped(%q) = %v %v", text, got, found)
		}
	}
	for _, text := range []string{"", "31/2/2026", "mañana"} {
		if _, found := ParseTyped(text); found {
			t.Errorf("ParseTyped(%q) was accepted", text)
		}
	}
	if Typed(date) != "18/4/2026" {
		t.Errorf("Typed = %q", Typed(date))
	}
}

func TestADateIsWrittenOutAsPeopleSayIt(t *testing.T) {
	date := Date{Year: 2026, Month: time.October, Day: 23}
	if words := LongWords(date); words != "viernes, 23 de octubre" {
		t.Errorf("got %q", words)
	}
}

func TestADayWithoutItsYearReadsShort(t *testing.T) {
	if words := DayWords(Date{Year: 2026, Month: time.October, Day: 24}); words != "24 oct" {
		t.Errorf("DayWords = %q", words)
	}
}

func TestAWeekStartsOnMonday(t *testing.T) {
	wednesday := Date{Year: 2026, Month: time.November, Day: 4}
	week := wednesday.Week()
	if week[0] != (Date{Year: 2026, Month: time.November, Day: 2}) || week[6] != (Date{Year: 2026, Month: time.November, Day: 8}) {
		t.Errorf("week %v", week)
	}
	if ShortWeekday(week[0]) != "lun" || ShortWeekday(wednesday) != "mié" || ShortWeekday(week[6]) != "dom" {
		t.Errorf("names %s %s %s", ShortWeekday(week[0]), ShortWeekday(wednesday), ShortWeekday(week[6]))
	}
	if sunday := week[6]; sunday.WeekStart() != week[0] {
		t.Errorf("a Sunday belongs to the week before: %v", sunday.WeekStart())
	}
}

func TestAMonthGridHoldsWholeWeeks(t *testing.T) {
	grid := Date{Year: 2026, Month: time.November, Day: 15}.MonthGrid()
	// November 2026 starts on a Sunday and ends on a Monday: six weeks, from 26 October to 6 December.
	if len(grid) != 42 || grid[0] != (Date{Year: 2026, Month: time.October, Day: 26}) || grid[41] != (Date{Year: 2026, Month: time.December, Day: 6}) {
		t.Errorf("%d days, from %v to %v", len(grid), grid[0], grid[len(grid)-1])
	}
	if MonthWords(grid[10]) != "noviembre de 2026" {
		t.Errorf("MonthWords = %q", MonthWords(grid[10]))
	}
}

func TestADayIsWithinARangeThatIncludesItsEnds(t *testing.T) {
	first, last := Date{Year: 2026, Month: 5, Day: 1}, Date{Year: 2026, Month: 5, Day: 3}
	if !first.Within(first, last) || !last.Within(first, last) || last.AddDays(1).Within(first, last) {
		t.Error("the ends of a range are not in it, or the day after is")
	}
}
