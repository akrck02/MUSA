package events

import (
	"testing"
	"time"

	"github.com/Itros97/MUSA/calendar"
)

func TestAFairIsHappeningUpcomingOrPastFromToday(t *testing.T) {
	today := calendar.Date{Year: 2026, Month: time.April, Day: 18}
	fair := Event{Status: Open, StartDate: today, EndDate: today.AddDays(1)}
	if TimingOf(fair, today) != Happening || WhenWords(fair, today) != "En curso" {
		t.Error("a fair of today is not happening")
	}
	coming := Event{Status: Open, StartDate: today.AddDays(1), EndDate: today.AddDays(2)}
	if TimingOf(coming, today) != Upcoming || WhenWords(coming, today) != "Mañana" {
		t.Errorf("tomorrow's fair: %q", WhenWords(coming, today))
	}
	coming.StartDate = today.AddDays(50)
	if WhenWords(coming, today) != "Dentro de 50 días" {
		t.Errorf("a later fair: %q", WhenWords(coming, today))
	}
	closed := fair
	closed.Status = Closed
	if TimingOf(closed, today) != Past {
		t.Error("a closed fair of today is not past")
	}
}
