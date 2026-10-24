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

// DayWords is a day without its year, for lists that are all of one year: "24 oct".
func DayWords(date Date) string { return fmt.Sprintf("%d %s", date.Day, monthName(date)) }

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

// shortWeekdays are the days as a calendar's heading writes them, Monday first.
var shortWeekdays = [...]string{"lun", "mar", "mié", "jue", "vie", "sáb", "dom"}

// ShortWeekday is the day of the week of date as a calendar's heading writes it: "mié".
func ShortWeekday(date Date) string {
	return shortWeekdays[(int(date.At(0, time.UTC).Weekday())+daysInWeek-1)%daysInWeek]
}

// MonthWords is a month and its year: "noviembre de 2026".
func MonthWords(date Date) string {
	return fmt.Sprintf("%s de %d", fullMonthNames[date.Month-1], date.Year)
}

// LongRangeWords writes when something happens as people say it, the year only when it is not
// this year's: "Sábado 7 de noviembre" for one day, "12 y 13 de diciembre" for two, "12 al 14 de
// diciembre" for more, "30 de noviembre al 2 de diciembre" across months.
func LongRangeWords(start, end Date, thisYear int) string {
	year := ""
	if end.Year != thisYear {
		year = fmt.Sprintf(" de %d", end.Year)
	}
	month := func(date Date) string { return fullMonthNames[date.Month-1] }
	switch {
	case start == end:
		weekday := weekdayNames[start.At(0, time.UTC).Weekday()]
		return fmt.Sprintf("%s%s %d de %s%s", strings.ToUpper(weekday[:1]), weekday[1:], start.Day, month(start), year)
	case start.Month != end.Month || start.Year != end.Year:
		return fmt.Sprintf("%d de %s al %d de %s%s", start.Day, month(start), end.Day, month(end), year)
	case end == start.AddDays(1):
		return fmt.Sprintf("%d y %d de %s%s", start.Day, end.Day, month(end), year)
	}
	return fmt.Sprintf("%d al %d de %s%s", start.Day, end.Day, month(end), year)
}
