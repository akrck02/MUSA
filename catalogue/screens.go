package catalogue

import (
	"time"

	ui "github.com/akrck02/go-ui"
)

// noteTime is how long a note about what an action did stays on screen.
const noteTime = 3 * time.Second

// Screens are the catalogue's pages: the list, and the editor of one product over it. They keep
// what they show between frames and read it again after each change.
type Screens struct {
	store *Store
	newID func() string

	products []Product
	tags     []Tag
	query    Query
	loadErr  error

	kindTab   int  // index in kindTabs
	tagTab    int  // 0 for every tag, then each tag of the kind, then the archived
	searching bool // the search field is open

	choosingKind bool
	stockOf      string // the product whose quick stock sheet is open
	editor       *editor

	note        string
	noteWarning bool
	noteUntil   time.Time
}

// NewScreens draws the catalogue kept in store.
func NewScreens(store *Store, newID func() string) *Screens {
	return &Screens{store: store, newID: newID}
}

// Reload reads the catalogue again.
func (screens *Screens) Reload() {
	screens.products, screens.loadErr = screens.store.All()
	if screens.loadErr == nil {
		screens.tags, screens.loadErr = screens.store.Tags()
	}
}

// Back closes what is open over the list; false when only the list was open.
func (screens *Screens) Back() bool {
	switch {
	case screens.editor != nil:
		screens.editor = nil
	case screens.choosingKind || screens.stockOf != "":
		screens.choosingKind, screens.stockOf = false, ""
	default:
		return false
	}
	return true
}

// Draw draws the open page of the catalogue.
func (screens *Screens) Draw(context *ui.Context) {
	if screens.editor != nil {
		screens.drawEditor(context)
	} else {
		screens.drawList(context)
	}
	if time.Now().Before(screens.noteUntil) {
		context.Toast(screens.note, screens.noteWarning)
		context.RequestFrame(time.Until(screens.noteUntil))
	}
}

// tell shows what an action did, or why it failed, and reads the catalogue again.
func (screens *Screens) tell(err error, done string) {
	screens.note, screens.noteWarning = done, false
	if err != nil {
		screens.note, screens.noteWarning = "No se pudo guardar: "+err.Error(), true
	}
	screens.noteUntil = time.Now().Add(noteTime)
	screens.Reload()
}

func (screens *Screens) product(id string) (Product, bool) {
	for _, product := range screens.products {
		if product.ID == id {
			return product, true
		}
	}
	return Product{}, false
}

// EditFirst opens the editor of the first product of the list, when there is one.
func (screens *Screens) EditFirst() {
	if shown := Filter(screens.products, screens.query); len(shown) > 0 {
		screens.edit(shown[len(shown)/2])
	}
}
