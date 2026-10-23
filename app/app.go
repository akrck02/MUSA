// Package app is MUSA as a whole: the theme, the places a person moves between (home, catalogue,
// fairs, sales) and the frame around them, an index on the left on a wide window and a bar at the
// bottom on a phone. Each place's screens live in its own package.
package app

import (
	"time"

	"github.com/akrck02/bubble"
	"github.com/akrck02/bubble/goui"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/demo"
	"github.com/Itros97/MUSA/events"
	"github.com/Itros97/MUSA/sales"
	"github.com/Itros97/MUSA/storage"
)

// Accent is MUSA's colour: the terracotta of its studio palette.
const Accent = "#9A4532"

// Place is one of the parts of MUSA the index leads to.
type Place int

// The places, in the index's order.
const (
	Home Place = iota
	Catalogue
	Fairs
	Sales
)

// Options are how the program was started.
type Options struct {
	Dark bool
	// DevelopmentTools shows, on the home page, buttons that fill the database with toy data or
	// empty it.
	DevelopmentTools bool
}

// App is everything MUSA remembers between frames.
type App struct {
	stores  demo.Stores
	theme   ui.Theme
	options Options
	place   Place

	catalogue *catalogue.Screens
	fairs     *events.Screens
	sales     *sales.Screens
	home      homeState
}

// New opens MUSA over an open database.
func New(database *storage.Database, options Options) (*App, error) {
	mode := bubble.ModeLight
	if options.Dark {
		mode = bubble.ModeDark
	}
	accent, err := bubble.ParseColor(Accent)
	if err != nil {
		return nil, err
	}
	design, err := bubble.NewTheme(bubble.Options{Style: bubble.StylePaper, Mode: mode, Density: bubble.DensityComfortable, Accent: accent})
	if err != nil {
		return nil, err
	}
	now := time.Now
	products := catalogue.NewStore(database, now, storage.NewID)
	sold := sales.NewStore(database, products, now, storage.NewID)
	fairs := events.NewStore(database, sold, now, storage.NewID)
	program := &App{
		stores:  demo.Stores{Database: database, Products: products, Sales: sold, Events: fairs, Now: now, NewID: storage.NewID},
		theme:   goui.Theme(design),
		options: options,
	}
	program.catalogue = catalogue.NewScreens(products, storage.NewID)
	program.fairs = events.NewScreens(fairs, products, storage.NewID, now)
	program.sales = sales.NewScreens(sold, products, storage.NewID, now)
	return program, nil
}

// Open moves to a place, reading its data again: other places may have changed it.
func (program *App) Open(place Place) {
	program.place = place
	switch place {
	case Catalogue:
		program.catalogue.Reload()
	case Fairs:
		program.fairs.Reload()
	case Sales:
		program.sales.Reload()
	}
}

// Build draws one frame of MUSA; it is the function given to window.Run.
func (program *App) Build(context *ui.Context) {
	context.Theme = program.theme
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
	default:
		program.drawHome(context)
	}
}

// Stores are the stores MUSA reads and writes through, for the demo to fill.
func (program *App) Stores() demo.Stores { return program.stores }

// Show opens one of the pages inside a place, for captures and tests: "product" (the first
// product's editor), "fair" (the open fair's detail) or "order" (a new order).
func (program *App) Show(page string) {
	switch page {
	case "product":
		program.Open(Catalogue)
		program.catalogue.EditFirst()
	case "fair":
		program.Open(Fairs)
		program.fairs.OpenFirstOpen()
	case "order":
		program.Open(Sales)
		program.sales.StartOrder()
	}
}
