package catalogue

import (
	"fmt"

	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/look"
)

// Measures of the sheet of tags, in pixels.
const (
	tagBarGap      = 12
	tagsSheetWidth = 640 // room for a row of eight icons
	tagRowPadding  = 10
	tagRowGap      = 14
	tagTileSide    = 44
	tagTileRadius  = 10
	iconChoiceSide = 44
	iconChoiceGap  = 8
	iconColumns    = 8
	chosenBorder   = 2
)

// tagsSheet lists every tag with its icon and how many products carry it; «Cambiar icono» opens
// a choice of icons under the tag.
func (screens *Screens) tagsSheet(context *ui.Context) {
	theme := context.Theme
	if context.Sheet("Etiquetas", ui.SheetOptions{Width: ui.Fixed(tagsSheetWidth)}, func() {
		look.Paragraph(context, "Elige el icono con el que se ve cada etiqueta en el puesto y en tus productos.", ui.TextStyle{Color: theme.InkSoft})
		for index, tag := range screens.tags {
			context.Push(tag.ID)
			if index > 0 {
				context.Separator()
			}
			screens.tagRow(context, tag)
			if screens.iconsOf == tag.ID {
				screens.iconChoices(context, tag)
			}
			context.Pop()
		}
		context.Box(ui.Layout{Width: ui.Grow(), AlignX: ui.End}, func() {
			if context.PrimaryButton("Listo") {
				screens.closeTagsSheet()
			}
		})
	}) {
		screens.closeTagsSheet()
	}
}

func (screens *Screens) closeTagsSheet() { screens.editingTags, screens.iconsOf = false, "" }

// tagRow is a tag's icon, its name and how many products carry it, and «Cambiar icono».
func (screens *Screens) tagRow(context *ui.Context, tag Tag) {
	theme := context.Theme
	context.Box(ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(0, tagRowPadding), Gap: tagRowGap, AlignY: ui.Center}, func() {
		side := ui.Fixed(tagTileSide)
		context.Frame(ui.Layout{Width: side, Height: side, AlignX: ui.Center, AlignY: ui.Center}, ui.Look{Fill: theme.Sunken, Radius: tagTileRadius}, func() {
			context.Icon(tag.Icon())
		})
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow()}, func() {
			context.Text(Plural(tag.Name), ui.TextStyle{Weight: draw2d.Bold})
			context.Text(productsWords(screens.countAllWithTag(tag.ID)), ui.TextStyle{Size: theme.SmallTextSize, Color: theme.InkSoft})
		})
		icon := "chevron-right"
		if screens.iconsOf == tag.ID {
			icon = "chevron-down"
		}
		clicked := context.FlatIconButton(icon, "Cambiar icono##"+tag.ID)
		// Every row says «Cambiar icono»; a screen reader needs to hear whose.
		context.Describe(ui.Description{Role: ui.ButtonRole, Label: "Cambiar el icono de " + Plural(tag.Name)})
		if clicked {
			screens.toggleIcons(tag.ID)
		}
	})
}

func (screens *Screens) toggleIcons(tagID string) {
	if screens.iconsOf == tagID {
		screens.iconsOf = ""
		return
	}
	screens.iconsOf = tagID
}

func (screens *Screens) countAllWithTag(id string) int {
	return len(Filter(screens.products, Query{TagIDs: []string{id}}))
}

func productsWords(count int) string {
	if count == 1 {
		return "1 producto"
	}
	return fmt.Sprintf("%d productos", count)
}

// iconChoices is the grid of icons a tag can take, its own marked, and «Volver al icono
// automático» when the artist chose one.
func (screens *Screens) iconChoices(context *ui.Context, tag Tag) {
	theme := context.Theme
	context.Box(ui.Layout{Width: ui.Grow(), Columns: iconColumns, Gap: iconChoiceGap}, func() {
		for _, icon := range TagIconChoices {
			normal := ui.Look{Border: theme.Line, BorderWidth: 1, Radius: tagTileRadius}
			if icon == tag.Icon() {
				normal = ui.Look{Fill: theme.PrimaryTint, Border: theme.Primary, BorderWidth: chosenBorder, Radius: tagTileRadius}
			}
			hovered := normal
			hovered.Border = theme.ControlLine
			side := ui.Fixed(iconChoiceSide)
			layout := ui.Layout{Width: ui.Grow(), Height: side, AlignX: ui.Center, AlignY: ui.Center}
			label := "Icono " + icon
			clicked := context.Clickable(label, layout, look.Answering(theme, normal, hovered), func() {
				context.Icon(icon)
			})
			context.Describe(ui.Description{Role: ui.ButtonRole, Label: label})
			if clicked {
				screens.tell(screens.store.SetTagIcon(tag.ID, icon), "Icono cambiado")
			}
		}
	})
	if tag.ChosenIcon != "" && context.FlatIconButton("rotate-ccw", "Volver al icono automático") {
		screens.tell(screens.store.SetTagIcon(tag.ID, ""), "Icono automático")
	}
}
