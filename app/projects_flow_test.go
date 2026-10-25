package app

import (
	"testing"

	"github.com/Itros97/MUSA/projects"
	"github.com/Itros97/MUSA/sales"
	"github.com/akrck02/go-ui/uitest"
)

func orderID(test *testing.T, program *App, customer string) string {
	return saleOf(test, program, customer).ID
}

// saleOf is the first sale of a customer.
func saleOf(test *testing.T, program *App, customer string) sales.Sale {
	test.Helper()
	all, err := program.stores.Sales.All()
	if err != nil {
		test.Fatal(err)
	}
	for _, sale := range all {
		if sale.Customer == customer {
			return sale
		}
	}
	test.Fatalf("no sale of %s", customer)
	return sales.Sale{}
}

// projectTitled is the stored project with a title, and whether there is one.
func projectTitled(test *testing.T, program *App, title string) (projects.Project, bool) {
	test.Helper()
	all, err := program.stores.Projects.All()
	if err != nil {
		test.Fatal(err)
	}
	for _, project := range all {
		if project.Title == title {
			return project, true
		}
	}
	return projects.Project{}, false
}

// openProject opens the board and the editor of a project.
func openProject(test *testing.T, title string) (*App, *uitest.Window) {
	test.Helper()
	program, window := openApp(test, true)
	program.Open(Projects)
	settle(window)
	project, _ := projectTitled(test, program, title)
	program.projects.Edit(project.ID)
	settle(window)
	return program, window
}

func TestSearchingProjectsKeepsOnlyWhatMatches(test *testing.T) {
	program, window := openApp(test, true)
	program.Open(Projects)
	settle(window)
	steps(test, window, click(window, "Buscar"))
	settle(window)
	window.Type("taller")
	settle(window)
	if _, err := window.Element("Logo para tienda"); err != nil {
		test.Errorf("the project for «Taller Norte» is not found by «taller»")
	}
	if _, err := window.Element("Retrato de familia"); err == nil {
		test.Errorf("«Retrato de familia» still shows when searching «taller»")
	}
}

func TestTheEditorMovesAProjectToTheNextColumn(test *testing.T) {
	program, window := openProject(test, "Retrato de familia")
	steps(test, window, click(window, "Pasar a «Dibujando»"))
	settle(window)
	if moved, _ := projectTitled(test, program, "Retrato de familia"); moved.Stage != projects.Drawing {
		test.Errorf("the project is %s after «Pasar a «Dibujando»»", moved.Stage)
	}
}

func TestDeliveringAProjectRecordsItsSale(test *testing.T) {
	program, window := openProject(test, "Logo para tienda")
	before := salesCount(test, program)
	steps(test, window, click(window, "Pasar a «Entregado»"))
	settle(window)
	if after := salesCount(test, program); after != before+1 {
		test.Errorf("delivering a project left %d sales from %d", after, before)
	}
}

func TestAProjectIsDeleted(test *testing.T) {
	program, window := openProject(test, "Retrato de familia")
	steps(test, window, click(window, "Eliminar proyecto"))
	settle(window)
	if _, found := projectTitled(test, program, "Retrato de familia"); found {
		test.Errorf("the project is still there after deleting it")
	}
}

func TestAProjectsChangesAreSaved(test *testing.T) {
	program, window := openProject(test, "Retrato de familia")
	steps(test, window, typeInto(window, "Email o teléfono", "lucia@example.com"), click(window, "Guardar"))
	settle(window)
	project, _ := projectTitled(test, program, "Retrato de familia")
	person, _, err := program.stores.Customers.Find(project.CustomerID)
	if err != nil || person.Contact != "lucia@example.com" {
		test.Errorf("the contact is %q after saving (%v)", person.Contact, err)
	}
}

func TestADeliveredProjectIsDraggedBack(test *testing.T) {
	program, window := openApp(test, true)
	program.Open(Projects)
	settle(window)
	delivered, _ := projectTitled(test, program, "Retrato de gato")
	card, err := window.ElementIn(delivered.ID, "Retrato de gato")
	if err != nil {
		test.Fatal(err)
	}
	centreX, centreY := card.Rect.X+card.Rect.Width/2, card.Rect.Y+card.Rect.Height/2
	window.Drag(centreX, centreY, centreX-card.Rect.Width-40, centreY)
	settle(window)
	if moved, _ := projectTitled(test, program, "Retrato de gato"); moved.Stage != projects.ToReview {
		test.Errorf("the delivered project is %s after dragging it to «Para revisar»", moved.Stage)
	}
}

func TestTheBoardStartsCompactAndOpensUpOnRequest(test *testing.T) {
	program, window := openApp(test, true)
	program.Open(Projects)
	window.Frames(2)
	compact, err := window.Element("Retrato de familia")
	if err != nil {
		test.Fatal(err)
	}
	steps(test, window, click(window, "Vista amplia"))
	window.Frames(2)
	spacious, err := window.Element("Retrato de familia")
	if err != nil {
		test.Fatal(err)
	}
	if compact.Rect.Height >= spacious.Rect.Height/2 {
		test.Errorf("the board does not start compact: a card is %v tall, %v opened up", compact.Rect.Height, spacious.Rect.Height)
	}
	if _, err := window.Element("Vista compacta"); err != nil {
		test.Errorf("the spacious view offers no way back: %v", err)
	}
}

func TestTheBoardOpensInTheViewChosenLastTime(test *testing.T) {
	program, window := openApp(test, true)
	program.Open(Projects)
	window.Frames(2)
	steps(test, window, click(window, "Vista amplia"))
	reopened, err := New(program.stores.Database, Options{})
	if err != nil {
		test.Fatal(err)
	}
	reopened.Open(Projects)
	again, err := uitest.New(reopened.Build, uitest.Size(1280, 1600))
	if err != nil {
		test.Fatal(err)
	}
	again.Frames(2)
	if _, err := again.Element("Vista compacta"); err != nil {
		test.Errorf("MUSA forgot the spacious view: %v", err)
	}
}
