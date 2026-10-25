package app

import (
	"fmt"
	"strings"
	"testing"

	"github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"
)

// explorePages are the pages a person can reach, each opened from a fresh MUSA with the demo's data.
var explorePages = []string{"home", "calendar", "catalogue", "digital", "product", "versions", "fairs", "fair",
	"orders", "order", "history", "stand-choice", "stand", "stand-basket", "fair-stand", "fair-pay", "fair-cash",
	"fair-short", "fair-card", "fair-paid", "projects", "project", "networks"}

// openPage opens one of explorePages.
func openPage(program *App, page string) {
	places := map[string]Place{"home": Home, "catalogue": Catalogue, "fairs": Fairs, "orders": Sales, "stand-choice": Stand, "projects": Projects}
	if place, found := places[page]; found {
		program.Open(place)
		return
	}
	program.Show(page)
}

// pressableLabels are the names of what can be pressed on the page, in reading order.
func pressableLabels(window *uitest.Window) []ui.AccessibleNode {
	var pressable []ui.AccessibleNode
	for _, node := range window.Context().ScreenReaderNodes() {
		if node.Pressable {
			pressable = append(pressable, node)
		}
	}
	return pressable
}

// troubleWords are words MUSA only shows when something went wrong underneath.
var troubleWords = []string{"no se pudo", "storage", "sql", "error", "panic", "not found"}

func TestEveryControlOnEveryPageCanBePressed(test *testing.T) {
	if testing.Short() {
		test.Skip("presses every control of every page, one fresh MUSA each")
	}
	for _, page := range explorePages {
		program, window := openApp(test, true)
		openPage(program, page)
		window.Frames(3)
		controls := pressableLabels(window)
		if len(controls) == 0 {
			test.Errorf("%s: nothing can be pressed", page)
		}
		for index := range controls {
			test.Run(fmt.Sprintf("%s/%d", page, index), func(test *testing.T) { pressAndCheck(test, page, index) })
		}
	}
}

// pressAndCheck opens a page in a fresh MUSA, presses its control at index, and fails when that
// breaks or shows trouble.
func pressAndCheck(test *testing.T, page string, index int) {
	program, window := openApp(test, true)
	openPage(program, page)
	window.Frames(3)
	controls := pressableLabels(window)
	if index >= len(controls) {
		test.Skipf("the page changed: %d controls now", len(controls))
	}
	control := controls[index]
	defer func() {
		if recovered := recover(); recovered != nil {
			test.Fatalf("pressing %q on %s panicked: %v", control.Label, page, recovered)
		}
	}()
	window.Context().ScreenReaderAct(control.ID, ui.ReaderPress)
	window.Frames(4)
	for _, text := range window.Texts() {
		lowered := strings.ToLower(text)
		for _, word := range troubleWords {
			if strings.Contains(lowered, word) {
				test.Errorf("pressing %q on %s shows %q", control.Label, page, text)
			}
		}
	}
}
