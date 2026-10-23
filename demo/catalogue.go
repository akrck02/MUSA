package demo

import (
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/money"
)

// catalogue saves the made-up products, with low and sold-out stock among them.
func (filler *filler) catalogue() {
	paper := "Papel de algodón 300 g"
	vinyl := "Vinilo"
	filler.physical("Gato astronauta", []string{"Láminas"}, paper, catalogue.Active,
		filler.item("A5", 800, 200, "LAM-GAT-A5", 23, 3), filler.item("A4", 1200, 350, "LAM-GAT-A4", 5, 3),
		filler.item("A3", 1500, 600, "LAM-GAT-A3", 1, 2))
	filler.physical("Luna", []string{"Pegatinas"}, vinyl, catalogue.Active,
		filler.item("Mate", 300, 60, "PEG-LUN-M", 3, 5), filler.item("Brillo", 300, 70, "PEG-LUN-B", 47, 5))
	filler.physical("Bosque de noche", []string{"Láminas"}, paper, catalogue.Active,
		filler.item("A4", 1000, 300, "LAM-BOS-A4", 11, 3), filler.item("A3", 1800, 600, "LAM-BOS-A3", 0, 2))
	filler.physical("Rana feliz", []string{"Llaveros"}, "Acrílico 5 cm", catalogue.Active,
		filler.item("", 600, 180, "LLA-RAN", 16, 3))
	filler.physical("Ola", []string{"Pegatinas"}, vinyl, catalogue.Active, filler.item("", 300, 60, "PEG-OLA", 16, 5))
	filler.physical("Nube", []string{"Pegatinas"}, vinyl, catalogue.Active, filler.item("", 300, 60, "PEG-NUB", 32, 5))
	filler.physical("Seta", []string{"Pegatinas"}, vinyl, catalogue.Active, filler.item("", 300, 60, "PEG-SET", 0, 5))
	filler.physical("Zorro", []string{"Láminas"}, paper, catalogue.Active,
		filler.item("A5", 800, 200, "LAM-ZOR-A5", 6, 3), filler.item("A4", 1500, 350, "LAM-ZOR-A4", 9, 3))
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
