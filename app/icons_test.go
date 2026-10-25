package app

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/opengl/overlay/lucide"
)

// TestEveryIconDrawnExists: an icon's name that Lucide does not have draws nothing, and a button
// with only an icon then shows an empty circle.
func TestEveryIconDrawnExists(test *testing.T) {
	for _, page := range explorePages {
		program, window := openApp(test, true)
		openPage(program, page)
		window.Frames(3)
		for _, command := range window.DrawList().Commands {
			if command.Kind != draw2d.IconKind {
				continue
			}
			if _, found := lucide.Codepoint(command.Text); !found {
				test.Errorf("%s: the icon %q does not exist", page, command.Text)
			}
		}
	}
}
