package projects

import (
	"strconv"

	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/customers"
	"github.com/Itros97/MUSA/look"
)

// networkSettings is the sheet where the artist chooses the social networks a customer can be
// found on: each with its name and the link to a profile, where {user} stands for the handle.
type networkSettings struct {
	typed   []customers.Network // as typed, one per network, the last one new
	problem string
}

// ShowNetworks opens the networks sheet, for captures and tests.
func (screens *Screens) ShowNetworks() { screens.openNetworks() }

// openNetworks opens the sheet with the networks as they are, and an empty one to add.
func (screens *Screens) openNetworks() {
	typed := append([]customers.Network(nil), screens.networks...)
	screens.settings = &networkSettings{typed: append(typed, customers.Network{})}
}

func (screens *Screens) drawNetworks(context *ui.Context) {
	settings := screens.settings
	theme := context.Theme
	if context.Sheet("Redes sociales", ui.SheetOptions{}, func() {
		look.Paragraph(context, "Dónde encuentras a tus clientes. En el enlace, "+customers.UserMark+" es su nombre de usuario.",
			ui.TextStyle{Color: theme.InkSoft})
		for index := range settings.typed {
			network := &settings.typed[index]
			context.Push(strconv.Itoa(index))
			screens.networkRow(context, network)
			context.Pop()
		}
		if settings.problem != "" {
			context.Text(settings.problem, ui.TextStyle{Color: theme.Danger})
		}
		context.Box(ui.Layout{Width: ui.Grow(), AlignX: ui.End}, func() {
			if context.PrimaryButton("Guardar redes") {
				screens.saveNetworks()
			}
		})
	}) {
		screens.settings = nil
	}
}

func (screens *Screens) networkRow(context *ui.Context, network *customers.Network) {
	name, pattern := "Nombre", "Enlace con "+customers.UserMark
	if network.ID == "" {
		name, pattern = "Nueva red", "https://ejemplo.com/"+customers.UserMark
	}
	context.Box(ui.Layout{Direction: ui.Vertical, Width: ui.Grow(), Gap: 6}, func() {
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignY: ui.Center}, func() {
			context.IconSized(nameOr(network.Icon, "at-sign"), context.Theme.IconSize, context.Theme.InkSoft)
			context.TextFieldWith("name", &network.Name, ui.TextFieldOptions{Placeholder: name})
			if network.ID != "" && context.FlatIconButton("trash-2", "##Quitar "+network.Name) {
				screens.deleteNetwork(*network)
			}
		})
		context.TextFieldWith("pattern", &network.LinkPattern, ui.TextFieldOptions{Placeholder: pattern})
	})
}

// saveNetworks stores every network typed; an empty new line is left out.
func (screens *Screens) saveNetworks() {
	for _, network := range screens.settings.typed {
		if network.ID == "" && network.Name == "" && network.LinkPattern == "" {
			continue
		}
		problem, err := screens.customers.SaveNetwork(network)
		if err != nil {
			screens.tell(true, "No se pudieron guardar las redes")
			return
		}
		if problem != "" {
			screens.settings.problem = nameOr(network.Name, "Nueva red") + ": " + customers.ProblemWords(problem)
			return
		}
	}
	screens.settings = nil
	screens.Reload()
	screens.tell(false, "Redes guardadas")
}

func nameOr(name, otherwise string) string {
	if name == "" {
		return otherwise
	}
	return name
}

// deleteNetwork removes a network and opens the sheet again without it.
func (screens *Screens) deleteNetwork(network customers.Network) {
	if err := screens.customers.DeleteNetwork(network.ID); err != nil {
		screens.tell(true, "No se pudo quitar "+network.Name)
		return
	}
	screens.Reload()
	screens.openNetworks()
	screens.tell(false, network.Name+" quitada")
}
