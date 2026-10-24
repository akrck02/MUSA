package look

import ui "github.com/akrck02/go-ui"

// Measures of a sync status line, in pixels, as in the author's design.
const (
	syncIconSize = 16
	syncTextSize = 15
	syncGap      = 8
)

// Unsynced is what a page synced with an outside service says while it is not connected to one.
const Unsynced = "Sin sincronizar"

// SyncStatus is a soft line that says how a page stands with the outside service it copies
// (Trello, a shop): a circling arrow and words.
func SyncStatus(context *ui.Context, words string) {
	theme := context.Theme
	context.Box(ui.Layout{Gap: syncGap, AlignY: ui.Center}, func() {
		context.IconSized("refresh-cw", syncIconSize, theme.InkSoft)
		context.Text(words, ui.TextStyle{Size: syncTextSize, Color: theme.InkSoft})
	})
}
