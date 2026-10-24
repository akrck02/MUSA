// Command musa is MUSA: the inventory and business app of an artist who also sells physical stock.
// The same program runs on the desktop and, packaged with Go UI for Android's go-ui-apk, on a phone.
//
//	go run ./cmd/musa                   # the app, over the database in your configuration folder
//	go run ./cmd/musa -demo             # a throwaway database filled with toy data
//	go run ./cmd/musa -database my.db   # another database file
//	go run ./cmd/musa -demo -capture home.png   # draw one frame to a picture and exit
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/akrck02/go-renderer/opengl/overlay"

	"github.com/akrck02/go-ui/window"

	"github.com/Itros97/MUSA/app"
	"github.com/Itros97/MUSA/demo"
	"github.com/Itros97/MUSA/storage"
)

func main() {
	options := window.Options{Title: "MUSA", Width: 1280, Height: 820}
	options.RegisterFlags(flag.CommandLine)
	demoMode := flag.Bool("demo", false, "open a throwaway database in memory, filled with toy data")
	path := flag.String("database", "", "the database file; empty for the one in the app's folder")
	place := flag.Int("place", 0, "the place to open on: 0 home, 1 catalogue, 2 fairs, 3 orders, 4 stand, 5 projects (for captures; see app.Show)")
	page := flag.String("show", "", "a page inside the place to open on: product, digital, fair, order, history, calendar, project, networks or stand… (for captures; see app.Show)")
	tools := flag.Bool("tools", true, "show the development tools on the home page")
	today := flag.String("today", "", "a day as 2026-11-04 to open on instead of today, for captures")
	flag.Parse()
	if !wasSet("dark") {
		options.Dark = window.SystemPrefersDark()
	}
	// MUSA writes in the system's own font, as Bubble 4's system-ui does; without it, Go's fonts
	// still read well, so a missing font is not worth stopping for.
	options.Fonts, _ = overlay.SystemFonts()

	clock, err := clockFor(*today)
	if err != nil {
		fmt.Fprintln(os.Stderr, "musa:", err)
		os.Exit(1)
	}
	if err := run(options, *demoMode, *path, app.Place(*place), *page, *tools, clock); err != nil {
		fmt.Fprintln(os.Stderr, "musa:", err)
		os.Exit(1)
	}
}

// middayHour is the hour a fixed day's clock shows: the middle of a working day.
const middayHour = 12

// clockFor is a clock stopped at noon of day ("2026-11-04"), or nil for the real one when day is empty.
func clockFor(day string) (func() time.Time, error) {
	if day == "" {
		return nil, nil
	}
	date, err := time.ParseInLocation(time.DateOnly, day, time.Local)
	if err != nil {
		return nil, err
	}
	noon := date.Add(middayHour * time.Hour)
	return func() time.Time { return noon }, nil
}

func run(options window.Options, demoMode bool, path string, place app.Place, page string, tools bool, clock func() time.Time) error {
	if demoMode {
		path = storage.InMemory
	}
	if path == "" {
		folder, err := dataFolder()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(folder, 0o755); err != nil {
			return err
		}
		path = filepath.Join(folder, "musa.db")
	}
	database, err := storage.Open(path)
	if err != nil {
		return err
	}
	defer database.Close()
	program, err := app.New(database, app.Options{Dark: options.Dark, DevelopmentTools: tools, OpenLink: window.OpenLink, Now: clock})
	if err != nil {
		return err
	}
	if demoMode {
		if err := demo.Fill(program.Stores()); err != nil {
			return err
		}
	}
	program.Open(place)
	program.Show(page)
	return window.Run(options, program.Build)
}

func wasSet(name string) bool {
	set := false
	flag.Visit(func(each *flag.Flag) { set = set || each.Name == name })
	return set
}
