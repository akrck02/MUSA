package stand

import (
	"fmt"
	"slices"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/money"
)

// Measures of the question of which version is sold, in pixels.
const (
	versionRowPadding = 16
	versionRowGap     = 10
	versionRowRadius  = 12
	versionNameSize   = 18
	versionPriceSize  = 14
	versionsPillGap   = 4
)

// ByProduct groups items into the versions of each product, keeping the order in which each
// product first appears: the stand shows one tile per product.
func ByProduct(items []Item) [][]Item {
	var groups [][]Item
	where := map[string]int{}
	for _, item := range items {
		index, found := where[item.ProductID]
		if !found {
			where[item.ProductID] = len(groups)
			groups = append(groups, []Item{item})
			continue
		}
		groups[index] = append(groups[index], item)
	}
	return groups
}

// productTile is a product's tile: the version's own tile when there is one, or a tile for all of
// them that asks which one is sold.
func (screens *Screens) productTile(context *ui.Context, versions []Item, fill draw2d.Color) {
	if len(versions) == 1 {
		screens.tile(context, versions[0], fill)
		return
	}
	summary := versions[0]
	summary.VariantID, summary.Version, summary.Left = "", "", 0
	summary.Label = summary.Name
	chosen := 0
	for _, version := range versions {
		summary.Left += version.Left
		summary.Price = min(summary.Price, version.Price)
		chosen += screens.basket.Quantity(version.VariantID)
	}
	theme := context.Theme
	colours := tilePaint(theme, chosen > 0, summary.Left == 0)
	label := fmt.Sprintf("%s · %d versiones", summary.Name, len(versions))
	normal := ui.Look{Fill: colours.surface, Border: colours.border, BorderWidth: tileBorder, Radius: tileRadius, Clip: true}
	pressed := normal
	pressed.Border = theme.ControlLine
	layout := ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: ui.Pad(tileBorder)}
	if context.Clickable(label, layout, ui.Looks{Normal: normal, Hovered: normal, Pressed: pressed}, func() {
		screens.picture(context, summary, chosen, colours.fade(fill))
		screens.groupWords(context, summary, versions, colours)
	}) {
		screens.askVersion(versions)
	}
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: label})
}

// askVersion opens the question of which version is sold, unless all of them are sold out.
func (screens *Screens) askVersion(versions []Item) {
	if !slices.ContainsFunc(versions, func(version Item) bool { return version.Left > 0 }) {
		screens.tell(true, "No quedan más")
		return
	}
	screens.asking = versions
}

// groupWords is the lower half of a product's tile: its name, «desde» its lowest price when the
// versions cost differently, and how many versions it has.
func (screens *Screens) groupWords(context *ui.Context, summary Item, versions []Item, colours paint) {
	price := money.Format(summary.Price)
	if slices.ContainsFunc(versions, func(version Item) bool { return version.Price != summary.Price }) {
		price = "desde " + price
	}
	padding := ui.Padding{Left: wordsPaddingX, Right: wordsPaddingX, Top: wordsPaddingTop, Bottom: wordsPaddingBelow}
	context.Frame(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Padding: padding}, ui.Look{Clip: true}, func() {
		context.Text(summary.Name, ui.TextStyle{Size: nameSize, Weight: draw2d.Bold, Color: colours.ink})
		context.Box(ui.Layout{Height: ui.Fixed(rowSpace)}, nil)
		context.Box(ui.Layout{Width: ui.Grow(), AlignX: ui.SpaceBetween, AlignY: ui.Center}, func() {
			context.Text(price, ui.TextStyle{Size: priceSize, Color: colours.soft})
			// A layered sheet and the count: «3 versiones» in words does not fit beside «desde 10 €».
			pill := ui.Layout{Padding: ui.PadXY(leftPaddingX, leftPaddingY), Gap: versionsPillGap, AlignY: ui.Center}
			context.Frame(pill, ui.Look{Fill: colours.sunken, Radius: roundRadius}, func() {
				context.IconSized("layers", leftSize, colours.quiet)
				context.Text(fmt.Sprint(len(versions)), ui.TextStyle{Size: leftSize, Color: colours.quiet})
			})
		})
	})
}

// drawVersionQuestion asks which version of a product is sold: a row per version with its price
// and what is left. Choosing one puts it in the basket; a sold-out one cannot be chosen.
func (screens *Screens) drawVersionQuestion(context *ui.Context) {
	if len(screens.asking) == 0 {
		return
	}
	versions := screens.asking
	if context.Sheet(versions[0].Name, ui.SheetOptions{}, func() {
		context.Text("¿Qué versión vendes?", ui.TextStyle{Color: context.Theme.InkSoft})
		for _, version := range versions {
			context.Push(version.VariantID)
			if screens.versionRow(context, version) {
				screens.basket.Add(version)
				screens.asking = nil
			}
			context.Pop()
		}
	}) {
		screens.asking = nil
	}
}

// versionRow is one version in the question: its name and price, and what is left.
func (screens *Screens) versionRow(context *ui.Context, version Item) bool {
	theme := context.Theme
	colours := tilePaint(theme, false, version.Left == 0)
	normal := ui.Look{Border: colours.border, BorderWidth: tileBorder, Radius: versionRowRadius}
	hovered := normal
	hovered.Border = theme.ControlLine
	layout := ui.Layout{Width: ui.Grow(), Padding: ui.Pad(versionRowPadding), Gap: versionRowGap, AlignY: ui.Center}
	looks := ui.Looks{Normal: normal, Hovered: hovered, Pressed: hovered}
	if version.Left == 0 {
		looks.Hovered, looks.Pressed = normal, normal
	}
	label := version.Version
	clicked := context.Clickable(label, layout, looks, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow()}, func() {
			context.Text(version.Version, ui.TextStyle{Size: versionNameSize, Weight: draw2d.Bold, Color: colours.ink})
			context.Text(money.Format(version.Price), ui.TextStyle{Size: versionPriceSize, Color: colours.soft})
		})
		if chosen := screens.basket.Quantity(version.VariantID); chosen > 0 {
			screens.versionInBasket(context, version, chosen)
		}
		leftPill(context, version.Left, colours)
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: label})
	return clicked && version.Left > 0
}

// versionInBasket is how many of a version are in the basket, and a «−» that takes one out: the
// product's tile cannot say which version to take out.
func (screens *Screens) versionInBasket(context *ui.Context, version Item, chosen int) {
	theme := context.Theme
	context.Text(fmt.Sprintf("%d %s", chosen, screens.inBasketWords()), ui.TextStyle{Size: leftSize, Color: theme.Primary})
	side := ui.Fixed(minusSide)
	layout := ui.Layout{Width: side, Height: side, AlignX: ui.Center, AlignY: ui.Center}
	round := ui.Look{Fill: theme.Sunken, Radius: minusSide / 2}
	clicked := context.Clickable("Quitar uno", layout, ui.Looks{Normal: round, Hovered: round, Pressed: round}, func() {
		context.Text("−", ui.TextStyle{Size: minusSize, Color: theme.Ink, LineHeight: 1})
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: "Quitar uno"}) // a sign alone says nothing
	if clicked {
		screens.basket.Remove(version.VariantID)
	}
}
