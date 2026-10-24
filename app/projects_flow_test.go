package app

import (
	"testing"

	"github.com/Itros97/MUSA/projects"
	"github.com/Itros97/MUSA/sales"
	"github.com/akrck02/go-ui/uitest"
)

func orderID(t *testing.T, program *App, customer string) string {
	return saleOf(t, program, customer).ID
}

// saleOf is the first sale of a customer.
func saleOf(t *testing.T, program *App, customer string) sales.Sale {
	t.Helper()
	all, err := program.stores.Sales.All()
	if err != nil {
		t.Fatal(err)
	}
	for _, sale := range all {
		if sale.Customer == customer {
			return sale
		}
	}
	t.Fatalf("no sale of %s", customer)
	return sales.Sale{}
}

// projectTitled is the stored project with a title, and whether there is one.
func projectTitled(t *testing.T, program *App, title string) (projects.Project, bool) {
	t.Helper()
	all, err := program.stores.Projects.All()
	if err != nil {
		t.Fatal(err)
	}
	for _, project := range all {
		if project.Title == title {
			return project, true
		}
	}
	return projects.Project{}, false
}

// openProject opens the board and the editor of a project.
func openProject(t *testing.T, title string) (*App, *uitest.Window) {
	t.Helper()
	program, window := openApp(t, true)
	program.Open(Projects)
	settle(window)
	project, _ := projectTitled(t, program, title)
	program.projects.Edit(project.ID)
	settle(window)
	return program, window
}

func TestSearchingProjectsKeepsOnlyWhatMatches(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Projects)
	settle(window)
	steps(t, window, click(window, "Buscar"))
	settle(window)
	window.Type("taller")
	settle(window)
	if _, err := window.Element("Logo para tienda"); err != nil {
		t.Errorf("the project for «Taller Norte» is not found by «taller»")
	}
	if _, err := window.Element("Retrato de familia"); err == nil {
		t.Errorf("«Retrato de familia» still shows when searching «taller»")
	}
}

func TestTheEditorMovesAProjectToTheNextColumn(t *testing.T) {
	program, window := openProject(t, "Retrato de familia")
	steps(t, window, click(window, "Pasar a dibujando"))
	settle(window)
	if moved, _ := projectTitled(t, program, "Retrato de familia"); moved.Stage != projects.Drawing {
		t.Errorf("the project is %s after «Pasar a dibujando»", moved.Stage)
	}
}

func TestDeliveringAProjectRecordsItsSale(t *testing.T) {
	program, window := openProject(t, "Logo para tienda")
	before := salesCount(t, program)
	steps(t, window, click(window, "Pasar a entregado"))
	settle(window)
	if after := salesCount(t, program); after != before+1 {
		t.Errorf("delivering a project left %d sales from %d", after, before)
	}
}

func TestAProjectIsDeleted(t *testing.T) {
	program, window := openProject(t, "Retrato de familia")
	steps(t, window, click(window, "Eliminar proyecto"))
	settle(window)
	if _, found := projectTitled(t, program, "Retrato de familia"); found {
		t.Errorf("the project is still there after deleting it")
	}
}

func TestAProjectsChangesAreSaved(t *testing.T) {
	program, window := openProject(t, "Retrato de familia")
	steps(t, window, typeInto(window, "Email o teléfono", "lucia@example.com"), click(window, "Guardar"))
	settle(window)
	project, _ := projectTitled(t, program, "Retrato de familia")
	person, _, err := program.stores.Customers.Find(project.CustomerID)
	if err != nil || person.Contact != "lucia@example.com" {
		t.Errorf("the contact is %q after saving (%v)", person.Contact, err)
	}
}
