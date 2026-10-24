package catalogue

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/look"
)

// Measures of the product editor, in pixels, as in the author's design. A border takes no room in
// go-ui, so paddings beside one are a pixel larger than the design's CSS.
const (
	editorTitleSize    = 30
	editorTitleHeight  = 1.2
	editorTopGap       = 14
	editorTopBottom    = 20
	backButtonSide     = 48
	backIconSize       = 22
	headerButtonRadius = 14
	ghostPaddingX      = 20 + 1
	ghostPaddingY      = 14 + 1
	ghostSize          = 17
	savePaddingX       = 28
	savePaddingY       = 14
	saveSize           = 18
	headerButtonGap    = 8
	editorSideWidth    = 400
	editorColumnsGap   = 28
	pictureRadius      = 20
	pictureLetterSize  = 120
	pictureLetterAlpha = 0x99 / 255.0
	photoButtonBottom  = 14
	photoPaddingX      = 18
	photoPaddingY      = 10
	photoSize          = 16
	photoIconSize      = 18
	photoGap           = 8
	fieldTop           = 18
	fieldLabelSize     = 15
	fieldLabelBottom   = 6
	inputRadius        = 12
	inputPaddingX      = 16 + 1 - fieldInset
	inputPaddingY      = 14 + 1
	inputSize          = 18
	fieldInset         = 10 // the room go-ui keeps inside a field, before its text
	kindTop            = 8
	kindGap            = 10
	kindIconSize       = 16
	tagGap             = 8
	tagPaddingX        = 12
	tagPaddingY        = 6
)

func (screens *Screens) drawEditor(context *ui.Context) {
	screens.editorHeader(context)
	if screens.editor == nil {
		return
	}
	side := ui.Fixed(editorSideWidth)
	if look.Narrow(context) {
		side = ui.Grow()
	}
	context.Box(ui.Layout{Direction: look.Across(context), Width: ui.Grow(), Gap: editorColumnsGap, AlignY: ui.Start}, func() {
		context.Box(ui.Layout{Direction: ui.Vertical, Width: side}, func() {
			screens.basics(context)
			screens.tagSection(context)
		})
		context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow()}, func() {
			screens.variantSection(context)
			screens.moreDetails(context)
		})
	})
	// Archiving closes the editor while it draws: what is left of it is not drawn.
	if screens.editor != nil && screens.editor.asking == "delete" {
		screens.askDelete(context)
	}
}

// editorHeader is a round back button and the product's name, with Duplicar and Guardar.
func (screens *Screens) editorHeader(context *ui.Context) {
	theme := context.Theme
	edited := &screens.editor.form
	title := nameOr(edited.product.Name, "Nuevo producto")
	context.Box(ui.Layout{Width: ui.Grow(), Gap: editorTopGap, AlignY: ui.Center, Padding: ui.Padding{Bottom: editorTopBottom}}, func() {
		round := ui.Look{Fill: theme.Sunken, Radius: backButtonSide / 2}
		layout := ui.Layout{Width: ui.Fixed(backButtonSide), Height: ui.Fixed(backButtonSide), AlignX: ui.Center, AlignY: ui.Center}
		back := context.Clickable("Volver", layout, ui.Looks{Normal: round, Hovered: round, Pressed: round}, func() {
			context.IconSized("chevron-left", backIconSize, theme.Ink)
		})
		context.Describe(ui.Description{Role: ui.ButtonRole, Label: "Volver"}) // an icon alone has no name to read
		if back {
			screens.editor = nil
			return
		}
		context.Text(title, ui.TextStyle{Size: editorTitleSize, Weight: draw2d.Bold, LineHeight: editorTitleHeight})
		context.Spacer()
		context.Box(ui.Layout{Gap: headerButtonGap, AlignY: ui.Center}, func() {
			if !edited.isNew && ghostButton(context, "Duplicar") {
				screens.duplicate()
				return
			}
			if saveButton(context, "Guardar") {
				screens.save()
			}
		})
	})
}

// ghostButton is a quiet button: words on the surface with a fine line.
func ghostButton(context *ui.Context, words string) bool {
	theme := context.Theme
	normal := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: headerButtonRadius}
	hovered := normal
	hovered.Fill = theme.Sunken
	return wordsButton(context, words, look.Answering(theme, normal, hovered),
		ui.PadXY(ghostPaddingX, ghostPaddingY), ui.TextStyle{Size: ghostSize, Weight: draw2d.Medium})
}

// saveButton is the editor's main button, on the accent.
func saveButton(context *ui.Context, words string) bool {
	theme := context.Theme
	normal := ui.Look{Fill: theme.Primary, Radius: headerButtonRadius}
	hovered := normal
	hovered.Fill = theme.Primary.Mix(theme.Ink, look.HoverDarkening)
	return wordsButton(context, words, look.Answering(theme, normal, hovered),
		ui.PadXY(savePaddingX, savePaddingY), ui.TextStyle{Size: saveSize, Weight: draw2d.Bold, Color: theme.OnPrimary})
}

// wordsButton is a button of only words, as CSS's button: as tall as its letters.
func wordsButton(context *ui.Context, words string, looks ui.Looks, padding ui.Padding, style ui.TextStyle) bool {
	style.LineHeight = ui.LettersLineHeight
	clicked := context.Clickable(words, ui.Layout{Padding: padding}, looks, func() { context.Text(words, style) })
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: words})
	return clicked
}

// basics is the picture, the name and the kind, which never changes once made.
func (screens *Screens) basics(context *ui.Context) {
	editing := screens.editor
	edited := &editing.form
	theme := context.Theme
	screens.editorPicture(context)
	look.Space(context, fieldTop)
	fieldLabel(context, "Nombre")
	input := ui.Look{Fill: theme.Surface, Border: theme.Line, BorderWidth: 1, Radius: inputRadius}
	context.Frame(ui.Layout{Width: ui.Grow(), Padding: ui.PadXY(inputPaddingX, inputPaddingY)}, input, func() {
		context.TextFieldWith("name", &edited.name, ui.TextFieldOptions{Placeholder: "Nombre del producto", Plain: true,
			Size: inputSize, Height: ui.Fixed(inputSize * theme.LineHeight)})
	})
	if words := editing.problemFor(CodeBlankName, "", ""); words != "" {
		context.Text(words, ui.TextStyle{Size: fieldLabelSize, Color: theme.Danger})
	}
	look.Space(context, kindTop)
	context.Box(ui.Layout{Gap: kindGap, AlignY: ui.Center}, func() {
		context.IconSized("lock", kindIconSize, theme.InkSoft)
		context.Text(KindWords(edited.product.Kind, edited.tags)+" · el tipo no se cambia", ui.TextStyle{Size: fieldLabelSize, Color: theme.InkSoft})
	})
}

// fieldLabel is the soft name over a field.
func fieldLabel(context *ui.Context, words string) {
	context.Box(ui.Layout{Padding: ui.Padding{Bottom: fieldLabelBottom}}, func() {
		context.Text(words, ui.TextStyle{Size: fieldLabelSize, Color: context.Theme.InkSoft})
	})
}

// editorPicture is the product's picture, square, with its first letter until it has a photo, and
// «Cambiar foto» on a pill at its foot.
func (screens *Screens) editorPicture(context *ui.Context) {
	theme := context.Theme
	product := screens.editor.form.product
	name := nameOr(product.Name, screens.editor.form.name)
	fill := ListFill(screens.positions[product.ID])
	if _, stored := screens.positions[product.ID]; !stored {
		fill = PictureFill(name)
	}
	layout := ui.Layout{Direction: ui.Stacked, Width: ui.Grow(), Height: ui.Fixed(editorSideWidth), AlignX: ui.Center, AlignY: ui.Center}
	context.Frame(layout, ui.Look{Fill: fill, Radius: pictureRadius}, func() {
		context.Text(Initial(nameOr(name, "?")), ui.TextStyle{Size: pictureLetterSize, Weight: draw2d.Bold,
			Color: draw2d.Color{1, 1, 1, pictureLetterAlpha}, LineHeight: ui.LettersLineHeight})
		foot := ui.Layout{Width: ui.Grow(), Height: ui.Grow(), AlignX: ui.Center, AlignY: ui.End, Padding: ui.Padding{Bottom: photoButtonBottom}}
		context.Box(foot, func() {
			pill := ui.Look{Fill: theme.Surface, Radius: photoSize * 2}
			button := ui.Layout{Padding: ui.PadXY(photoPaddingX, photoPaddingY), Gap: photoGap, AlignY: ui.Center}
			if context.Clickable("Cambiar foto", button, ui.Looks{Normal: pill, Hovered: pill, Pressed: pill}, func() {
				context.IconSized("image", photoIconSize, theme.Ink)
				context.Text("Cambiar foto", ui.TextStyle{Size: photoSize, Weight: draw2d.Medium, LineHeight: ui.LettersLineHeight})
			}) {
				// TODO(akrck02): choose a photo of the work from the device's pictures (proposal 3, R12).
				screens.tell(nil, "Las fotos llegarán pronto")
			}
		})
	})
}

// tagSection is the product's tags as pills, and «+ Añadir», which opens a field with suggestions.
func (screens *Screens) tagSection(context *ui.Context) {
	editing := screens.editor
	edited := &editing.form
	look.Space(context, fieldTop)
	fieldLabel(context, "Etiquetas")
	context.Box(ui.Layout{Width: ui.Grow(), Gap: tagGap, Wrap: true}, func() {
		for index, name := range edited.tags {
			context.Push(name)
			if tagPill(context, name, "Quitar "+name) {
				edited.tags = append(edited.tags[:index:index], edited.tags[index+1:]...)
			}
			context.Pop()
		}
		if tagPill(context, "+ Añadir", "Añadir etiqueta") {
			editing.addingTag = !editing.addingTag
		}
	})
	if !editing.addingTag {
		return
	}
	look.Space(context, tagGap)
	context.Box(ui.Layout{Width: ui.Grow(), Gap: tagGap, AlignY: ui.Center}, func() {
		context.TextFieldWith("tag", &edited.tagTyped, ui.TextFieldOptions{Placeholder: "Lámina, pegatina, gatos…"})
		if context.IconButton("check", "##Confirmar etiqueta") {
			edited.addTag(edited.tagTyped)
		}
	})
	look.Space(context, tagGap)
	context.Box(ui.Layout{Width: ui.Grow(), Gap: tagGap, Wrap: true}, func() {
		for _, tag := range SuggestTags(screens.tags, edited.tags, edited.tagTyped) {
			context.Push("suggested-" + tag.ID)
			if tagPill(context, "+ "+tag.Name, "Añadir "+tag.Name) {
				edited.addTag(tag.Name)
			}
			context.Pop()
		}
	})
}

// tagPill is a tag on a sunken pill; tapping it does what label says (a tag is removed).
func tagPill(context *ui.Context, words, label string) bool {
	theme := context.Theme
	normal := ui.Look{Fill: theme.Sunken, Radius: fieldLabelSize * 2}
	hovered := normal
	hovered.Fill = theme.Sunken.Mix(theme.Ink, look.HoverDarkening/2)
	clicked := context.Clickable(label, ui.Layout{Padding: ui.PadXY(tagPaddingX, tagPaddingY)}, look.Answering(theme, normal, hovered), func() {
		context.Text(words, ui.TextStyle{Size: fieldLabelSize})
	})
	context.Describe(ui.Description{Role: ui.ButtonRole, Label: label})
	return clicked
}
