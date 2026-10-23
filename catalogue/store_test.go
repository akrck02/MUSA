package catalogue

import (
	"errors"
	"slices"
	"testing"
)

func TestASavedProductIsReadBackWithItsVariantsAndTags(t *testing.T) {
	store := testStore(t)
	draft := physicalProduct("p1", "  Lámina gato ")
	draft.Variants = []Variant{{ID: "a4", Name: "A4", Price: 1200, Cost: 350, SKU: " LAM-A4 ", Stock: 14, MinimumStock: 4}, {ID: "a3", Name: "A3", Price: 2000, SKU: " "}}
	mustSave(t, store, ProductDraft{Product: draft, TagNames: []string{"Gatos", "gatos", "Espacio"}})

	product := mustFind(t, store, "p1")
	if product.Name != "Lámina gato" || product.CreatedAt.IsZero() || len(product.Variants) != 2 {
		t.Fatalf("read back %+v", product)
	}
	if product.Variants[0].SKU != "LAM-A4" || product.Variants[0].MinimumStock != 4 || product.Variants[1].SKU != "" {
		t.Errorf("variants %+v", product.Variants)
	}
	if got := []string{product.Tags[0].Name, product.Tags[1].Name}; len(product.Tags) != 2 || !slices.Equal(got, []string{"Gatos", "Espacio"}) {
		t.Errorf("tags %+v", product.Tags)
	}
}

func TestEditingKeepsTheCreationTimeAndReusesTags(t *testing.T) {
	store := testStore(t)
	first := mustSave(t, store, ProductDraft{Product: physicalProduct("p1", "Lámina"), TagNames: []string{"Gatos"}})
	edited := first
	edited.Name = "Lámina nueva"
	second := mustSave(t, store, ProductDraft{Product: edited, TagNames: []string{"GATOS"}})
	if !second.CreatedAt.Equal(first.CreatedAt) || !second.UpdatedAt.After(first.UpdatedAt) {
		t.Errorf("times %v %v", second.CreatedAt, second.UpdatedAt)
	}
	tags, _ := store.Tags()
	if len(tags) != 1 || second.Tags[0].ID != first.Tags[0].ID {
		t.Errorf("tags %+v", tags)
	}
}

func TestASKUUsedByAnotherProductIsRejected(t *testing.T) {
	store := testStore(t)
	owner := physicalProduct("p1", "Lámina")
	owner.Variants[0].SKU = "LAM"
	mustSave(t, store, ProductDraft{Product: owner})
	other := physicalProduct("p2", "Otra")
	other.Variants[0].SKU = "lam"
	_, problems, err := store.Save(ProductDraft{Product: other})
	if err != nil || len(problems) != 1 || problems[0].Code != CodeDuplicateSKU || problems[0].OwnerName != "Lámina" {
		t.Fatalf("problems %+v %v", problems, err)
	}
	if _, found, _ := store.Find("p2"); found {
		t.Error("an invalid draft was stored")
	}
}

func TestStockMovedElsewhereWhileEditingIsKept(t *testing.T) {
	cases := map[string]struct{ typed, baseline, elsewhere, want int }{
		"editor did not touch it":   {typed: 5, baseline: 5, elsewhere: -2, want: 3},
		"editor change on top":      {typed: 8, baseline: 5, elsewhere: -2, want: 6},
		"nothing changed elsewhere": {typed: 9, baseline: 5, elsewhere: 0, want: 9},
		"never below zero":          {typed: 0, baseline: 5, elsewhere: -4, want: 0},
	}
	for name, each := range cases {
		store := testStore(t)
		saved := mustSave(t, store, ProductDraft{Product: physicalProduct("p1", "Lámina")})
		variantID := saved.Variants[0].ID
		if err := store.AdjustStock(variantID, each.elsewhere); err != nil {
			t.Fatal(err)
		}
		edited := saved
		edited.Variants = []Variant{saved.Variants[0]}
		edited.Variants[0].Stock = each.typed
		mustSave(t, store, ProductDraft{Product: edited, StockBaseline: map[string]int{variantID: each.baseline}})
		if got := mustFind(t, store, "p1").Variants[0].Stock; got != each.want {
			t.Errorf("%s: stock %d; want %d", name, got, each.want)
		}
	}
}

func TestADuplicatedProductIsADraftWithNewIdsAndNoStock(t *testing.T) {
	store := testStore(t)
	original := physicalProduct("p1", "Lámina")
	original.Variants[0].SKU = "LAM"
	mustSave(t, store, ProductDraft{Product: original, TagNames: []string{"Gatos"}})
	copyID, err := store.Duplicate("p1", "Lámina (copia)")
	if err != nil {
		t.Fatal(err)
	}
	copied := mustFind(t, store, copyID)
	if copied.Status != Draft || copied.Name != "Lámina (copia)" || len(copied.Tags) != 1 {
		t.Errorf("copy %+v", copied)
	}
	if variant := copied.Variants[0]; variant.ID == "p1-v1" || variant.SKU != "" || variant.Stock != 0 || variant.Price != 1000 {
		t.Errorf("variant %+v", variant)
	}
	if _, err := store.Duplicate("missing", "x"); !hasCode(err, CodeProductNotFound) {
		t.Errorf("duplicating a missing product: %v", err)
	}
}

func TestArchivingDeletingAndStockChangesAreStored(t *testing.T) {
	store := testStore(t)
	mustSave(t, store, ProductDraft{Product: physicalProduct("p1", "Lámina")})
	store.SetArchived("p1", true)
	if mustFind(t, store, "p1").Status != Archived {
		t.Error("not archived")
	}
	store.SetArchived("p1", false)
	if mustFind(t, store, "p1").Status != Active {
		t.Error("not restored")
	}
	store.AdjustStock("p1-v1", 3)
	if stock, found, _ := store.Stock("p1-v1"); !found || stock != 8 {
		t.Errorf("stock %d", stock)
	}
	store.Delete("p1")
	if _, found, _ := store.Find("p1"); found {
		t.Error("a deleted product is read")
	}
	if all, _ := store.All(); len(all) != 0 {
		t.Errorf("%d products after deleting", len(all))
	}
}

func TestAVariantRemovedInTheEditorLeavesTheProduct(t *testing.T) {
	store := testStore(t)
	product := physicalProduct("p1", "Lámina")
	product.Variants = []Variant{{ID: "a", Name: "A4"}, {ID: "b", Name: "A3"}}
	saved := mustSave(t, store, ProductDraft{Product: product})
	saved.Variants = saved.Variants[:1]
	mustSave(t, store, ProductDraft{Product: saved})
	if got := mustFind(t, store, "p1").Variants; len(got) != 1 || got[0].ID != "a" {
		t.Errorf("variants %+v", got)
	}
}

func hasCode(err error, code Code) bool {
	var failure *Error
	return errors.As(err, &failure) && failure.Code == code
}

func TestEachErrorConstructorCarriesItsOwnCode(t *testing.T) {
	cases := map[Code]error{
		CodeProductNotFound: productNotFound("x"),
		CodeVariantNotFound: variantNotFound("x"),
		CodeStorageFailed:   storageFailed("x", errors.New("cause")),
	}
	for code, err := range cases {
		if !hasCode(err, code) {
			t.Errorf("%v does not carry %s", err, code)
		}
	}
}
