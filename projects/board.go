package projects

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/customers"
	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
)

// Measures of the board, in pixels.
const (
	bandHeight  = 56
	networkSize = 32
	soonDays    = 1 // a deadline this close, or closer, is coloured
)

// drawBoard is one column per stage, each with its projects and «Añadir proyecto».
func (screens *Screens) drawBoard(context *ui.Context) {
	theme := context.Theme
	context.Grid(look.Columns(context, len(Stages)), func() {
		for _, stage := range Stages {
			var inStage []Project
			for _, project := range screens.projects {
				if project.Stage == stage {
					inStage = append(inStage, project)
				}
			}
			column := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(12), Gap: 10}
			context.Frame(column, ui.Look{Fill: theme.Sunken.WithAlpha(0.6), Radius: theme.PanelRadius}, func() {
				context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(4, 2)}, func() {
					context.Box(ui.Layout{Width: ui.Grow()}, func() { context.Text(StageWords(stage), ui.TextStyle{Weight: draw2d.Bold}) })
					context.Text(fmt.Sprint(len(inStage)), ui.TextStyle{Color: theme.InkSoft})
				})
				for _, project := range inStage {
					context.Push(project.ID)
					screens.card(context, project)
					context.Pop()
				}
				context.Push(string(stage))
				if stage != Delivered && context.FlatIconButton("plus", "Añadir proyecto") {
					screens.startProject(stage)
				}
				context.Pop()
			})
		}
	})
}

// card is one project: a band of its colour, the title, who it is for with their networks, the
// product it makes, when it is due and its price. Tapping it opens the editor.
func (screens *Screens) card(context *ui.Context, project Project) {
	theme := context.Theme
	person := screens.people[project.CustomerID]
	ink := theme.Ink
	if project.Stage == Delivered {
		ink = theme.InkSoft
	}
	if look.ClickableCard(context, project.Title, ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(12), Gap: 8}, func() {
		tone := catalogue.TagTone(theme, catalogue.TagColorFor(project.Title))
		context.Frame(ui.Layout{Width: ui.Grow(), Height: ui.Fixed(bandHeight)}, ui.Look{Fill: tone.Fill, Radius: theme.ControlRadius}, func() {})
		look.Paragraph(context, project.Title, ui.TextStyle{Weight: draw2d.Bold, Color: ink})
		screens.whoFor(context, person)
		if label := screens.productLabel(project.VariantID); label != "" {
			look.Pill(context, label, look.Neutral(theme))
		}
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.Center}, func() {
			screens.deadlinePill(context, project)
			context.Spacer()
			context.Text(money.Format(project.Price), ui.TextStyle{Weight: draw2d.Bold, Color: ink})
		})
	}) {
		screens.edit(project)
	}
}

// whoFor is the customer's name and a button for each network they are on.
func (screens *Screens) whoFor(context *ui.Context, person customers.Customer) {
	theme := context.Theme
	context.Box(ui.Layout{Width: ui.Grow(), Gap: 6, AlignY: ui.Center}, func() {
		context.IconSized("user", theme.SmallTextSize, theme.InkSoft)
		context.Text(person.Name, ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
		for _, network := range screens.networks {
			handle, found := person.Handle(network.ID)
			if !found {
				continue
			}
			if context.FlatIconButton(network.Icon, "##"+network.Name+" "+handle) {
				screens.open(network.Link(handle))
			}
		}
	})
}

// open opens a profile; where links cannot be opened, it says so.
func (screens *Screens) open(link string) {
	if screens.openLink == nil || !screens.openLink(link) {
		screens.tell(true, "No se pudo abrir "+link)
	}
}

func (screens *Screens) productLabel(variantID string) string {
	for _, choice := range screens.choices {
		if choice.variantID == variantID {
			return choice.label
		}
	}
	return ""
}

// deadlinePill says when the project is due: red when late, amber when due today or tomorrow.
func (screens *Screens) deadlinePill(context *ui.Context, project Project) {
	theme := context.Theme
	today := calendar.Of(screens.now())
	if project.Stage == Delivered {
		look.Pill(context, "Entregado", look.Positive(theme))
		return
	}
	words := DeadlineWords(project.Deadline, today)
	if words == "" {
		return
	}
	tone := look.Neutral(theme)
	switch {
	case project.Late(today):
		tone = look.Danger(theme)
	case project.Deadline.AddDays(-soonDays).Before(today.AddDays(1)):
		tone = look.Warning(theme)
	}
	look.Pill(context, words, tone)
}
