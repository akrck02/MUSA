package projects

import (
	"strconv"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/customers"
	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
)

// Measures of the board, in pixels, as in the author's design. A border takes no room in go-ui,
// so paddings beside one are a pixel larger than the design's CSS.
const (
	boardTop          = 18
	boardGap          = 16
	columnRadius      = 18
	columnPadding     = 14
	columnMinimum     = 560
	columnTitleSize   = 17
	columnTitleBottom = 12
	cardRadius        = 14
	cardPadding       = 14 + 1
	cardBottom        = 12
	stripHeight       = 70
	stripRadius       = 10
	stripBottom       = 10
	cardTitleSize     = 16
	whoSize           = 14
	whoTop            = 4
	whoGap            = 6
	networkSize       = 30
	networkIconSize   = 15
	networkGap        = 4
	linkedTop         = 10
	linkedSize        = 13
	linkedPadX        = 10 + 1
	linkedPadY        = 4 + 1
	footTop           = 10
	footSize          = 14
	duePadX           = 10
	duePadY           = 4
	addSize           = 15
	addIconSize       = 16
	addPadX           = 4
	addPadY           = 8
	addGap            = 8
	deliveredOpacity  = 0.7 // a delivered card fades into its column
	soonDays          = 1   // a deadline this close, or closer, is coloured
)

// drawBoard is one column per stage, each with its projects and «Añadir proyecto».
func (screens *Screens) drawBoard(context *ui.Context) {
	look.Space(context, boardTop)
	context.Box(ui.Layout{Width: ui.Grow(), Gap: boardGap, Columns: look.Columns(context, len(Stages))}, func() {
		for _, stage := range Stages {
			screens.column(context, stage)
		}
	})
}

// column is one stage: its name and count, its projects and a link to add one.
func (screens *Screens) column(context *ui.Context, stage Stage) {
	theme := context.Theme
	var inStage []Project
	for _, project := range screens.projects {
		if project.Stage == stage {
			inStage = append(inStage, project)
		}
	}
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Height: ui.Grow().Between(columnMinimum, 0),
		Padding: ui.Pad(columnPadding)}
	context.Frame(layout, ui.Look{Fill: theme.Sunken, Radius: columnRadius}, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Left: 4, Top: 2, Right: 4, Bottom: columnTitleBottom}}, func() {
			context.Box(ui.Layout{Width: ui.Grow()}, func() {
				context.Text(StageWords(stage), ui.TextStyle{Size: columnTitleSize, Weight: draw2d.Bold})
			})
			context.Text(strconv.Itoa(len(inStage)), ui.TextStyle{Size: columnTitleSize, Weight: draw2d.Medium, Color: theme.InkQuiet})
		})
		for _, project := range inStage {
			context.Push(project.ID)
			screens.card(context, project)
			context.Pop()
		}
		context.Push(string(stage))
		if addLink(context) {
			screens.startProject(stage)
		}
		context.Pop()
	})
}

// addLink is «Añadir proyecto» under a column's cards.
func addLink(context *ui.Context) bool {
	theme := context.Theme
	layout := ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(addPadX, addPadY), Gap: addGap, AlignY: ui.Center}
	clicked := context.Clickable("Añadir proyecto", layout, ui.Looks{}, func() {
		context.IconSized("plus", addIconSize, theme.InkSoft)
		context.Text("Añadir proyecto", ui.TextStyle{Size: addSize, Color: theme.InkSoft})
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: "Añadir proyecto"})
	return clicked
}

// palette is the colours of a card: a delivered one is faded into its column.
type palette struct {
	surface, line, ink, soft, sunken draw2d.Color
	fade                             func(draw2d.Color) draw2d.Color
}

func paletteOf(theme ui.Theme, delivered bool) palette {
	fade := func(colour draw2d.Color) draw2d.Color { return colour }
	if delivered {
		fade = func(colour draw2d.Color) draw2d.Color { return colour.Mix(theme.Sunken, 1-deliveredOpacity) }
	}
	return palette{surface: fade(theme.Surface), line: fade(theme.Line), ink: fade(theme.Ink), soft: fade(theme.InkSoft),
		sunken: fade(theme.Sunken), fade: fade}
}

// card is one project: a strip of its colour, the title, who it is for with their networks, the
// product it makes, when it is due and its price. Tapping it opens the editor.
func (screens *Screens) card(context *ui.Context, project Project) {
	colours := paletteOf(context.Theme, project.Stage == Delivered)
	normal := ui.Look{Fill: colours.surface, Border: colours.line, BorderWidth: 1, Radius: cardRadius}
	hovered := normal
	hovered.Border = colours.fade(context.Theme.ControlLine)
	context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.Padding{Bottom: cardBottom}}, func() {
		layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(cardPadding)}
		if context.Clickable(project.Title, layout, ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}, func() {
			strip := ui.Look{Fill: colours.fade(catalogue.ListFill(context.Theme, screens.created(project))), Radius: stripRadius}
			context.Frame(ui.Layout{Width: ui.Grow(), Height: ui.Fixed(stripHeight)}, strip, nil)
			look.Space(context, stripBottom)
			look.Paragraph(context, project.Title, ui.TextStyle{Size: cardTitleSize, Weight: draw2d.Bold, Color: colours.ink})
			look.Space(context, whoTop)
			screens.whoFor(context, screens.people[project.CustomerID], colours)
			look.Space(context, linkedTop)
			linkedPill(context, screens.productLabel(project.VariantID), colours)
			look.Space(context, footTop)
			context.Box(ui.Layout{Width: ui.Grow(), AlignY: ui.Center}, func() {
				screens.duePill(context, project, colours)
				context.Spacer()
				context.Text(money.Format(project.Price), ui.TextStyle{Size: cardTitleSize, Weight: draw2d.Bold, Color: colours.ink})
			})
		}) {
			screens.edit(project)
		}
	})
}

// created is a project's place among all of them by when they were made, which picks its colour.
func (screens *Screens) created(project Project) int {
	position := 0
	for _, other := range screens.projects {
		if other.CreatedAt.Before(project.CreatedAt) || (other.CreatedAt.Equal(project.CreatedAt) && other.sequence < project.sequence) {
			position++
		}
	}
	return position
}

// whoFor is the customer's name and a round button for each network they are on, in the order
// their profiles were added.
func (screens *Screens) whoFor(context *ui.Context, person customers.Customer, colours palette) {
	context.Box(ui.Layout{Width: ui.Grow(), Gap: whoGap, AlignY: ui.Center, Wrap: true}, func() {
		context.IconSized("user", whoSize, colours.soft)
		context.Text(person.Name, ui.TextStyle{Size: whoSize, Color: colours.soft})
		context.Box(ui.Layout{Gap: networkGap, Padding: ui.Padding{Left: whoGap}}, func() {
			for _, profile := range person.Profiles {
				network, found := screens.network(profile.NetworkID)
				if found && networkButton(context, network, profile.Handle, colours) {
					screens.open(network.Link(profile.Handle))
				}
			}
		})
	})
}

// network finds a network by its id.
func (screens *Screens) network(id string) (customers.Network, bool) {
	for _, network := range screens.networks {
		if network.ID == id {
			return network, true
		}
	}
	return customers.Network{}, false
}

// networkButton is a network's icon on a small sunken circle.
func networkButton(context *ui.Context, network customers.Network, handle string, colours palette) bool {
	label := network.Name + " " + handle
	side := ui.Fixed(networkSize)
	layout := ui.Layout{Width: side, Height: side, AlignX: ui.Center, AlignY: ui.Center}
	normal := ui.Look{Fill: colours.sunken, Radius: networkSize / 2}
	clicked := context.Clickable(label, layout, ui.Looks{Normal: normal, Hovered: normal, Pressed: normal}, func() {
		context.IconSized(network.Icon, networkIconSize, colours.ink)
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: label})
	return clicked
}

// linkedPill is the product a project makes on an outlined pill, or a dashed-looking quiet one
// asking to link one.
func linkedPill(context *ui.Context, label string, colours palette) {
	icon, ink := "cloud", colours.ink
	if label == "" {
		icon, label, ink = "plus", "Asociar producto", colours.soft
	}
	looks := ui.Look{Border: colours.line, BorderWidth: 1, Radius: linkedSize}
	context.Frame(ui.Layout{Padding: ui.PadXY(linkedPadX, linkedPadY), Gap: whoGap, AlignY: ui.Center}, looks, func() {
		context.IconSized(icon, linkedSize, ink)
		context.Text(label, ui.TextStyle{Size: linkedSize, Color: ink})
	})
}

// productLabel is the name of the product a project makes, without its version.
func (screens *Screens) productLabel(variantID string) string {
	for _, choice := range screens.choices {
		if choice.variantID == variantID {
			return choice.product
		}
	}
	return ""
}

// duePill says when the project is due, red when late and amber when due today or tomorrow; a
// delivered one says the day it was delivered.
func (screens *Screens) duePill(context *ui.Context, project Project, colours palette) {
	theme := context.Theme
	today := calendar.Of(screens.now())
	words, ink := DeadlineWords(project.Deadline, today), colours.soft
	switch {
	case project.Stage == Delivered:
		words = DeliveredWords(project.DeliveredAt)
	case project.Late(today):
		ink = theme.Danger
	case !project.Deadline.Zero() && project.Deadline.AddDays(-soonDays).Before(today.AddDays(1)):
		ink = theme.Warning
	}
	if words == "" {
		return
	}
	context.Frame(ui.Layout{Padding: ui.PadXY(duePadX, duePadY), Gap: whoGap, AlignY: ui.Center},
		ui.Look{Fill: colours.sunken, Radius: footSize}, func() {
			context.IconSized("calendar", footSize, ink)
			context.Text(words, ui.TextStyle{Size: footSize, Color: ink})
		})
}

// open opens a profile; where links cannot be opened, it says so.
func (screens *Screens) open(link string) {
	if screens.openLink == nil || !screens.openLink(link) {
		screens.tell(true, "No se pudo abrir "+link)
	}
}
