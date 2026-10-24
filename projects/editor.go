package projects

import (
	"slices"
	"strings"

	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/customers"
	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
)

// editor is a project being written, as typed: the project, who it is for, and the text fields.
type editor struct {
	project  Project
	person   customers.Customer
	handles  map[string]string // by network id
	price    string
	deadline string
	product  int // 0 none, then 1 + index in choices
	problems map[string]string
}

// startProject opens the editor of a new project in a column.
func (screens *Screens) startProject(stage Stage) {
	screens.editor = &editor{project: Project{Stage: stage}, handles: map[string]string{}, problems: map[string]string{}}
}

// edit opens the editor of a project, with its customer's details.
func (screens *Screens) edit(project Project) {
	person := screens.people[project.CustomerID]
	editing := &editor{project: project, person: person, handles: map[string]string{}, problems: map[string]string{},
		price: money.FormatForInput(project.Price)}
	for _, profile := range person.Profiles {
		editing.handles[profile.NetworkID] = profile.Handle
	}
	if !project.Deadline.Zero() {
		editing.deadline = calendar.Typed(project.Deadline)
	}
	editing.product = 1 + slices.IndexFunc(screens.choices, func(choice productChoice) bool { return choice.variantID == project.VariantID })
	if project.VariantID == "" {
		editing.product = 0
	}
	screens.editor = editing
}

func (screens *Screens) drawEditor(context *ui.Context) {
	editing := screens.editor
	title := "Nuevo proyecto"
	if editing.project.ID != "" {
		title = editing.project.Title
	}
	if context.Sheet(title, ui.SheetOptions{}, func() {
		look.Field(context, "Título", "title", &editing.project.Title, "Retrato de familia", editing.problems["title"])
		screens.customerFields(context)
		screens.workFields(context)
		screens.editorActions(context)
	}) {
		screens.editor = nil
	}
}

// customerFields are who the project is for: their name, how to reach them and their handles.
func (screens *Screens) customerFields(context *ui.Context) {
	editing := screens.editor
	theme := context.Theme
	context.Box(ui.Layout{Width: ui.Grow(), Columns: look.Columns(context, 2), Gap: theme.Gap}, func() {
		look.Field(context, "Para", "customer", &editing.person.Name, "Lucía", editing.problems["customer"])
		look.Field(context, "Contacto", "contact", &editing.person.Contact, "Email o teléfono", "")
	})
	if look.Link(context, "Editar redes") {
		screens.openNetworks()
	}
	context.Box(ui.Layout{Width: ui.Grow(), Columns: look.Columns(context, len(screens.networks)), Gap: theme.Gap}, func() {
		for _, network := range screens.networks {
			handle := editing.handles[network.ID]
			look.Field(context, network.Name, "network-"+network.ID, &handle, "@"+strings.ToLower(network.Name), "")
			editing.handles[network.ID] = handle
		}
	})
}

// workFields are what the project is: its price, when it is due and the product it makes.
func (screens *Screens) workFields(context *ui.Context) {
	editing := screens.editor
	theme := context.Theme
	context.Box(ui.Layout{Width: ui.Grow(), Columns: look.Columns(context, 2), Gap: theme.Gap}, func() {
		look.Field(context, "Precio", "price", &editing.price, "40", editing.problems["price"])
		look.Field(context, "Fecha de entrega", "deadline", &editing.deadline, "18/11/2026", editing.problems["deadline"])
	})
	options := []string{"Sin producto del catálogo"}
	for _, choice := range screens.choices {
		options = append(options, choice.label)
	}
	look.Label(context, "Producto")
	context.Select("product", &editing.product, options)
	if problem := editing.problems["product"]; problem != "" {
		context.Text(problem, ui.TextStyle{Size: theme.SmallTextSize, Color: theme.Danger})
	}
}

// editorActions are delete, move to the next column and save.
func (screens *Screens) editorActions(context *ui.Context) {
	editing := screens.editor
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.Center}, func() {
		saved := editing.project.ID != ""
		if saved && editing.project.Stage != Delivered && context.DangerIconButton("trash-2", "##Eliminar proyecto") {
			problems, err := screens.store.Delete(editing.project.ID)
			screens.afterAction(problems, err, "Proyecto eliminado")
		}
		context.Spacer()
		next := nextStage(editing.project.Stage)
		if saved && next != "" && context.IconButton("arrow-right", "Pasar a «"+StageWords(next)+"»") {
			screens.save(next)
		}
		if editing.project.Stage != Delivered && context.PrimaryButton("Guardar") {
			screens.save("")
		}
	})
}

func nextStage(stage Stage) Stage {
	index := slices.Index(Stages, stage)
	if index < 0 || index == len(Stages)-1 {
		return ""
	}
	return Stages[index+1]
}

// save keeps the customer and the project, and moves it to stage when one is given.
func (screens *Screens) save(stage Stage) {
	editing := screens.editor
	editing.problems = map[string]string{}
	if !screens.readTyped() {
		return
	}
	screens.knownPerson()
	for networkID, handle := range editing.handles {
		editing.person.Profiles = append(slices.DeleteFunc(editing.person.Profiles, func(profile customers.Profile) bool {
			return profile.NetworkID == networkID
		}), customers.Profile{NetworkID: networkID, Handle: handle})
	}
	person, problem, err := screens.customers.Save(editing.person)
	if problem != "" {
		editing.problems["customer"] = customers.ProblemWords(problem)
	}
	if err != nil || problem != "" {
		screens.tellFailure(err)
		return
	}
	editing.project.CustomerID = person.ID
	project, problems, err := screens.store.Save(editing.project)
	if screens.showProblems(problems) || err != nil {
		screens.tellFailure(err)
		return
	}
	if stage == "" {
		screens.afterAction(nil, nil, "Proyecto guardado")
		return
	}
	problems, err = screens.store.Move(project.ID, stage)
	screens.afterAction(problems, err, StageWords(stage)+": "+project.Title)
}

// knownPerson makes a new project's customer the one already saved with the same name, so
// writing «Ana» twice gives one Ana with her networks, not two.
func (screens *Screens) knownPerson() {
	editing := screens.editor
	if editing.person.ID != "" {
		return
	}
	for _, person := range screens.people {
		if strings.EqualFold(strings.TrimSpace(person.Name), strings.TrimSpace(editing.person.Name)) {
			editing.person.ID = person.ID
			return
		}
	}
}

// readTyped reads the price, the deadline and the product the editor shows as text.
func (screens *Screens) readTyped() bool {
	editing := screens.editor
	price, found := money.Parse(editing.price)
	if !found {
		editing.problems["price"] = "Escribe el precio como 40 o 12,50"
	}
	editing.project.Price = price
	editing.project.Deadline = calendar.Date{}
	if strings.TrimSpace(editing.deadline) != "" {
		deadline, read := calendar.ParseTyped(editing.deadline)
		if !read {
			editing.problems["deadline"] = "Escribe el día como 18/11/2026"
		}
		editing.project.Deadline = deadline
	}
	editing.project.VariantID = ""
	if editing.product > 0 && editing.product <= len(screens.choices) {
		editing.project.VariantID = screens.choices[editing.product-1].variantID
	}
	return len(editing.problems) == 0
}

// fieldOf is the editor field each problem is shown beside.
var fieldOf = map[Code]string{CodeBlankTitle: "title", CodeNoCustomer: "customer", CodeNegativePrice: "price", CodeProductNotDigital: "product"}

func (screens *Screens) showProblems(problems []Code) bool {
	for _, code := range problems {
		screens.editor.problems[fieldOf[code]] = ProblemWords(code)
	}
	return len(problems) > 0
}

func (screens *Screens) tellFailure(err error) {
	if err != nil {
		screens.tell(true, err.Error())
	}
}

// afterAction closes the editor when an action went through, or says why it did not.
func (screens *Screens) afterAction(problems []Code, err error, done string) {
	switch {
	case err != nil:
		screens.tell(true, err.Error())
	case len(problems) > 0:
		screens.tell(true, ProblemWords(problems[0]))
	default:
		screens.editor = nil
		screens.tell(false, done)
	}
}
