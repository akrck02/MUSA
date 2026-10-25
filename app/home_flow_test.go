package app

import (
	"strconv"
	"testing"
	"time"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/projects"
)

func TestTheHomePageLinksLeadWhereTheySay(test *testing.T) {
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
		program, window := openApp(test, true)
		program.Open(Home)
		settle(window)
		steps(test, window, click(window, each.link))
		settle(window)
		if program.place != each.place {
			test.Errorf("«%s» went to place %d, not %d", each.link, program.place, each.place)
		}
	}
}

func TestSeeTheWholeMonthOpensTheCalendar(test *testing.T) {
	program, window := openApp(test, true)
	program.Open(Home)
	settle(window)
	steps(test, window, click(window, "Ver el mes entero"))
	settle(window)
	if _, err := window.Element("Mes siguiente"); err != nil {
		test.Errorf("the month is not shown: %v", err)
	}
}

func TestTheCalendarMovesFromMonthToMonth(test *testing.T) {
	program, window := openApp(test, true)
	program.Show("calendar")
	settle(window)
	today := program.today.date
	next := calendar.Date{Year: today.Year, Month: today.Month, Day: 1}.AddMonths(1)
	steps(test, window, click(window, "Mes siguiente"))
	settle(window)
	if _, err := window.Element(calendar.MonthTitle(next)); err != nil {
		test.Fatalf("«Mes siguiente» does not show %s", calendar.MonthTitle(next))
	}
	steps(test, window, click(window, "Mes anterior"), click(window, "Mes anterior"))
	settle(window)
	previous := calendar.Date{Year: today.Year, Month: today.Month, Day: 1}.AddMonths(-1)
	if _, err := window.Element(calendar.MonthTitle(previous)); err != nil {
		test.Errorf("«Mes anterior» does not show %s", calendar.MonthTitle(previous))
	}
}

func TestNewFairFromTheCalendarOpensTheFairEditor(test *testing.T) {
	program, window := openApp(test, true)
	program.Show("calendar")
	settle(window)
	steps(test, window, click(window, "Nuevo evento"))
	settle(window)
	if _, err := window.Element("Nuevo evento"); err != nil || program.place != Fairs {
		test.Errorf("«Nueva feria» did not open a new fair (place %d): %v", program.place, err)
	}
}

func TestEmptyingTheDatabaseAsksFirstAndThenEmptiesIt(test *testing.T) {
	program, window := openApp(test, true)
	program.Open(Home)
	settle(window)
	steps(test, window, click(window, "Vaciar base de datos"))
	settle(window)
	if salesCount(test, program) == 0 {
		test.Fatalf("the database was emptied before the question was answered")
	}
	steps(test, window, click(window, "Vaciar"))
	settle(window)
	if count := salesCount(test, program); count != 0 {
		test.Errorf("%d sales are left after emptying the database", count)
	}
}

func TestTheDemoDataFillsAnEmptyDatabase(test *testing.T) {
	program, window := openApp(test, false)
	program.Open(Home)
	settle(window)
	steps(test, window, click(window, "Rellenar con datos de prueba"))
	settle(window)
	if salesCount(test, program) == 0 {
		test.Errorf("the demo data did not fill the database")
	}
}

// Switching a page's tab changes what is under the tabs, not the page: the whole screen must not
// fade and rise again, which looks like a flicker.
func TestSwitchingATabDoesNotReplayThePageEntrance(test *testing.T) {
	program, window := openApp(test, true)
	program.Open(Home)
	settle(window)
	home := program.pageName()
	steps(test, window, click(window, "Calendario"))
	if program.pageName() != home {
		test.Errorf("the calendar tab made the home page another page")
	}
	program.Open(Sales)
	settle(window)
	orders := program.pageName()
	steps(test, window, click(window, "Historial de ventas"))
	if program.pageName() != orders {
		test.Errorf("the history tab made the orders page another page")
	}
}

// busyToday gives today six projects due, more than a cell or a day's detail shows at once.
func busyToday(program *App) {
	for index := range 6 {
		project := projects.Project{ID: "busy " + strconv.Itoa(index), Title: "Encargo " + strconv.Itoa(index), Deadline: program.today.date}
		program.today.plan.projects = append(program.today.plan.projects, project)
	}
}

func TestADayWithMoreThanTwoEntriesShowsTwoAndHowManyMore(test *testing.T) {
	program, window := openApp(test, true)
	program.Open(Home)
	program.home.tab = 1
	busyToday(program)
	window.Frames(2)
	entries := len(program.today.plan.on(program.today.date))
	if _, err := window.Element("+" + strconv.Itoa(entries-monthEntries)); err != nil {
		test.Errorf("today's cell does not say how many more it holds: %v", err)
	}
	if _, err := window.Element("Encargo 5"); err == nil {
		test.Errorf("today's cell shows more entries than fit")
	}
}

func TestSeeingAllOfABusyDayWidensItsDetail(test *testing.T) {
	program, window := openApp(test, true)
	program.Open(Home)
	program.home.tab = 1
	busyToday(program)
	program.home.dayShown = program.today.date
	window.Frames(2)
	if _, err := window.ElementIn("day detail", "Encargo 5"); err == nil {
		test.Fatalf("a busy day's detail lists every entry before «Ver todos»")
	}
	title := calendar.LongWords(program.today.date)
	before, _ := window.Element(title)
	steps(test, window, click(window, "Ver todos ("+strconv.Itoa(len(program.today.plan.on(program.today.date)))+")"))
	window.Wait(time.Second)
	window.Frames(2)
	after, _ := window.Element(title)
	// The card stays centred, so a wider card starts further left.
	if !program.home.dayWhole || after.Rect.X >= before.Rect.X {
		test.Errorf("«Ver todos» did not widen the day's detail (its title from %v to %v)", before.Rect.X, after.Rect.X)
	}
	if _, err := window.Element("Encargo 5"); err != nil {
		test.Errorf("the whole day does not list its last entry: %v", err)
	}
}

// Two choices of a segmented track never touch: a hovered one would merge into the chosen one.
func TestTheHomeViewsKeepRoomBetweenThem(test *testing.T) {
	program, window := openApp(test, true)
	program.Open(Home)
	window.Frames(2)
	summary, err := window.Element("Resumen")
	if err != nil {
		test.Fatal(err)
	}
	month, err := window.Element("Calendario")
	if err != nil {
		test.Fatal(err)
	}
	if month.Rect.X <= summary.Rect.X+summary.Rect.Width {
		test.Errorf("«Calendario» starts at %v, where «Resumen» ends", month.Rect.X)
	}
}
