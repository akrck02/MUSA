package app

import (
	"testing"

	"github.com/Itros97/MUSA/calendar"
)

func TestTheHomePageLinksLeadWhereTheySay(t *testing.T) {
	cases := []struct {
		link  string
		place Place
	}{
		{"Ver el evento", Fairs},
		{"Ir al puesto", Stand},
		{"Ver ventas", Sales},
		{"Ver los 5", Catalogue},
	}
	for _, each := range cases {
		program, window := openApp(t, true)
		program.Open(Home)
		settle(window)
		steps(t, window, click(window, each.link))
		settle(window)
		if program.place != each.place {
			t.Errorf("«%s» went to place %d, not %d", each.link, program.place, each.place)
		}
	}
}

func TestSeeTheWholeMonthOpensTheCalendar(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Home)
	settle(window)
	steps(t, window, click(window, "Ver el mes entero"))
	settle(window)
	if _, err := window.Element("Mes siguiente"); err != nil {
		t.Errorf("the month is not shown: %v", err)
	}
}

func TestTheCalendarMovesFromMonthToMonth(t *testing.T) {
	program, window := openApp(t, true)
	program.Show("calendar")
	settle(window)
	today := program.today.date
	next := calendar.Date{Year: today.Year, Month: today.Month, Day: 1}.AddMonths(1)
	steps(t, window, click(window, "Mes siguiente"))
	settle(window)
	if _, err := window.Element(calendar.MonthTitle(next)); err != nil {
		t.Fatalf("«Mes siguiente» does not show %s", calendar.MonthTitle(next))
	}
	steps(t, window, click(window, "Mes anterior"), click(window, "Mes anterior"))
	settle(window)
	previous := calendar.Date{Year: today.Year, Month: today.Month, Day: 1}.AddMonths(-1)
	if _, err := window.Element(calendar.MonthTitle(previous)); err != nil {
		t.Errorf("«Mes anterior» does not show %s", calendar.MonthTitle(previous))
	}
}

func TestNewFairFromTheCalendarOpensTheFairEditor(t *testing.T) {
	program, window := openApp(t, true)
	program.Show("calendar")
	settle(window)
	steps(t, window, click(window, "Nuevo evento"))
	settle(window)
	if _, err := window.Element("Nuevo evento"); err != nil || program.place != Fairs {
		t.Errorf("«Nueva feria» did not open a new fair (place %d): %v", program.place, err)
	}
}

func TestEmptyingTheDatabaseAsksFirstAndThenEmptiesIt(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Home)
	settle(window)
	steps(t, window, click(window, "Vaciar base de datos"))
	settle(window)
	if salesCount(t, program) == 0 {
		t.Fatalf("the database was emptied before the question was answered")
	}
	steps(t, window, click(window, "Vaciar"))
	settle(window)
	if count := salesCount(t, program); count != 0 {
		t.Errorf("%d sales are left after emptying the database", count)
	}
}

func TestTheDemoDataFillsAnEmptyDatabase(t *testing.T) {
	program, window := openApp(t, false)
	program.Open(Home)
	settle(window)
	steps(t, window, click(window, "Rellenar con datos de prueba"))
	settle(window)
	if salesCount(t, program) == 0 {
		t.Errorf("the demo data did not fill the database")
	}
}

// Switching a page's tab changes what is under the tabs, not the page: the whole screen must not
// fade and rise again, which looks like a flicker.
func TestSwitchingATabDoesNotReplayThePageEntrance(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Home)
	settle(window)
	home := program.pageName()
	steps(t, window, click(window, "Calendario"))
	if program.pageName() != home {
		t.Errorf("the calendar tab made the home page another page")
	}
	program.Open(Sales)
	settle(window)
	orders := program.pageName()
	steps(t, window, click(window, "Historial de ventas"))
	if program.pageName() != orders {
		t.Errorf("the history tab made the orders page another page")
	}
}
