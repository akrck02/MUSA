package demo

import (
	"slices"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/customers"
	"github.com/Itros97/MUSA/money"
	"github.com/Itros97/MUSA/projects"
)

// commission is one made-up project: for whom, on which networks, how far it has got and when it is due.
type commission struct {
	title, customer string
	profiles        []customers.Profile
	stage           projects.Stage
	dueInDays       int
	price           money.Money
	product         string // a key of filler.variants, or empty
}

// projects fills the board: two to start, one being drawn, one to review that is late, one delivered.
func (filler *filler) projects(today calendar.Date) {
	instagram := func(handle string) customers.Profile {
		return customers.Profile{NetworkID: "instagram", Handle: handle}
	}
	bluesky := func(handle string) customers.Profile { return customers.Profile{NetworkID: "bluesky", Handle: handle} }
	x := func(handle string) customers.Profile { return customers.Profile{NetworkID: "x", Handle: handle} }
	for _, each := range []commission{
		{"Retrato de familia", "Lucía", []customers.Profile{instagram("lucia.dibuja"), bluesky("lucia.bsky.social")}, projects.ToStart, 6, 4000, "Retrato digital · Busto"},
		{"Fondos con su perro", "Marcos", []customers.Profile{instagram("marcos_y_toby")}, projects.ToStart, 9, 1500, ""},
		{"Retrato de pareja", "Ana", []customers.Profile{x("ana_lee"), instagram("ana.lee")}, projects.Drawing, 1, 4000, "Retrato digital · Busto"},
		{"Logo para tienda", "Taller Norte", []customers.Profile{instagram("tallernorte")}, projects.ToReview, -1, 6000, ""},
		{"Retrato de gato", "Pablo", []customers.Profile{bluesky("pablo.bsky.social")}, projects.Delivered, -5, 4000, "Retrato digital · Busto"},
	} {
		filler.commission(each, today)
	}
}

func (filler *filler) commission(each commission, today calendar.Date) {
	if filler.err != nil {
		return
	}
	person, problem, err := filler.stores.Customers.Save(customers.Customer{Name: each.customer, Profiles: each.profiles})
	filler.stop(len(problem), problem, err)
	project, problems, err := filler.stores.Projects.Save(projects.Project{Title: each.title, CustomerID: person.ID,
		Deadline: today.AddDays(each.dueInDays), Price: each.price, VariantID: filler.variants[each.product]})
	filler.stop(len(problems), problems, err)
	// Walk it along the board, as the artist would, up to its stage.
	for _, stage := range projects.Stages[1 : slices.Index(projects.Stages, each.stage)+1] {
		problems, err := filler.stores.Projects.Move(project.ID, stage)
		filler.stop(len(problems), problems, err)
	}
}
