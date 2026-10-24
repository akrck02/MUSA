package projects

import (
	"fmt"
	"time"

	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/customers"
	"github.com/Itros97/MUSA/look"
)

// noteTime is how long a note about what an action did stays on screen.
const noteTime = 3 * time.Second

// Screens are the projects pages: the board, and the editor of one project over it.
type Screens struct {
	store     *Store
	customers *customers.Store
	products  *catalogue.Store
	newID     func() string
	now       func() time.Time
	openLink  func(link string) bool

	projects []Project
	people   map[string]customers.Customer
	networks []customers.Network
	choices  []productChoice
	loadErr  error
	editor   *editor
	settings *networkSettings // the networks sheet, while open
	// searching shows the search field; query is what is typed in it.
	searching bool
	query     string

	note        string
	noteWarning bool
	noteUntil   time.Time
}

// productChoice is a digital product a project can make, as the editor lists it.
type productChoice struct {
	variantID, label string
	product          string // its product's name alone
}

// NewScreens draws the projects kept in store; openLink opens a customer's profile, and may be
// nil where nothing can open links.
func NewScreens(store *Store, people *customers.Store, products *catalogue.Store, newID func() string,
	now func() time.Time, openLink func(link string) bool) *Screens {
	return &Screens{store: store, customers: people, products: products, newID: newID, now: now, openLink: openLink}
}

// Reload reads the projects, their customers, the networks and the digital products again.
func (screens *Screens) Reload() {
	screens.projects, screens.loadErr = screens.store.All()
	if screens.loadErr != nil {
		return
	}
	screens.networks, screens.loadErr = screens.customers.Networks()
	if screens.loadErr != nil {
		return
	}
	all, err := screens.customers.All()
	screens.people, screens.loadErr = map[string]customers.Customer{}, err
	for _, person := range all {
		screens.people[person.ID] = person
	}
	screens.choices, screens.loadErr = digitalChoices(screens.products)
}

func digitalChoices(products *catalogue.Store) ([]productChoice, error) {
	all, err := products.All()
	if err != nil {
		return nil, err
	}
	var choices []productChoice
	for _, product := range all {
		if product.Kind != catalogue.Digital || product.Status == catalogue.Archived {
			continue
		}
		for _, variant := range product.Variants {
			label := product.Name
			if variant.Name != "" {
				label += " · " + variant.Name
			}
			choices = append(choices, productChoice{variantID: variant.ID, label: label, product: product.Name})
		}
	}
	return choices, nil
}

// Back closes the editor or the networks sheet; false when only the board was open.
func (screens *Screens) Back() bool {
	if screens.editor == nil && screens.settings == nil {
		return false
	}
	screens.editor, screens.settings = nil, nil
	return true
}

// searchWidth is the width of the search field, in pixels.
const searchWidth = 320

// Draw draws the board, and the editor over it when one is open.
func (screens *Screens) Draw(context *ui.Context) {
	look.PageHeader(context, "Proyectos", screens.subtitle(), func() {
		if screens.searching {
			context.Box(ui.Layout{Width: ui.Fixed(searchWidth)}, func() {
				context.TextFieldWith("search", &screens.query, ui.TextFieldOptions{Placeholder: "Buscar por título o cliente"})
			})
		}
		if look.ToolButton(context, "search", "Buscar") {
			screens.searching = !screens.searching
		}
		if look.MainButton(context, "plus", "Nuevo proyecto") {
			screens.startProject(ToStart)
		}
	})
	if screens.loadErr != nil {
		look.Empty(context, "triangle-alert", "No se pudieron leer los proyectos", screens.loadErr.Error(), "")
		return
	}
	screens.drawBoard(context)
	if screens.editor != nil {
		screens.drawEditor(context)
	}
	if screens.settings != nil {
		screens.drawNetworks(context)
	}
	if time.Now().Before(screens.noteUntil) {
		context.Toast(screens.note, screens.noteWarning)
		context.RequestFrame(time.Until(screens.noteUntil))
	}
}

// subtitle counts the projects, and those still to deliver that are due from today to the end of
// the week.
func (screens *Screens) subtitle() string {
	today := calendar.Of(screens.now())
	thisWeek := 0
	for _, project := range screens.projects {
		due := project.Deadline
		if project.Stage != Delivered && !due.Zero() && !due.Before(today) && due.Before(today.WeekStart().AddDays(7)) {
			thisWeek++
		}
	}
	return fmt.Sprintf("%s · %d por entregar esta semana", projectCount(len(screens.projects)), thisWeek)
}

func (screens *Screens) tell(warning bool, words string) {
	screens.note, screens.noteWarning, screens.noteUntil = words, warning, time.Now().Add(noteTime)
	screens.Reload()
}

// Edit opens the editor of the project with this id, when there is one.
func (screens *Screens) Edit(id string) {
	for _, project := range screens.projects {
		if project.ID == id {
			screens.edit(project)
			return
		}
	}
}

// ShowEditor opens the editor of the first project, for captures.
func (screens *Screens) ShowEditor() {
	if len(screens.projects) > 0 {
		screens.edit(screens.projects[0])
	}
}
