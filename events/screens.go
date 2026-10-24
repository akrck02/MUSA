package events

import (
	"slices"
	"strings"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/look"
)

// noteTime is how long a note about what an action did stays on screen.
const noteTime = 3 * time.Second

// Screens are the fairs' pages: the list, one fair's detail, and the sheets over them.
type Screens struct {
	store    *Store
	products *catalogue.Store
	newID    func() string
	now      func() time.Time

	events  []Event
	detail  *Detail
	loadErr error
	editor  *eventEditor
	sheet   sheet

	newExpense expenseForm
	detailTab  int

	// lastSale is the quick sale that can still be undone, what it says and until when.
	lastSale  string
	lastWords string
	lastUntil time.Time

	note        string
	noteWarning bool
	noteUntil   time.Time
}

// NewScreens draws the fairs kept in store, taking stock from products.
func NewScreens(store *Store, products *catalogue.Store, newID func() string, now func() time.Time) *Screens {
	return &Screens{store: store, products: products, newID: newID, now: now}
}

// Reload reads the fairs again, and the open fair's detail.
func (screens *Screens) Reload() {
	screens.events, screens.loadErr = screens.store.All()
	if screens.detail == nil {
		return
	}
	detail, err := screens.store.Detail(screens.detail.Event.ID)
	if err != nil {
		screens.detail = nil
		return
	}
	screens.detail = &detail
}

// Back closes what is open; false when only the list was open.
func (screens *Screens) Back() bool {
	switch {
	case screens.editor != nil:
		screens.editor = nil
	case screens.sheet.kind != noSheet:
		screens.sheet = sheet{}
	case screens.detail != nil:
		screens.detail = nil
		screens.Reload()
	default:
		return false
	}
	return true
}

// Draw draws the open page of the fairs.
func (screens *Screens) Draw(context *ui.Context) {
	if screens.detail != nil {
		screens.drawDetail(context)
	} else {
		screens.drawList(context)
	}
	if screens.editor != nil {
		screens.drawEditor(context)
	}
	if time.Now().Before(screens.noteUntil) {
		context.Toast(screens.note, screens.noteWarning)
		context.RequestFrame(time.Until(screens.noteUntil))
	}
}

// tell shows the first refusal or the error, or done when the action went through, and reads
// the fairs again; it returns whether the action went through.
func (screens *Screens) tell(refused []Refusal, err error, done string) bool {
	screens.note, screens.noteWarning = done, false
	switch {
	case err != nil:
		screens.note, screens.noteWarning = err.Error(), true
	case len(refused) > 0:
		screens.note, screens.noteWarning = RefusalWords(refused[0]), true
	}
	screens.noteUntil = time.Now().Add(noteTime)
	screens.Reload()
	return !screens.noteWarning
}

// Spaces of the list of events, in pixels.
const (
	listTop          = 28 // from the page's header to the first group
	groupTop         = 24 // above each group after the first
	groupLabelBottom = 10 // from a group's name to its first event
	cardGap          = 10 // between two events
)

// drawList is every event, grouped into those happening now, those to come and those past.
func (screens *Screens) drawList(context *ui.Context) {
	look.PageHeader(context, "Eventos", "Ferias, mercados y demás: qué llevas, cuánto vendes y cuánto te cuesta ir", func() {
		if look.MainButton(context, "plus", "Nuevo evento") {
			today := calendar.Of(screens.now())
			screens.StartFair(today)
		}
	})
	switch {
	case screens.loadErr != nil:
		look.Empty(context, "triangle-alert", "No se pudieron leer los eventos", screens.loadErr.Error(), "")
	case len(screens.events) == 0:
		look.Empty(context, "calendar-days", "Aún no hay eventos",
			"Apunta tu próximo evento para llevar la cuenta de lo que llevas y lo que vendes.", "")
	}
	today := calendar.Of(screens.now())
	look.Space(context, listTop-groupTop)
	for _, group := range []struct {
		timing Timing
		title  string
	}{{Happening, "En curso"}, {Upcoming, "Próximos"}, {Past, "Pasados"}} {
		var shown []Event
		for _, event := range screens.events {
			if TimingOf(event, today) == group.timing {
				shown = append(shown, event)
			}
		}
		if len(shown) == 0 {
			continue
		}
		if group.timing == Upcoming { // the next one first
			slices.SortStableFunc(shown, func(one, other Event) int { return strings.Compare(one.StartDate.String(), other.StartDate.String()) })
		}
		look.Space(context, groupTop)
		look.Label(context, group.title)
		look.Space(context, groupLabelBottom)
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: cardGap}, func() {
			for _, event := range shown {
				context.Push(event.ID)
				screens.eventCard(context, event, today)
				context.Pop()
			}
		})
	}
}

func (screens *Screens) eventCard(context *ui.Context, event Event, today calendar.Date) {
	theme := context.Theme
	layout := ui.Layout{Width: ui.Grow(), Padding: ui.Pad(14), Gap: 12, AlignY: ui.Center}
	if look.ClickableCard(context, event.Name, layout, func() {
		colour := theme.Primary
		if TimingOf(event, today) == Past {
			colour = theme.InkQuiet
		}
		context.IconSized("calendar-days", theme.IconSize*1.3, colour)
		context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 2}, ui.Look{Clip: true}, func() {
			context.Text(event.Name, ui.TextStyle{Weight: draw2d.Bold})
			context.Text(event.WhenAndWhere(), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
		})
		tone := look.Neutral(theme)
		if TimingOf(event, today) == Happening {
			tone = look.Positive(theme)
		}
		look.Pill(context, WhenWords(event, today), tone)
	}) {
		screens.open(event.ID)
	}
}

func statusPill(context *ui.Context, status Status) {
	tone := look.Positive(context.Theme)
	if status == Closed {
		tone = look.Neutral(context.Theme)
	}
	look.Pill(context, statusWords[status], tone)
}

func (screens *Screens) open(id string) {
	detail, err := screens.store.Detail(id)
	if err != nil {
		screens.tell(nil, err, "")
		return
	}
	screens.detail, screens.detailTab, screens.lastSale = &detail, 0, ""
}

// OpenFair opens one fair's detail on its stand.
func (screens *Screens) OpenFair(id string) {
	screens.Reload()
	screens.open(id)
}

// eventEditor is the fair editor's text, as typed.
type eventEditor struct {
	draft      Draft
	start, end string
	problems   map[string]string
}

func (screens *Screens) drawEditor(context *ui.Context) {
	editing := screens.editor
	title := "Nuevo evento"
	if editing.draft.Name != "" && screens.detail != nil {
		title = "Editar evento"
	}
	if context.Sheet(title, ui.SheetOptions{}, func() {
		look.Field(context, "Nombre", "name", &editing.draft.Name, "Feria del cómic", editing.problems["name"])
		look.Field(context, "Lugar", "location", &editing.draft.Location, "Pabellón 3", "")
		context.Box(ui.Layout{Width: ui.Grow(), Columns: look.Columns(context, 2), Gap: context.Theme.Gap}, func() {
			look.Field(context, "Empieza", "start", &editing.start, "18/4/2026", editing.problems["start"])
			look.Field(context, "Termina", "end", &editing.end, "19/4/2026", editing.problems["end"])
		})
		context.Box(ui.Layout{Width: ui.Grow(), AlignX: ui.End}, func() {
			if context.PrimaryButton("Guardar") {
				screens.saveEvent()
			}
		})
	}) {
		screens.editor = nil
	}
}

func (screens *Screens) saveEvent() {
	editing := screens.editor
	editing.problems = map[string]string{}
	var startRead, endRead bool
	editing.draft.StartDate, startRead = calendar.ParseTyped(editing.start)
	editing.draft.EndDate, endRead = calendar.ParseTyped(editing.end)
	if !startRead {
		editing.problems["start"] = "Escribe el día como 18/4/2026"
	}
	if !endRead {
		editing.problems["end"] = "Escribe el día como 19/4/2026"
	}
	if len(editing.problems) > 0 {
		return
	}
	refused, err := screens.store.Save(editing.draft)
	for _, refusal := range refused {
		field := map[Code]string{CodeBlankName: "name", CodeEndBeforeStart: "end"}[refusal.Code]
		editing.problems[field] = RefusalWords(refusal)
	}
	if err != nil || len(refused) > 0 {
		return
	}
	id := editing.draft.ID
	screens.editor = nil
	screens.tell(nil, nil, "Evento guardado")
	screens.open(id)
}

// OpenFirstOpen opens the detail of the first fair still open that has sales, or of the first one.
func (screens *Screens) OpenFirstOpen() {
	for _, event := range screens.events {
		if detail, err := screens.store.Detail(event.ID); err == nil && event.Status == Open && len(detail.Sales) > 0 {
			screens.detail = &detail
			return
		}
	}
	if len(screens.events) > 0 {
		screens.open(screens.events[0].ID)
	}
}

// Detail is the fair whose detail is open, or the zero Detail when none is.
func (screens *Screens) Detail() Detail {
	if screens.detail == nil {
		return Detail{}
	}
	return *screens.detail
}

// StartFair opens the editor of a new fair on day.
func (screens *Screens) StartFair(day calendar.Date) {
	screens.editor = &eventEditor{draft: Draft{ID: screens.newID()}, start: calendar.Typed(day), end: calendar.Typed(day)}
}
