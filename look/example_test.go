package look_test

import (
	"fmt"

	ui "github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"

	"github.com/Itros97/MUSA/look"
)

// A screen is a function drawn every frame: Empty both draws the empty state and says whether its
// action was pressed in this frame. The test window presses it as a person would.
func ExampleEmpty() {
	screen := func(context *ui.Context) {
		if look.Empty(context, "package", "Aún no hay productos", "Crea el primero para empezar a vender.", "Crear producto") {
			fmt.Println("the action was pressed")
		}
	}
	window, err := uitest.New(screen, uitest.Size(800, 600))
	if err != nil {
		fmt.Println(err)
		return
	}
	window.Frame()
	if err := window.Click("Crear producto"); err != nil {
		fmt.Println(err)
	}
	// Output: the action was pressed
}
