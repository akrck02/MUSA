package catalogue

import (
	"time"

	"github.com/Itros97/MUSA/look"
	"github.com/akrck02/go-renderer/draw2d"
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
	// positions are where each product, by its ID, stands in the whole catalogue: its colour
	positions map[string]int
	// unitsSold counts the units of each version, by its ID, sold this month; nil when unknown
	unitsSold     func() (map[string]int, error)
	soldThisMonth map[string]int
	tags          []Tag
	query         Query
	loadErr       error

	kindTab   int                               // index in kindTabs
	tagTab    int                               // 0 for every tag, then each tag of the kind, then the archived
	searching bool                              // the search field is open
	shapes    map[draw2d.Color]*look.TopRounded // the pictures' fills, kept so drawing them allocates nothing

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

// CountSalesWith makes the list say how many times each digital product sold this month, from
// count: the units of each version, by its ID, sold this month.
func (screens *Screens) CountSalesWith(count func() (map[string]int, error)) {
	screens.unitsSold = count
}

// Reload reads the catalogue again.
func (screens *Screens) Reload() {
	screens.products, screens.loadErr = screens.store.All()
	if screens.loadErr == nil {
		screens.tags, screens.loadErr = screens.store.Tags()
	}
	if screens.loadErr == nil && screens.unitsSold != nil {
		screens.soldThisMonth, screens.loadErr = screens.unitsSold()
	}
	screens.positions = make(map[string]int, len(screens.products))
	for position, product := range screens.products {
		screens.positions[product.ID] = position
	}
}

// soldOf is how many units of product sold this month.
func (screens *Screens) soldOf(product Product) int {
	units := 0
	for _, variant := range product.Variants {
		units += screens.soldThisMonth[variant.ID]
	}
	return units
}

// ShowDigital opens the list on the digital products.
func (screens *Screens) ShowDigital() {
	screens.editor = nil
	screens.kindTab, screens.tagTab = 1, 0
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

// ShowRestock shows only the physical products that need restocking.
func (screens *Screens) ShowRestock() {
	screens.editor = nil
	screens.kindTab, screens.tagTab = 0, 0
	screens.query.LowStockOnly = true
}

// EditFirstWith opens the editor of the first product shown with this many versions, or of the
// first one when none has as many, for captures.
func (screens *Screens) EditFirstWith(versions int) {
	shown := Filter(screens.products, screens.query)
	for _, product := range shown {
		if len(product.Variants) == versions {
			screens.edit(product)
			return
		}
	}
	if len(shown) > 0 {
		screens.edit(shown[0])
	}
}

// Editing tells whether a product's editor is open.
func (screens *Screens) Editing() bool { return screens.editor != nil }
