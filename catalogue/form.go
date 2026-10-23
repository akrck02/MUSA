package catalogue

import (
	"strconv"

	"github.com/Itros97/MUSA/money"
)

// form is the product editor's text, as typed: numbers stay text until saving, so a half-typed
// price is never lost or rounded.
type form struct {
	product     Product // what is not typed: id, kind, times
	isNew       bool
	name        string
	description string
	status      int // index in statuses
	tags        []string
	tagTyped    string
	material    string
	delivery    string
	turnaround  string
	variants    []variantForm
	baseline    map[string]int
}

// variantForm is one variant's fields, as typed.
type variantForm struct {
	id, name, price, cost, sku, stock, minimum string
	weight, width, height, depth               string
}

// newForm is an empty editor for a new product of kind, with one unnamed variant.
func newForm(kind Kind, newID func() string) form {
	return form{product: Product{ID: newID(), Kind: kind}, isNew: true, status: 1,
		variants: []variantForm{{id: newID(), stock: "0"}}, baseline: map[string]int{}}
}

// formOf is the editor filled with a stored product.
func formOf(product Product) form {
	edited := form{product: product, name: product.Name, description: product.Description, material: product.Material,
		delivery: product.DeliveryFormat, baseline: map[string]int{}}
	for index, status := range statuses {
		if status == product.Status {
			edited.status = index
		}
	}
	for _, tag := range product.Tags {
		edited.tags = append(edited.tags, tag.Name)
	}
	if product.TurnaroundDays > 0 {
		edited.turnaround = strconv.Itoa(product.TurnaroundDays)
	}
	for _, variant := range product.Variants {
		edited.variants = append(edited.variants, variantFormOf(variant))
		edited.baseline[variant.ID] = variant.Stock
	}
	return edited
}

func variantFormOf(variant Variant) variantForm {
	optional := func(value int) string {
		if value == 0 {
			return ""
		}
		return strconv.Itoa(value)
	}
	centimetres := func(millimetres int) string {
		if millimetres == 0 {
			return ""
		}
		return FormatCentimetres(millimetres)
	}
	return variantForm{id: variant.ID, name: variant.Name, price: money.FormatForInput(variant.Price),
		cost: money.FormatForInput(variant.Cost), sku: variant.SKU, stock: strconv.Itoa(variant.Stock),
		minimum: optional(variant.MinimumStock), weight: optional(variant.WeightGrams),
		width: centimetres(variant.WidthMillimetres), height: centimetres(variant.HeightMillimetres), depth: centimetres(variant.DepthMillimetres)}
}

// fieldKey names one field of one variant, for the problems shown beside it.
func fieldKey(variantID, field string) string { return variantID + "/" + field }

// typing is what the artist reads beside a field whose text is not a number of its kind.
var typing = map[string]string{
	"price": "Escribe un precio como 12 o 12,50", "cost": "Escribe un coste como 3 o 3,50",
	"stock": "Escribe un número entero", "minimum": "Escribe un número entero", "weight": "Escribe los gramos",
	"width": "Escribe centímetros, como 29,7", "height": "Escribe centímetros, como 29,7", "depth": "Escribe centímetros, como 29,7",
	"turnaround": "Escribe los días",
}

// draft reads the typed text into a draft; unreadable fields come back as problems, keyed by field.
func (edited form) draft() (ProductDraft, map[string]string) {
	unreadable := map[string]string{}
	product := edited.product
	product.Name, product.Description, product.Status = edited.name, edited.description, statuses[edited.status]
	product.Material, product.DeliveryFormat = edited.material, edited.delivery
	product.TurnaroundDays = readCount(edited.turnaround, "turnaround", unreadable)
	product.Variants = nil
	for _, typed := range edited.variants {
		product.Variants = append(product.Variants, typed.variant(product.Kind, unreadable))
	}
	tags := edited.tags
	if TagDisplayName(edited.tagTyped) != "" {
		tags = append(tags, edited.tagTyped) // a tag typed but not yet added is meant
	}
	return ProductDraft{Product: product, TagNames: tags, StockBaseline: edited.baseline}, unreadable
}

func (typed variantForm) variant(kind Kind, unreadable map[string]string) Variant {
	key := func(field string) string { return fieldKey(typed.id, field) }
	variant := Variant{ID: typed.id, Name: typed.name, Price: readMoney(typed.price, key("price"), unreadable),
		Cost: readMoney(typed.cost, key("cost"), unreadable)}
	if kind != Physical {
		return variant
	}
	variant.SKU = typed.sku
	variant.Stock = readCount(typed.stock, key("stock"), unreadable)
	variant.MinimumStock = readCount(typed.minimum, key("minimum"), unreadable)
	variant.WeightGrams = readCount(typed.weight, key("weight"), unreadable)
	variant.WidthMillimetres = readCentimetres(typed.width, key("width"), unreadable)
	variant.HeightMillimetres = readCentimetres(typed.height, key("height"), unreadable)
	variant.DepthMillimetres = readCentimetres(typed.depth, key("depth"), unreadable)
	return variant
}

// readMoney reads a price; empty is zero, which a free cost often is.
func readMoney(text, key string, unreadable map[string]string) money.Money {
	if text == "" {
		return 0
	}
	amount, found := money.Parse(text)
	if !found {
		unreadable[key] = typing[fieldOf(key)]
	}
	return amount
}

func readCount(text, key string, unreadable map[string]string) int {
	if text == "" {
		return 0
	}
	count, found := ParseCount(text)
	if !found {
		unreadable[key] = typing[fieldOf(key)]
	}
	return count
}

func readCentimetres(text, key string, unreadable map[string]string) int {
	if text == "" {
		return 0
	}
	millimetres, found := ParseCentimetres(text)
	if !found {
		unreadable[key] = typing[fieldOf(key)]
	}
	return millimetres
}

func fieldOf(key string) string {
	for index := len(key) - 1; index >= 0; index-- {
		if key[index] == '/' {
			return key[index+1:]
		}
	}
	return key
}
