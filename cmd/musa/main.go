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
	place := flag.Int("place", 0, "the place to open on: 0 home, 1 catalogue, 2 fairs, 3 orders, 4 stand, 5 projects (for captures)")
	page := flag.String("show", "", "a page inside the place to open on: product, fair, order or stand (for captures)")
	tools := flag.Bool("tools", true, "show the development tools on the home page")
	flag.Parse()
	if !wasSet("dark") {
		options.Dark = window.SystemPrefersDark()
	}

	if err := run(options, *demoMode, *path, app.Place(*place), *page, *tools); err != nil {
		fmt.Fprintln(os.Stderr, "musa:", err)
		os.Exit(1)
	}
}

func run(options window.Options, demoMode bool, path string, place app.Place, page string, tools bool) error {
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
	program, err := app.New(database, app.Options{Dark: options.Dark, DevelopmentTools: tools, OpenLink: window.OpenLink})
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
