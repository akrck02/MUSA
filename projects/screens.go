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

	note        string
	noteWarning bool
	noteUntil   time.Time
}

// productChoice is a digital product a project can make, as the editor lists it.
type productChoice struct {
	variantID, label string
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
			choices = append(choices, productChoice{variantID: variant.ID, label: label})
		}
	}
	return choices, nil
}

// Back closes the editor; false when only the board was open.
func (screens *Screens) Back() bool {
	if screens.editor == nil {
		return false
	}
	screens.editor = nil
	return true
}

// Draw draws the board, and the editor over it when one is open.
func (screens *Screens) Draw(context *ui.Context) {
	look.PageHeader(context, "Proyectos", screens.subtitle(), func() {
		if context.PrimaryIconButton("plus", "Nuevo proyecto") {
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
	if time.Now().Before(screens.noteUntil) {
		context.Toast(screens.note, screens.noteWarning)
		context.RequestFrame(time.Until(screens.noteUntil))
	}
}

// subtitle counts the projects still to deliver, and those due this week.
func (screens *Screens) subtitle() string {
	today := calendar.Of(screens.now())
	open, thisWeek := 0, 0
	for _, project := range screens.projects {
		if project.Stage == Delivered {
			continue
		}
		open++
		if !project.Deadline.Zero() && project.Deadline.Before(today.WeekStart().AddDays(7)) {
			thisWeek++
		}
	}
	return fmt.Sprintf("%d por entregar · %d vencen esta semana", open, thisWeek)
}

func (screens *Screens) tell(warning bool, words string) {
	screens.note, screens.noteWarning, screens.noteUntil = words, warning, time.Now().Add(noteTime)
	screens.Reload()
}

// ShowEditor opens the editor of the first project, for captures.
func (screens *Screens) ShowEditor() {
	if len(screens.projects) > 0 {
		screens.edit(screens.projects[0])
	}
}
