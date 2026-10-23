package demo

import (
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/money"
)

// catalogue saves the made-up products, with low and sold-out stock among them.
func (filler *filler) catalogue() {
	paper := "Papel de algodón 300 g"
	filler.physical("Lámina «Gato astronauta»", []string{"Gatos", "Espacio"}, paper, catalogue.Active,
		filler.item("A4", 1200, 350, "LAM-GAT-A4", 14, 4), filler.item("A3", 2000, 600, "LAM-GAT-A3", 6, 2))
	filler.physical("Lámina «Bosque de niebla»", []string{"Paisaje"}, paper, catalogue.Active,
		filler.item("A4", 1200, 350, "LAM-BOS-A4", 10, 0), filler.item("A3", 2000, 600, "LAM-BOS-A3", 3, 3))
	filler.physical("Pegatina holográfica «Luna»", []string{"Pegatinas", "Espacio"}, "Vinilo holográfico", catalogue.Active,
		filler.item("", 300, 60, "PEG-LUNA", 40, 0))
	filler.physical("Pack de pegatinas «Hongos»", []string{"Pegatinas", "Hongos"}, "Vinilo mate", catalogue.Active,
		filler.item("", 600, 150, "PEG-HONG", 2, 5))
	filler.physical("Chapa «Rana con sombrero»", []string{"Chapas"}, "Metal 38 mm", catalogue.Active,
		filler.item("", 250, 70, "CHA-RANA", 25, 0))
	filler.physical("Tote bag «Gato astronauta»", []string{"Gatos", "Merch"}, "Algodón orgánico", catalogue.Active,
		filler.item("", 1800, 700, "TOT-GAT", 0, 2))
	filler.physical("Llavero acrílico «Dragón»", []string{"Merch"}, "Acrílico 5 cm", catalogue.Active,
		filler.item("", 800, 220, "LLA-DRAG", 12, 0))
	filler.physical("Fanzine «Cuaderno de bocetos»", []string{"Fanzines"}, "Grapado A5, 24 páginas", catalogue.Draft,
		filler.item("", 700, 250, "FAN-BOC", 9, 0))
	filler.digital("Retrato digital", "PNG 4000 px", 21, filler.tier("Busto", 3500), filler.tier("Cuerpo entero", 7000))
	filler.digital("Icono para redes", "PNG 1000 px", 7, filler.tier("", 2000))
}

func (filler *filler) item(name string, price, cost money.Money, sku string, stock, minimum int) catalogue.Variant {
	return catalogue.Variant{ID: filler.stores.NewID(), Name: name, Price: price, Cost: cost, SKU: sku, Stock: stock, MinimumStock: minimum}
}

func (filler *filler) tier(name string, price money.Money) catalogue.Variant {
	return catalogue.Variant{ID: filler.stores.NewID(), Name: name, Price: price}
}

func (filler *filler) physical(name string, tags []string, material string, status catalogue.Status, variants ...catalogue.Variant) {
	filler.save(catalogue.Product{Kind: catalogue.Physical, Name: name, Status: status, Material: material, Variants: variants}, tags)
	for _, variant := range variants {
		filler.variants[variant.SKU] = variant.ID
	}
}

func (filler *filler) digital(name, format string, days int, variants ...catalogue.Variant) {
	filler.save(catalogue.Product{Kind: catalogue.Digital, Name: name, Status: catalogue.Active, DeliveryFormat: format,
		TurnaroundDays: days, Variants: variants}, []string{"Comisiones"})
	for _, variant := range variants {
		filler.variants[name+" · "+variant.Name] = variant.ID
	}
}

func (filler *filler) save(product catalogue.Product, tags []string) {
	if filler.err != nil {
		return
	}
	product.ID = filler.stores.NewID()
	_, problems, err := filler.stores.Products.Save(catalogue.ProductDraft{Product: product, TagNames: tags})
	filler.stop(len(problems), problems, err)
}
