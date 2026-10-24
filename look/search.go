package look

import ui "github.com/akrck02/go-ui"

// Search is a search a page opens from a button: whether its field shows, and what is typed in it.
// A page keeps one, draws SearchField where the field goes and SearchButton where the button goes.
type Search struct {
	Open bool
	Text string
	// opening is set on the frame the button opens the search, so the field takes the keyboard
	// the next time it is drawn: the artist types at once, without tapping the field.
	opening bool
}

// searchWidth is how wide the search field opens, in pixels, as in the author's design.
const searchWidth = 320

// searchFieldName names the field, so the button can give it the keyboard.
const searchFieldName = "search"

// SearchField is the search's field while the search is open; nothing while it is closed.
func SearchField(context *ui.Context, search *Search, placeholder string) {
	if !search.Open {
		return
	}
	context.Box(ui.Layout{Width: ui.Fixed(searchWidth)}, func() {
		if search.opening {
			context.FocusField(searchFieldName)
			search.opening = false
		}
		context.TextFieldWith(searchFieldName, &search.Text, ui.TextFieldOptions{Placeholder: placeholder})
	})
}

// SearchButton is the round button that opens the search, or closes it and forgets what was typed.
func SearchButton(context *ui.Context, search *Search) {
	if !ToolButton(context, "search", "Buscar") {
		return
	}
	search.Open, search.opening, search.Text = !search.Open, !search.Open, ""
}
