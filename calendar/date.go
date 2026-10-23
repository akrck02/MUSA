// Package calendar is days without hours: when a fair starts and ends, or which day a sale was
// made. It reads and writes them as ISO text ("2026-04-18") and in Spanish words ("18 abr 2026").
package calendar

import (
	"fmt"
	"time"
)

// isoLayout is how dates are stored: as text, so text order is date order.
const isoLayout = "2006-01-02"

// Date is one day of the calendar, with no hour and no time zone.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// Of is the day a moment falls on, in its own time zone.
func Of(moment time.Time) Date {
	year, month, day := moment.Date()
	return Date{Year: year, Month: month, Day: day}
}

// Parse reads a stored date ("2026-04-18").
func Parse(text string) (Date, error) {
	moment, err := time.Parse(isoLayout, text)
	if err != nil {
		return Date{}, fmt.Errorf("%s: %q: %w", CodeInvalidDate, text, err)
	}
	return Of(moment), nil
}

// String writes the date as it is stored ("2026-04-18").
func (date Date) String() string { return date.At(0, time.UTC).Format(isoLayout) }

// At is the moment of the day at hour o'clock in zone.
func (date Date) At(hour int, zone *time.Location) time.Time {
	return time.Date(date.Year, date.Month, date.Day, hour, 0, 0, 0, zone)
}

// AddDays is the day count days later, or earlier when count is negative.
func (date Date) AddDays(count int) Date { return Of(date.At(0, time.UTC).AddDate(0, 0, count)) }

// AddMonths is the same day count months later; a day the month lacks rolls over, as in Go's time.
func (date Date) AddMonths(count int) Date { return Of(date.At(0, time.UTC).AddDate(0, count, 0)) }

// Before tells whether date comes before other.
func (date Date) Before(other Date) bool { return date.String() < other.String() }

// Zero tells whether the date was never set.
func (date Date) Zero() bool { return date == Date{} }

// CodeInvalidDate is the code of text that is not a stored date.
const CodeInvalidDate = "calendar.invalid-date"

// daysInWeek is how many days a week has.
const daysInWeek = 7

// WeekStart is the Monday of the week date is in: weeks start on Monday in Spain.
func (date Date) WeekStart() Date {
	sinceMonday := (int(date.At(0, time.UTC).Weekday()) + daysInWeek - 1) % daysInWeek
	return date.AddDays(-sinceMonday)
}

// Week is the seven days of the week date is in, Monday first.
func (date Date) Week() []Date {
	start := date.WeekStart()
	days := make([]Date, daysInWeek)
	for index := range days {
		days[index] = start.AddDays(index)
	}
	return days
}

// MonthGrid is the whole weeks that hold the month of date, Monday first: what a wall calendar
// shows, with the last days of the month before and the first of the next filling the edges.
func (date Date) MonthGrid() []Date {
	first := Date{Year: date.Year, Month: date.Month, Day: 1}
	next := first.AddMonths(1)
	var days []Date
	for day := first.WeekStart(); day.Before(next) || day.WeekStart() != day; day = day.AddDays(1) {
		days = append(days, day)
	}
	return days
}

// Within tells whether date falls between first and last, both included.
func (date Date) Within(first, last Date) bool { return !date.Before(first) && !last.Before(date) }
