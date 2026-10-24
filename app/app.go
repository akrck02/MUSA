// Package app is MUSA as a whole: the theme, the places a person moves between (home, the stand,
// catalogue, fairs, orders) and the frame around them, a bar along the bottom. Each place's screens live in its own package.
package app

import (
	"time"

	"github.com/akrck02/bubble"
	"github.com/akrck02/bubble/goui"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/customers"
	"github.com/Itros97/MUSA/demo"
	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/projects"
	"github.com/Itros97/MUSA/sales"
	"github.com/Itros97/MUSA/stand"
	"github.com/Itros97/MUSA/storage"
)

// Accent is MUSA's colour: a deep slate, quiet like ink, so the artwork brings the colour.
const Accent = "#3F4A5A"

// Place is one of the parts of MUSA the index leads to.
type Place int

// The places, in the index's order.
const (
	Home Place = iota
	Catalogue
	Fairs
	Sales
	Stand
	Projects
)

// Options are how the program was started.
type Options struct {
	Dark bool
	// DevelopmentTools shows, on the home page, buttons that fill the database with toy data or
	// empty it.
	DevelopmentTools bool
	// OpenLink opens a web address in the system's browser; nil where nothing can.
	OpenLink func(link string) bool
	// Now is the clock; nil for the real one. A fixed one lets captures show a chosen day.
	Now func() time.Time
}

// App is everything MUSA remembers between frames.
type App struct {
	stores  demo.Stores
	themes  [2]ui.Theme // the light theme and the dark one
	dark    bool
	options Options
	place   Place

	catalogue *catalogue.Screens
	fairs     *events.Screens
	sales     *sales.Screens
	stand     *stand.Screens
	projects  *projects.Screens
	home      homeState
	today     today
}

// New opens MUSA over an open database.
func New(database *storage.Database, options Options) (*App, error) {
	light, err := newTheme(bubble.ModeLight)
	if err != nil {
		return nil, err
	}
	dark, err := newTheme(bubble.ModeDark)
	if err != nil {
		return nil, err
	}
	now := time.Now
	if options.Now != nil {
		now = options.Now
	}
	products := catalogue.NewStore(database, now, storage.NewID)
	sold := sales.NewStore(database, products, now, storage.NewID)
	fairs := events.NewStore(database, sold, now, storage.NewID)
	people := customers.NewStore(database, now, storage.NewID)
	commissions := projects.NewStore(database, products, people, now, storage.NewID)
	program := &App{
		stores: demo.Stores{Database: database, Products: products, Sales: sold, Events: fairs, Customers: people,
			Projects: commissions, Now: now, NewID: storage.NewID},
		themes:  [2]ui.Theme{light, dark},
		dark:    options.Dark,
		options: options,
	}
	if err := program.chosenTheme(); err != nil {
		return nil, err
	}
	program.catalogue = catalogue.NewScreens(products, storage.NewID)
	program.catalogue.CountSalesWith(func() (map[string]int, error) {
		made, err := sold.All()
		today := calendar.Of(now())
		return sales.UnitsSince(made, calendar.Date{Year: today.Year, Month: today.Month, Day: 1}.At(0, now().Location())), err
	})
	program.fairs = events.NewScreens(fairs, products, storage.NewID, now)
	program.sales = sales.NewScreens(sold, products, storage.NewID, now)
	program.sales.NameEvents(func(id string) string {
		event, _, _ := fairs.Find(id) // a fair that cannot be read is named by its channel instead
		return event.Name
	})
	program.stand = stand.NewScreens(fairs, products, sold, storage.NewID, now)
	program.fairs.OpenStandWith(func(event events.Event) {
		program.Open(Stand)
		program.stand.Choose(stand.Fair{Event: event, Events: fairs, Products: products})
	})
	program.projects = projects.NewScreens(commissions, people, products, storage.NewID, now, options.OpenLink)
	program.Open(Home)
	return program, nil
}

// Open moves to a place, reading its data again: other places may have changed it.
func (program *App) Open(place Place) {
	program.place = place
	if place == Home {
		program.today = readToday(program.stores, program.stores.Now())
	}
	switch place {
	case Catalogue:
		program.catalogue.Reload()
	case Fairs:
		program.fairs.Reload()
	case Sales:
		program.sales.Reload()
	case Stand:
		program.stand.Reload()
	case Projects:
		program.projects.Reload()
	}
}

// newTheme is Bubble's paper theme in a mode, on MUSA's accent and with the design's text.
func newTheme(mode bubble.Mode) (ui.Theme, error) {
	accent, err := bubble.ParseColor(Accent)
	if err != nil {
		return ui.Theme{}, err
	}
	design, err := bubble.NewTheme(bubble.Options{Style: bubble.StylePaper, Mode: mode, Density: bubble.DensityComfortable, Accent: accent})
	if err != nil {
		return ui.Theme{}, err
	}
	return pageTheme(goui.Theme(design)), nil
}

// themeSetting is the setting that keeps the theme the artist chose: themeWords[true] or [false].
const themeSetting = "theme"

var themeWords = map[bool]string{true: "dark", false: "light"}

// Dark tells whether MUSA is drawn with its dark theme.
func (program *App) Dark() bool { return program.dark }

// SetDark draws MUSA with its dark theme, or its light one, from the next frame on, and keeps the
// choice for the next time MUSA opens. The theme changes even when the choice cannot be kept.
func (program *App) SetDark(dark bool) error {
	program.dark = dark
	return program.stores.Database.SetSetting(themeSetting, themeWords[dark])
}

// chosenTheme is the theme the artist last chose; without a choice, the one MUSA was opened with.
func (program *App) chosenTheme() error {
	words, found, err := program.stores.Database.Setting(themeSetting)
	if found {
		program.dark = words == themeWords[true]
	}
	return err
}

// Build draws one frame of MUSA; it is the function given to window.Run.
func (program *App) Build(context *ui.Context) {
	context.Theme = program.themes[0]
	if program.dark {
		context.Theme = program.themes[1]
	}
	if context.Back() && !program.back() {
		program.Open(Home)
	}
	program.frame(context)
}

// back hands the system's back to the open place; false means it had nothing to close.
func (program *App) back() bool {
	switch program.place {
	case Catalogue:
		return program.catalogue.Back()
	case Fairs:
		return program.fairs.Back()
	case Sales:
		return program.sales.Back()
	case Stand:
		return program.stand.Back()
	case Projects:
		return program.projects.Back()
	}
	return false
}

func (program *App) page(context *ui.Context) {
	switch program.place {
	case Catalogue:
		program.catalogue.Draw(context)
	case Fairs:
		program.fairs.Draw(context)
	case Sales:
		program.sales.Draw(context)
	case Stand:
		program.stand.Draw(context)
	case Projects:
		program.projects.Draw(context)
	default:
		program.drawHome(context)
	}
}

// Stores are the stores MUSA reads and writes through, for the demo to fill.
func (program *App) Stores() demo.Stores { return program.stores }

// Show opens one of the pages inside a place, for captures and tests: "product" (the editor of the first
// product of one version), "versions" (of two), "fair" (the open fair's detail), "order" (a new order), "project" (a project's editor), "networks" (the social networks), "history" (the sales history), "calendar" (the month on the home page), "stand" (selling
// from the whole stock), "stand-basket" (with a basket) or "stand-cash" (its cash dialog).
func (program *App) Show(page string) {
	switch page {
	case "product", "versions":
		program.Open(Catalogue)
		program.catalogue.EditFirstWith(map[string]int{"product": 1, "versions": 2}[page])
	case "digital":
		program.Open(Catalogue)
		program.catalogue.ShowDigital()
	case "fair":
		program.Open(Fairs)
		program.fairs.OpenFirstOpen()
	case "order":
		program.Open(Sales)
		program.sales.StartOrder()
	case "project":
		program.Open(Projects)
		program.projects.ShowEditor()
	case "networks":
		program.Open(Projects)
		program.projects.ShowNetworks()
	case "calendar":
		program.Open(Home)
		program.home.tab = 1
	case "history":
		program.Open(Sales)
		program.sales.ShowHistory()
	case "stand", "stand-basket":
		program.Open(Stand)
		program.stand.Choose(stand.WholeStock{Sales: program.stores.Sales, Products: program.stores.Products, Now: program.stores.Now})
		if page == "stand-basket" {
			program.stand.ShowBasket()
		}
	default:
		program.showFairStand(page)
	}
}

// showFairStand opens today's fair on the stand at a step of a sale, for the captures of the
// author's design: "fair" with a basket, "fair-pay", "fair-cash", "fair-short", "fair-card" and
// "fair-paid".
func (program *App) showFairStand(page string) {
	steps := map[string]func(){
		"fair-stand": program.stand.ShowBasket,
		"fair-pay":   program.stand.ShowPaymentChoice,
		"fair-cash":  func() { program.stand.ShowCashPayment(1000, 1000) },
		"fair-short": func() { program.stand.ShowCashPayment(500) },
		"fair-card":  program.stand.ShowCardPayment,
		"fair-paid":  func() { program.stand.ShowPaid(2000) },
	}
	show, found := steps[page]
	if !found {
		return
	}
	program.Open(Stand)
	if program.stand.ChooseTodaysFair() {
		show()
	}
}
