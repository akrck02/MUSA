package events

import (
	"strconv"

	ui "github.com/akrck02/go-ui"

	"github.com/Itros97/MUSA/look"
	"github.com/Itros97/MUSA/money"
)

func (screens *Screens) drawSheet(context *ui.Context) {
	titles := map[sheetKind]string{takeSheet: "Llevar stock al evento", sellSheet: "Vender", returnSheet: "Devolver al catálogo",
		closeSheet: "¿Cerrar el evento?", deleteSheet: "¿Eliminar el evento?"}
	current := &screens.sheet
	if context.Sheet(titles[current.kind], ui.SheetOptions{}, func() {
		switch current.kind {
		case takeSheet:
			screens.takeFields(context)
		case sellSheet:
			look.Field(context, "Cantidad", "quantity", &current.quantity, "1", "")
			look.Field(context, "Precio por unidad (€)", "price", &current.price, "", "")
		case returnSheet:
			look.Field(context, "Cantidad", "quantity", &current.quantity, "1", "")
		case closeSheet:
			look.Paragraph(context, "Lo que no vendiste vuelve al catálogo y el evento queda de solo lectura.", ui.TextStyle{})
		case deleteSheet:
			look.Paragraph(context, "Su stock vuelve al catálogo. Solo se puede eliminar un evento sin ventas.", ui.TextStyle{})
		}
		if current.problem != "" {
			context.Text(current.problem, ui.TextStyle{Color: context.Theme.Danger})
		}
		context.Box(ui.Layout{Width: ui.Grow(), Gap: 8, AlignX: ui.End}, func() {
			if context.Button("Cancelar") {
				screens.sheet = sheet{}
			}
			if context.PrimaryButton(map[sheetKind]string{closeSheet: "Cerrar", deleteSheet: "Eliminar"}[current.kind] + "##confirm") {
				screens.confirmSheet()
			}
		})
	}) {
		screens.sheet = sheet{}
	}
}

func (screens *Screens) takeFields(context *ui.Context) {
	current := &screens.sheet
	if len(current.choices) == 0 {
		context.Text("No queda stock físico en el catálogo.", ui.TextStyle{Color: context.Theme.InkSoft})
		return
	}
	var words []string
	for _, each := range current.choices {
		words = append(words, each.words)
	}
	context.Select("variant", &current.choice, words)
	look.Field(context, "Cantidad", "quantity", &current.quantity, "1", "")
}

func (screens *Screens) confirmSheet() {
	current := &screens.sheet
	eventID := screens.detail.Event.ID
	quantity, err := strconv.Atoi(current.quantity)
	if err != nil && (current.kind == takeSheet || current.kind == sellSheet || current.kind == returnSheet) {
		current.problem = "Escribe un número entero"
		return
	}
	var refused []Refusal
	switch current.kind {
	case takeSheet:
		if len(current.choices) == 0 {
			screens.sheet = sheet{}
			return
		}
		refused, err = screens.store.MoveStock(eventID, current.choices[current.choice].variantID, quantity)
	case sellSheet:
		price, found := money.Parse(current.price)
		if !found {
			current.problem = "Escribe un precio como 12 o 12,50"
			return
		}
		refused, err = screens.store.Sell(eventID, current.variantID, quantity, price)
	case returnSheet:
		refused, err = screens.store.ReturnStock(eventID, current.variantID, quantity)
	case closeSheet:
		refused, err = screens.store.Close(eventID)
	case deleteSheet:
		refused, err = screens.store.Delete(eventID)
	}
	if err == nil && len(refused) > 0 {
		current.problem = RefusalWords(refused[0])
		return
	}
	kind := current.kind
	screens.sheet = sheet{}
	screens.tell(refused, err, map[sheetKind]string{takeSheet: "Stock llevado al evento", sellSheet: "Venta apuntada",
		returnSheet: "Stock devuelto al catálogo", closeSheet: "Evento cerrado", deleteSheet: "Evento eliminado"}[kind])
	if kind == deleteSheet {
		screens.detail = nil
		screens.Reload()
	}
}
