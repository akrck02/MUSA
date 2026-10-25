package catalogue

import (
	"errors"
	"slices"
	"testing"
)

func TestASavedProductIsReadBackWithItsVariantsAndTags(test *testing.T) {
	store := testStore(test)
	draft := physicalProduct("p1", "  Lámina gato ")
	draft.Variants = []Variant{{ID: "a4", Name: "A4", Price: 1200, Cost: 350, SKU: " LAM-A4 ", Stock: 14, MinimumStock: 4}, {ID: "a3", Name: "A3", Price: 2000, SKU: " "}}
	mustSave(test, store, ProductDraft{Product: draft, TagNames: []string{"Gatos", "gatos", "Espacio"}})

	product := mustFind(test, store, "p1")
	if product.Name != "Lámina gato" || product.CreatedAt.IsZero() || len(product.Variants) != 2 {
		test.Fatalf("read back %+v", product)
	}
	if product.Variants[0].SKU != "LAM-A4" || product.Variants[0].MinimumStock != 4 || product.Variants[1].SKU != "" {
		test.Errorf("variants %+v", product.Variants)
	}
	if got := []string{product.Tags[0].Name, product.Tags[1].Name}; len(product.Tags) != 2 || !slices.Equal(got, []string{"Gatos", "Espacio"}) {
		test.Errorf("tags %+v", product.Tags)
	}
}

func TestEditingKeepsTheCreationTimeAndReusesTags(test *testing.T) {
	store := testStore(test)
	first := mustSave(test, store, ProductDraft{Product: physicalProduct("p1", "Lámina"), TagNames: []string{"Gatos"}})
	edited := first
	edited.Name = "Lámina nueva"
	second := mustSave(test, store, ProductDraft{Product: edited, TagNames: []string{"GATOS"}})
	if !second.CreatedAt.Equal(first.CreatedAt) || !second.UpdatedAt.After(first.UpdatedAt) {
		test.Errorf("times %v %v", second.CreatedAt, second.UpdatedAt)
	}
	tags, _ := store.Tags()
	if len(tags) != 1 || second.Tags[0].ID != first.Tags[0].ID {
		test.Errorf("tags %+v", tags)
	}
}

func TestASKUUsedByAnotherProductIsRejected(test *testing.T) {
	store := testStore(test)
	owner := physicalProduct("p1", "Lámina")
	owner.Variants[0].SKU = "LAM"
	mustSave(test, store, ProductDraft{Product: owner})
	other := physicalProduct("p2", "Otra")
	other.Variants[0].SKU = "lam"
	_, problems, err := store.Save(ProductDraft{Product: other})
	if err != nil || len(problems) != 1 || problems[0].Code != CodeDuplicateSKU || problems[0].OwnerName != "Lámina" {
		test.Fatalf("problems %+v %v", problems, err)
	}
	if _, found, _ := store.Find("p2"); found {
		test.Error("an invalid draft was stored")
	}
}

func TestStockMovedElsewhereWhileEditingIsKept(test *testing.T) {
	cases := map[string]struct{ typed, baseline, elsewhere, want int }{
		"editor did not touch it":   {typed: 5, baseline: 5, elsewhere: -2, want: 3},
		"editor change on top":      {typed: 8, baseline: 5, elsewhere: -2, want: 6},
		"nothing changed elsewhere": {typed: 9, baseline: 5, elsewhere: 0, want: 9},
		"never below zero":          {typed: 0, baseline: 5, elsewhere: -4, want: 0},
	}
	for name, each := range cases {
		store := testStore(test)
		saved := mustSave(test, store, ProductDraft{Product: physicalProduct("p1", "Lámina")})
		variantID := saved.Variants[0].ID
		if err := store.AdjustStock(variantID, each.elsewhere); err != nil {
			test.Fatal(err)
		}
		edited := saved
		edited.Variants = []Variant{saved.Variants[0]}
		edited.Variants[0].Stock = each.typed
		mustSave(test, store, ProductDraft{Product: edited, StockBaseline: map[string]int{variantID: each.baseline}})
		if got := mustFind(test, store, "p1").Variants[0].Stock; got != each.want {
			test.Errorf("%s: stock %d; want %d", name, got, each.want)
		}
	}
}

func TestADuplicatedProductIsADraftWithNewIdsAndNoStock(test *testing.T) {
	store := testStore(test)
	original := physicalProduct("p1", "Lámina")
	original.Variants[0].SKU = "LAM"
	mustSave(test, store, ProductDraft{Product: original, TagNames: []string{"Gatos"}})
	copyID, err := store.Duplicate("p1", "Lámina (copia)")
	if err != nil {
		test.Fatal(err)
	}
	copied := mustFind(test, store, copyID)
	if copied.Status != Draft || copied.Name != "Lámina (copia)" || len(copied.Tags) != 1 {
		test.Errorf("copy %+v", copied)
	}
	if variant := copied.Variants[0]; variant.ID == "p1-v1" || variant.SKU != "" || variant.Stock != 0 || variant.Price != 1000 {
		test.Errorf("variant %+v", variant)
	}
	if _, err := store.Duplicate("missing", "x"); !hasCode(err, CodeProductNotFound) {
		test.Errorf("duplicating a missing product: %v", err)
	}
}

func TestArchivingDeletingAndStockChangesAreStored(test *testing.T) {
	store := testStore(test)
	mustSave(test, store, ProductDraft{Product: physicalProduct("p1", "Lámina")})
	store.SetArchived("p1", true)
	if mustFind(test, store, "p1").Status != Archived {
		test.Error("not archived")
	}
	store.SetArchived("p1", false)
	if mustFind(test, store, "p1").Status != Active {
		test.Error("not restored")
	}
	store.AdjustStock("p1-v1", 3)
	if stock, found, _ := store.Stock("p1-v1"); !found || stock != 8 {
		test.Errorf("stock %d", stock)
	}
	store.Delete("p1")
	if _, found, _ := store.Find("p1"); found {
		test.Error("a deleted product is read")
	}
	if all, _ := store.All(); len(all) != 0 {
		test.Errorf("%d products after deleting", len(all))
	}
}

func TestAVariantRemovedInTheEditorLeavesTheProduct(test *testing.T) {
	store := testStore(test)
	product := physicalProduct("p1", "Lámina")
	product.Variants = []Variant{{ID: "a", Name: "A4"}, {ID: "b", Name: "A3"}}
	saved := mustSave(test, store, ProductDraft{Product: product})
	saved.Variants = saved.Variants[:1]
	mustSave(test, store, ProductDraft{Product: saved})
	if got := mustFind(test, store, "p1").Variants; len(got) != 1 || got[0].ID != "a" {
		test.Errorf("variants %+v", got)
	}
}

func hasCode(err error, code Code) bool {
	var failure *Error
	return errors.As(err, &failure) && failure.Code == code
}

func TestEachErrorConstructorCarriesItsOwnCode(test *testing.T) {
	cases := map[Code]error{
		CodeProductNotFound: productNotFound("x"),
		CodeVariantNotFound: variantNotFound("x"),
		CodeStorageFailed:   storageFailed("x", errors.New("cause")),
	}
	for code, err := range cases {
		if !hasCode(err, code) {
			test.Errorf("%v does not carry %s", err, code)
		}
	}
}

func TestATagKeepsTheIconChosenForItAndForgetsItOnRequest(test *testing.T) {
	store := testStore(test)
	mustSave(test, store, ProductDraft{Product: physicalProduct("p1", "Luna"), TagNames: []string{"cosis"}})
	tag := mustFind(test, store, "p1").Tags[0]
	if tag.Icon() != "tag" {
		test.Fatalf("an unknown tag starts with %q, not the plain tag", tag.Icon())
	}
	if err := store.SetTagIcon(tag.ID, "sparkles"); err != nil {
		test.Fatal(err)
	}
	if icon := mustFind(test, store, "p1").Tags[0].Icon(); icon != "sparkles" {
		test.Errorf("the product's tag shows %q after choosing «sparkles»", icon)
	}
	if err := store.SetTagIcon(tag.ID, ""); err != nil {
		test.Fatal(err)
	}
	tags, err := store.Tags()
	if err != nil {
		test.Fatal(err)
	}
	if tags[0].Icon() != "tag" {
		test.Errorf("the tag shows %q after going back to the guessed icon", tags[0].Icon())
	}
}
