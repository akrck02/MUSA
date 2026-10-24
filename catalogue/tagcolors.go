package catalogue

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/look"
)

// tagTones are the soft fill and the deep ink of each tag colour, for the light theme and the
// dark one; each pair reads at 4.5 : 1 or more.
var tagTones = map[TagColor][2]look.Tone{
	Rose:   {{Fill: draw2d.MustHex("#FFD9E2"), Ink: draw2d.MustHex("#5C1130")}, {Fill: draw2d.MustHex("#6B2340"), Ink: draw2d.MustHex("#FFD9E2")}},
	Amber:  {{Fill: draw2d.MustHex("#FFE0B2"), Ink: draw2d.MustHex("#4A2800")}, {Fill: draw2d.MustHex("#5C3A00"), Ink: draw2d.MustHex("#FFE0B2")}},
	Lime:   {{Fill: draw2d.MustHex("#DDEFB0"), Ink: draw2d.MustHex("#243600")}, {Fill: draw2d.MustHex("#35501A"), Ink: draw2d.MustHex("#DDEFB0")}},
	Teal:   {{Fill: draw2d.MustHex("#B9EFE6"), Ink: draw2d.MustHex("#00382F")}, {Fill: draw2d.MustHex("#00504A"), Ink: draw2d.MustHex("#B9EFE6")}},
	Sky:    {{Fill: draw2d.MustHex("#CDE5FF"), Ink: draw2d.MustHex("#00325A")}, {Fill: draw2d.MustHex("#004A7A"), Ink: draw2d.MustHex("#CDE5FF")}},
	Violet: {{Fill: draw2d.MustHex("#E9DDFF"), Ink: draw2d.MustHex("#33196B")}, {Fill: draw2d.MustHex("#4A3380"), Ink: draw2d.MustHex("#E9DDFF")}},
	Slate:  {{Fill: draw2d.MustHex("#DEE3EA"), Ink: draw2d.MustHex("#29313A")}, {Fill: draw2d.MustHex("#3F4852"), Ink: draw2d.MustHex("#DEE3EA")}},
	Coral:  {{Fill: draw2d.MustHex("#FFDAD3"), Ink: draw2d.MustHex("#5A1A0D")}, {Fill: draw2d.MustHex("#73301F"), Ink: draw2d.MustHex("#FFDAD3")}},
}

// TagTone is the tone of a tag colour in the theme being drawn.
func TagTone(theme ui.Theme, color TagColor) look.Tone {
	tones, found := tagTones[color]
	if !found {
		return look.Neutral(theme)
	}
	if look.Dark(theme) {
		return tones[1]
	}
	return tones[0]
}

func chipLook(tone look.Tone) ui.ChipLook { return ui.ChipLook{Fill: tone.Fill, Ink: tone.Ink} }

// pictureFills are the quiet colours of the pictures that stand for a work until it has a photo:
// earthy pastels, softer than the tags', so a wall of them reads calm, as in the author's design.
var pictureFills = map[TagColor]draw2d.Color{
	Rose: draw2d.MustHex("#DDB9C5"), Amber: draw2d.MustHex("#EDDBA8"), Lime: draw2d.MustHex("#D3E0B9"),
	Teal: draw2d.MustHex("#B9CFC2"), Sky: draw2d.MustHex("#B7CCE0"), Violet: draw2d.MustHex("#CBC5E2"),
	Slate: draw2d.MustHex("#D8D4E2"), Coral: draw2d.MustHex("#E8C9B0"),
}

// PictureFill is the colour of the picture of a work called name: the same work always gets the
// same one. Pictures keep their light colours on the dark theme too, as the author's design does.
func PictureFill(name string) draw2d.Color {
	return pictureFills[TagColorFor(TagKey(name))]
}

// listFills are the colours of the pictures in a list, one after another, so neighbours never
// share one: the author's design paints its cards in this order.
var listFills = []draw2d.Color{
	draw2d.MustHex("#E8C9B0"), draw2d.MustHex("#B9CFC4"), draw2d.MustHex("#C9C4E0"), draw2d.MustHex("#EBD9A8"),
	draw2d.MustHex("#D9B8C4"), draw2d.MustHex("#B8CCE0"), draw2d.MustHex("#CFE0B8"), draw2d.MustHex("#E0C2A8"),
	draw2d.MustHex("#C4D4D9"), draw2d.MustHex("#E3B8B0"), draw2d.MustHex("#BFC9B0"), draw2d.MustHex("#D6C8E8"),
}

// ListFill is the colour of the picture at position in a list, on either theme.
func ListFill(position int) draw2d.Color {
	return listFills[position%len(listFills)]
}
