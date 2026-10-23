package app

import (
	"testing"
	"time"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/sales"
)

func TestTheAgendaPutsAFairOnEachOfItsDaysAndAnOrderOnTheDayItCame(t *testing.T) {
	first, last := calendar.Date{Year: 2026, Month: 5, Day: 1}, calendar.Date{Year: 2026, Month: 5, Day: 2}
	plan := agenda{
		fairs:    []events.Event{{Name: "Market", StartDate: first, EndDate: last}},
		onTheWay: []sales.Sale{{Customer: "Ana", SoldAt: last.At(12, time.UTC)}},
		zone:     time.UTC,
	}
	if entries := plan.on(first); len(entries) != 1 || !entries[0].fair {
		t.Errorf("first day: %+v", entries)
	}
	if entries := plan.on(last); len(entries) != 2 || entries[1].words != "Enviar a Ana" {
		t.Errorf("last day: %+v", entries)
	}
	if entries := plan.on(last.AddDays(1)); len(entries) != 0 {
		t.Errorf("the day after: %+v", entries)
	}
}
