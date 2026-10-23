package calendar

import "fmt"

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
