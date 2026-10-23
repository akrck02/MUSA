package calendar

import (
	"fmt"
	"strings"
	"time"
)

// monthNames are the short Spanish names of the months, January first.
var monthNames = [...]string{"ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"}

// Words writes the date the Spanish way, with a short month: "18 abr 2026".
func Words(date Date) string { return fmt.Sprintf("%d %s %d", date.Day, monthName(date), date.Year) }

// RangeWords writes two dates repeating only what changes: "18–19 abr 2026", "30 abr – 2 may 2026".
func RangeWords(start, end Date) string {
	switch {
	case start == end:
		return Words(start)
	case start.Year != end.Year:
		return Words(start) + " – " + Words(end)
	case start.Month != end.Month:
		return fmt.Sprintf("%d %s – %s", start.Day, monthName(start), Words(end))
	default:
		return fmt.Sprintf("%d–%s", start.Day, Words(end))
	}
}

func monthName(date Date) string { return monthNames[date.Month-1] }

// Typed writes a date the way the artist types one: "18/4/2026".
func Typed(date Date) string { return fmt.Sprintf("%d/%d/%d", date.Day, date.Month, date.Year) }

// ParseTyped reads a date typed as day/month/year ("18/4/2026", "18-04-2026"); found is false for
// anything else, including days a month does not have.
func ParseTyped(text string) (date Date, found bool) {
	var day, month, year int
	clean := strings.ReplaceAll(strings.TrimSpace(text), "-", "/")
	if _, err := fmt.Sscanf(clean, "%d/%d/%d", &day, &month, &year); err != nil {
		return Date{}, false
	}
	date = Date{Year: year, Month: time.Month(month), Day: day}
	if Of(date.At(0, time.UTC)) != date {
		return Date{}, false // 31/2 would roll over into March
	}
	return date, true
}

// fullMonthNames and weekdayNames are the whole Spanish names, for a date written out in full.
var (
	fullMonthNames = [...]string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto",
		"septiembre", "octubre", "noviembre", "diciembre"}
	weekdayNames = [...]string{"domingo", "lunes", "martes", "miércoles", "jueves", "viernes", "sábado"}
)

// LongWords writes a date out as people say it: "viernes, 23 de octubre".
func LongWords(date Date) string {
	weekday := date.At(0, time.UTC).Weekday()
	return fmt.Sprintf("%s, %d de %s", weekdayNames[weekday], date.Day, fullMonthNames[date.Month-1])
}
