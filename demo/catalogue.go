package demo

import (
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/money"
)

// catalogue saves the made-up works, as the author's design shows them: twenty-one physical ones
// (prints, stickers and key rings), five of them to restock, and three digital ones.
func (filler *filler) catalogue() {
	paper, vinyl, acrylic := "Papel de algodón 300 g", "Vinilo", "Acrílico 5 cm"
	print, sticker, keyRing := []string{"Lámina"}, []string{"Pegatina"}, []string{"Llavero"}
	filler.physical("Gato astronauta", print, paper, filler.item("A5", 800, 200, "LAM-GAT-A5", 12, 0),
		filler.item("A4", 1200, 350, "LAM-GAT-A4", 4, 0), filler.item("A3", 1500, 600, "LAM-GAT-A3", 1, 2))
	filler.physical("Luna", sticker, vinyl, filler.item("Mate", 300, 60, "PEG-LUN-M", 3, 5), filler.item("Brillo", 300, 70, "PEG-LUN-B", 20, 0))
	filler.physical("Bosque de noche", print, paper, filler.item("A4", 1000, 300, "LAM-BOS-A4", 9, 0), filler.item("A3", 1800, 600, "LAM-BOS-A3", 0, 0))
	filler.physical("Rana feliz", keyRing, acrylic, filler.item("", 600, 180, "LLA-RAN", 8, 0))
	filler.physical("Ola", sticker, vinyl, filler.item("", 300, 60, "PEG-OLA", 15, 0))
	filler.physical("Nube", sticker, vinyl, filler.item("", 300, 60, "PEG-NUB", 22, 0))
	filler.physical("Seta", sticker, vinyl, filler.item("", 300, 60, "PEG-SET", 0, 0))
	filler.physical("Zorro", print, paper, filler.item("A5", 800, 200, "LAM-ZOR-A5", 6, 0), filler.item("A4", 1500, 350, "LAM-ZOR-A4", 7, 0))
	filler.physical("Café", sticker, vinyl, filler.item("", 300, 60, "PEG-CAF", 1, 2))
	filler.physical("Planta", sticker, vinyl, filler.item("", 300, 60, "PEG-PLA", 6, 0))
	filler.physical("Cereza", sticker, vinyl, filler.item("", 300, 60, "PEG-CER", 4, 0))
	filler.physical("Cactus", sticker, vinyl, filler.item("", 300, 60, "PEG-CAC", 7, 0))
	filler.physical("Búho", keyRing, acrylic, filler.item("", 600, 180, "LLA-BUH", 2, 0))
	filler.physical("Taco", keyRing, acrylic, filler.item("", 600, 180, "LLA-TAC", 11, 0))
	filler.physical("Ballena", keyRing, acrylic, filler.item("", 600, 180, "LLA-BAL", 5, 0))
	filler.physical("Estrella", print, paper, filler.item("A4", 1000, 300, "LAM-EST-A4", 3, 0))
	filler.physical("Pez koi", print, paper, filler.item("A4", 1000, 300, "LAM-KOI-A4", 13, 0))
	filler.physical("Abeja", print, paper, filler.item("A4", 1000, 300, "LAM-ABE-A4", 6, 0))
	filler.physical("Donut", print, paper, filler.item("A4", 1000, 300, "LAM-DON-A4", 18, 0))
	filler.physical("Faro", print, paper, filler.item("A4", 1000, 300, "LAM-FAR-A4", 10, 0))
	filler.physical("Medusa", print, paper, filler.item("A4", 1000, 300, "LAM-MED-A4", 9, 0))
	filler.digital("Pack de pinceles", "Procreate", 0, filler.tier("", 1200))
	filler.digital("Fondos para móvil", "PNG", 0, filler.tier("", 500))
	filler.digital("Retrato por encargo", "PNG + PSD", 7, filler.tier("", 4000))
}

// item is a version with the stock it should have once the demo is filled: the fairs and orders
// made after take theirs from more, so the catalogue ends as the design shows it.
func (filler *filler) item(name string, price, cost money.Money, sku string, stock, minimum int) catalogue.Variant {
	filler.finalStock[sku] = stock
	return catalogue.Variant{ID: filler.stores.NewID(), Name: name, Price: price, Cost: cost, SKU: sku, Stock: stock + spareStock, MinimumStock: minimum}
}

// spareStock is what each version starts with on top of its final stock, for the demo's sales.
const spareStock = 40

// settleStock leaves every version with the stock item said, whatever the sales took.
func (filler *filler) settleStock() {
	products, err := filler.stores.Products.All()
	filler.stop(0, nil, err)
	for _, product := range products {
		for _, variant := range product.Variants {
			wanted, found := filler.finalStock[variant.SKU]
			if !found || filler.err != nil {
				continue
			}
			filler.stop(0, nil, filler.stores.Products.AdjustStock(variant.ID, wanted-variant.Stock))
		}
	}
}

func (filler *filler) tier(name string, price money.Money) catalogue.Variant {
	return catalogue.Variant{ID: filler.stores.NewID(), Name: name, Price: price}
}

func (filler *filler) physical(name string, tags []string, material string, variants ...catalogue.Variant) {
	filler.save(catalogue.Product{Kind: catalogue.Physical, Name: name, Status: catalogue.Active, Material: material, Variants: variants}, tags)
	for _, variant := range variants {
		filler.variants[variant.SKU] = variant.ID
	}
}

func (filler *filler) digital(name, format string, days int, variants ...catalogue.Variant) {
	filler.save(catalogue.Product{Kind: catalogue.Digital, Name: name, Status: catalogue.Active, DeliveryFormat: format,
		TurnaroundDays: days, Variants: variants}, nil)
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
