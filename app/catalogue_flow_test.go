package app

import (
	"slices"
	"testing"
	"time"

	"github.com/Itros97/MUSA/catalogue"
	"github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"
)

// productNamed is the stored product with a name, and whether there is one.
func productNamed(t *testing.T, program *App, name string) (catalogue.Product, bool) {
	t.Helper()
	all, err := program.stores.Products.All()
	if err != nil {
		t.Fatal(err)
	}
	for _, product := range all {
		if product.Name == name {
			return product, true
		}
	}
	return catalogue.Product{}, false
}

// settle lets animations end, so what a frame shows is where it stays.
func settle(window *uitest.Window) {
	window.Frames(2)
	window.Wait(time.Second)
	window.Frames(2)
}

// openProductEditor opens the catalogue and the editor of a product of one version.
func openProductEditor(t *testing.T, name string) (*App, *uitest.Window) {
	t.Helper()
	program, window := openApp(t, true)
	program.Open(Catalogue)
	settle(window)
	steps(t, window, click(window, name))
	settle(window)
	return program, window
}

func TestSearchingTheCatalogueKeepsOnlyWhatMatches(t *testing.T) {
	program, window := openApp(t, true)
	program.Open(Catalogue)
	settle(window)
	steps(t, window, click(window, "Buscar"))
	settle(window)
	window.Type("rana")
	settle(window)
	if _, err := window.Element("Rana feliz"); err != nil {
		t.Errorf("«Rana feliz» is not found by «rana»: %v", err)
	}
	if _, err := window.Element("Luna"); err == nil {
		t.Errorf("«Luna» still shows when searching «rana»")
	}
}

func TestTheStockSteppersChangeTheStockThatIsSaved(t *testing.T) {
	program, window := openProductEditor(t, "Rana feliz")
	before, _ := productNamed(t, program, "Rana feliz")
	steps(t, window, click(window, "Añadir uno"), click(window, "Añadir uno"), click(window, "Quitar uno"), click(window, "Guardar"))
	settle(window)
	after, _ := productNamed(t, program, "Rana feliz")
	if after.Variants[0].Stock != before.Variants[0].Stock+1 {
		t.Errorf("stock %d after +2 −1 from %d", after.Variants[0].Stock, before.Variants[0].Stock)
	}
}

func TestDuplicatingAProductMakesACopy(t *testing.T) {
	program, window := openProductEditor(t, "Rana feliz")
	steps(t, window, click(window, "Duplicar"))
	settle(window)
	if _, found := productNamed(t, program, "Rana feliz (copia)"); !found {
		t.Errorf("there is no «Rana feliz (copia)» after Duplicar")
	}
}

func TestAnArchivedProductLeavesTheListAndComesBack(t *testing.T) {
	program, window := openProductEditor(t, "Rana feliz")
	steps(t, window, click(window, "Más detalles SKU, medidas, peso y aviso de stock bajo"))
	settle(window)
	steps(t, window, click(window, "Archivar"))
	settle(window)
	archived, _ := productNamed(t, program, "Rana feliz")
	if archived.Status != catalogue.Archived {
		t.Fatalf("the product is %v after Archivar", archived.Status)
	}
	if _, err := window.Element("Rana feliz"); err == nil {
		t.Errorf("an archived product still shows in the list")
	}
	steps(t, window, click(window, "Archivados"))
	settle(window)
	steps(t, window, click(window, "Rana feliz"))
	settle(window)
	steps(t, window, click(window, "Más detalles SKU, medidas, peso y aviso de stock bajo"))
	settle(window)
	steps(t, window, click(window, "Restaurar"))
	settle(window)
	if restored, _ := productNamed(t, program, "Rana feliz"); restored.Status == catalogue.Archived {
		t.Errorf("the product is still archived after Restaurar")
	}
}

func TestDeletingAProductAsksFirst(t *testing.T) {
	program, window := openProductEditor(t, "Rana feliz")
	steps(t, window, click(window, "Más detalles SKU, medidas, peso y aviso de stock bajo"))
	settle(window)
	steps(t, window, click(window, "Eliminar"))
	settle(window)
	if _, found := productNamed(t, program, "Rana feliz"); !found {
		t.Fatalf("the product went before the question was answered")
	}
	steps(t, window, clickIn(window, "¿Eliminar «Rana feliz»?", "Eliminar"))
	settle(window)
	if _, found := productNamed(t, program, "Rana feliz"); found {
		t.Errorf("the product is still there after confirming")
	}
}

func TestATagIsAddedAndRemovedFromAProduct(t *testing.T) {
	program, window := openProductEditor(t, "Rana feliz")
	steps(t, window, click(window, "Añadir etiqueta"))
	settle(window)
	steps(t, window, typeInto(window, "Lámina, pegatina, gatos…", "verde"))
	steps(t, window, click(window, "Añadir etiqueta"))
	settle(window)
	steps(t, window, click(window, "Guardar"))
	settle(window)
	saved, _ := productNamed(t, program, "Rana feliz")
	if !hasTag(saved, "verde") {
		t.Fatalf("the product's tags are %v, without «verde»", saved.Tags)
	}
}

func hasTag(product catalogue.Product, name string) bool {
	for _, tag := range product.Tags {
		if tag.Name == name {
			return true
		}
	}
	return false
}

func TestASecondVersionNeedsNamesForBoth(t *testing.T) {
	program, window := openProductEditor(t, "Rana feliz")
	steps(t, window, click(window, "Añadir otra versión"))
	settle(window)
	steps(t, window, click(window, "Guardar"))
	settle(window)
	if saved, _ := productNamed(t, program, "Rana feliz"); len(saved.Variants) != 1 {
		t.Fatalf("two versions without names were saved")
	}
	if !slices.Contains(window.Texts(), "Con varias versiones, cada una necesita nombre") {
		t.Fatalf("the editor does not say why it did not save")
	}
	names := window.Context().ScreenReaderNodes()
	var fields []ui.AccessibleNode
	for _, node := range names {
		if node.Label == "Única" && node.Pressable {
			fields = append(fields, node)
		}
	}
	if len(fields) != 2 {
		t.Fatalf("%d version name fields, not 2", len(fields))
	}
	for index, words := range []string{"Verde", "Azul"} {
		field := fields[index].Area
		window.ClickAt(field.X+field.Width/2, field.Y+field.Height/2)
		window.Type(words)
	}
	steps(t, window, click(window, "Guardar"))
	settle(window)
	if saved, _ := productNamed(t, program, "Rana feliz"); len(saved.Variants) != 2 {
		t.Errorf("the product has %d versions after naming both", len(saved.Variants))
	}
}
