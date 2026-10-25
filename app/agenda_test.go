package app

import (
	"testing"
	"time"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/projects"
	"github.com/Itros97/MUSA/sales"
)

func TestTheAgendaPutsAFairOnEachOfItsDaysAnOrderOnTheDayItCameAndAProjectWhenItIsDue(test *testing.T) {
	first, last := calendar.Date{Year: 2026, Month: 5, Day: 1}, calendar.Date{Year: 2026, Month: 5, Day: 2}
	plan := agenda{
		fairs:    []events.Event{{Name: "Market", StartDate: first, EndDate: last}},
		onTheWay: []sales.Sale{{Customer: "Ana", SoldAt: last.At(12, time.UTC)}},
		projects: []projects.Project{{Title: "Portrait", Stage: projects.Drawing, Deadline: last}, {Title: "Done", Stage: projects.Delivered, Deadline: last}},
		zone:     time.UTC,
	}
	if entries := plan.on(first); len(entries) != 1 || !entries[0].fair {
		test.Errorf("first day: %+v", entries)
	}
	if entries := plan.on(last); len(entries) != 3 || entries[1].words != "Enviar a Ana" || entries[2].words != "Portrait" {
		test.Errorf("last day: %+v", entries)
	}
	if entries := plan.on(last.AddDays(1)); len(entries) != 0 {
		test.Errorf("the day after: %+v", entries)
	}
}

func TestTheMonthCountsItsSalesAndTheProjectsDelivered(test *testing.T) {
	cases := map[string][2]int{
		"en 1 venta":                            {1, 0},
		"en 23 ventas · 2 proyectos entregados": {23, 2},
		"en 0 ventas · 1 proyecto entregado":    {0, 1},
	}
	for want, counts := range cases {
		if got := monthWords(counts[0], counts[1]); got != want {
			test.Errorf("got %q, want %q", got, want)
		}
	}
}
