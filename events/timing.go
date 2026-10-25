package events

import (
	"fmt"
	"time"

	"github.com/Itros97/MUSA/calendar"
)

// Timing is where a fair stands from today: happening now, still to come, or past (or closed).
type Timing int

// The timings, in the order the list shows them.
const (
	Happening Timing = iota
	Upcoming
	Past
)

// TimingOf tells where a fair stands on today. A closed fair is past, whatever its dates.
func TimingOf(event Event, today calendar.Date) Timing {
	switch {
	case event.Status == Closed || event.EndDate.Before(today):
		return Past
	case today.Before(event.StartDate):
		return Upcoming
	default:
		return Happening
	}
}

// DaysUntil is how many days are left until the fair starts.
func DaysUntil(event Event, today calendar.Date) int {
	return int(event.StartDate.At(0, time.UTC).Sub(today.At(0, time.UTC)).Hours() / 24)
}

// WhenWords is the fair's timing in a few words: "En curso", "Mañana", "Dentro de 50 días", or its
// dates when it is past.
func WhenWords(event Event, today calendar.Date) string {
	switch TimingOf(event, today) {
	case Happening:
		return "En curso"
	case Upcoming:
		days := DaysUntil(event, today)
		if days == 1 {
			return "Mañana"
		}
		return fmt.Sprintf("Dentro de %d días", days)
	default:
		return statusWords[event.Status]
	}
}
