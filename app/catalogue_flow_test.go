package app

import (
	"slices"
	"testing"
	"time"

	"github.com/Itros97/MUSA/catalogue"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-ui"
	"github.com/akrck02/go-ui/uitest"
)

// productNamed is the stored product with a name, and whether there is one.
func productNamed(test *testing.T, program *App, name string) (catalogue.Product, bool) {
	test.Helper()
	all, err := program.stores.Products.All()
	if err != nil {
		test.Fatal(err)
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
func openProductEditor(test *testing.T, name string) (*App, *uitest.Window) {
	test.Helper()
	program, window := openApp(test, true)
	program.Open(Catalogue)
	settle(window)
	steps(test, window, click(window, name))
	settle(window)
	return program, window
}

func TestSearchingTheCatalogueKeepsOnlyWhatMatches(test *testing.T) {
	program, window := openApp(test, true)
	program.Open(Catalogue)
	settle(window)
	steps(test, window, click(window, "Buscar"))
	settle(window)
	window.Type("rana")
	settle(window)
	if _, err := window.Element("Rana feliz"); err != nil {
		test.Errorf("«Rana feliz» is not found by «rana»: %v", err)
	}
	if _, err := window.Element("Luna"); err == nil {
		test.Errorf("«Luna» still shows when searching «rana»")
	}
}

func TestTheStockSteppersChangeTheStockThatIsSaved(test *testing.T) {
	program, window := openProductEditor(test, "Rana feliz")
	before, _ := productNamed(test, program, "Rana feliz")
	steps(test, window, click(window, "Añadir uno"), click(window, "Añadir uno"), click(window, "Quitar uno"), click(window, "Guardar"))
	settle(window)
	after, _ := productNamed(test, program, "Rana feliz")
	if after.Variants[0].Stock != before.Variants[0].Stock+1 {
		test.Errorf("stock %d after +2 −1 from %d", after.Variants[0].Stock, before.Variants[0].Stock)
	}
}

func TestDuplicatingAProductMakesACopy(test *testing.T) {
	program, window := openProductEditor(test, "Rana feliz")
	steps(test, window, click(window, "Duplicar"))
	settle(window)
	if _, found := productNamed(test, program, "Rana feliz (copia)"); !found {
		test.Errorf("there is no «Rana feliz (copia)» after Duplicar")
	}
}

func TestAnArchivedProductLeavesTheListAndComesBack(test *testing.T) {
	program, window := openProductEditor(test, "Rana feliz")
	steps(test, window, click(window, "Más detalles SKU, medidas, peso y aviso de stock bajo"))
	settle(window)
	steps(test, window, click(window, "Archivar"))
	settle(window)
	archived, _ := productNamed(test, program, "Rana feliz")
	if archived.Status != catalogue.Archived {
		test.Fatalf("the product is %v after Archivar", archived.Status)
	}
	if _, err := window.Element("Rana feliz"); err == nil {
		test.Errorf("an archived product still shows in the list")
	}
	steps(test, window, click(window, "Archivados"))
	settle(window)
	steps(test, window, click(window, "Rana feliz"))
	settle(window)
	steps(test, window, click(window, "Más detalles SKU, medidas, peso y aviso de stock bajo"))
	settle(window)
	steps(test, window, click(window, "Restaurar"))
	settle(window)
	if restored, _ := productNamed(test, program, "Rana feliz"); restored.Status == catalogue.Archived {
		test.Errorf("the product is still archived after Restaurar")
	}
}

func TestDeletingAProductAsksFirst(test *testing.T) {
	program, window := openProductEditor(test, "Rana feliz")
	steps(test, window, click(window, "Más detalles SKU, medidas, peso y aviso de stock bajo"))
	settle(window)
	steps(test, window, click(window, "Eliminar"))
	settle(window)
	if _, found := productNamed(test, program, "Rana feliz"); !found {
		test.Fatalf("the product went before the question was answered")
	}
	steps(test, window, clickIn(window, "¿Eliminar «Rana feliz»?", "Eliminar"))
	settle(window)
	if _, found := productNamed(test, program, "Rana feliz"); found {
		test.Errorf("the product is still there after confirming")
	}
}

func TestATagIsAddedAndRemovedFromAProduct(test *testing.T) {
	program, window := openProductEditor(test, "Rana feliz")
	steps(test, window, typeInto(window, "Añade una etiqueta…", "verde"))
	window.PressKey(input.KeyEnter)
	settle(window)
	steps(test, window, click(window, "Guardar"))
	settle(window)
	saved, _ := productNamed(test, program, "Rana feliz")
	if !hasTag(saved, "verde") {
		test.Fatalf("the product's tags are %v, without «verde»", saved.Tags)
	}
	steps(test, window, click(window, "Rana feliz"))
	settle(window)
	steps(test, window, click(window, "Quitar verde"), click(window, "Guardar"))
	settle(window)
	if removed, _ := productNamed(test, program, "Rana feliz"); hasTag(removed, "verde") {
		test.Errorf("«verde» is still a tag after removing it")
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

func TestASecondVersionNeedsNamesForBoth(test *testing.T) {
	program, window := openProductEditor(test, "Rana feliz")
	steps(test, window, click(window, "Añadir otra versión"))
	settle(window)
	steps(test, window, click(window, "Guardar"))
	settle(window)
	if saved, _ := productNamed(test, program, "Rana feliz"); len(saved.Variants) != 1 {
		test.Fatalf("two versions without names were saved")
	}
	if !slices.Contains(window.Texts(), "Con varias versiones, cada una necesita nombre") {
		test.Fatalf("the editor does not say why it did not save")
	}
	names := window.Context().ScreenReaderNodes()
	var fields []ui.AccessibleNode
	for _, node := range names {
		if node.Label == "Única" && node.Pressable {
			fields = append(fields, node)
		}
	}
	if len(fields) != 2 {
		test.Fatalf("%d version name fields, not 2", len(fields))
	}
	for index, words := range []string{"Verde", "Azul"} {
		field := fields[index].Area
		window.ClickAt(field.X+field.Width/2, field.Y+field.Height/2)
		window.Type(words)
	}
	steps(test, window, click(window, "Guardar"))
	settle(window)
	if saved, _ := productNamed(test, program, "Rana feliz"); len(saved.Variants) != 2 {
		test.Errorf("the product has %d versions after naming both", len(saved.Variants))
	}
}

func TestACommaAddsEachTagTypedBeforeIt(test *testing.T) {
	_, window := openProductEditor(test, "Rana feliz")
	steps(test, window, typeInto(window, "Añade una etiqueta…", "verde, otoño,"))
	settle(window)
	for _, name := range []string{"Quitar verde", "Quitar otoño"} {
		if _, err := window.Element(name); err != nil {
			test.Errorf("the comma did not add the tag: %v", err)
		}
	}
}

func TestASuggestionAddsTheTagOfAnotherProduct(test *testing.T) {
	_, window := openProductEditor(test, "Rana feliz")
	steps(test, window, typeInto(window, "Añade una etiqueta…", "lám"))
	settle(window)
	steps(test, window, click(window, "Lámina"))
	settle(window)
	if _, err := window.Element("Quitar Lámina"); err != nil {
		test.Errorf("the suggestion did not add its tag: %v", err)
	}
}

func TestTappingATagsNameDoesNotRemoveIt(test *testing.T) {
	_, window := openProductEditor(test, "Rana feliz")
	if err := window.Click("Llavero"); err == nil {
		test.Errorf("a tag's name can be tapped: only its × removes it")
	}
	if _, err := window.Element("Quitar Llavero"); err != nil {
		test.Errorf("the tag went: %v", err)
	}
}

// tagNamed is the stored tag with a name.
func tagNamed(test *testing.T, program *App, name string) catalogue.Tag {
	test.Helper()
	tags, err := program.stores.Products.Tags()
	if err != nil {
		test.Fatal(err)
	}
	for _, tag := range tags {
		if tag.Name == name {
			return tag
		}
	}
	test.Fatalf("no tag called «%s»", name)
	return catalogue.Tag{}
}

func TestAnIconChosenForATagIsKeptAndCanBeForgotten(test *testing.T) {
	program, window := openApp(test, true)
	program.Open(Catalogue)
	settle(window)
	steps(test, window, click(window, "Etiquetas"))
	settle(window)
	steps(test, window, click(window, "Cambiar el icono de Pegatinas"))
	settle(window)
	steps(test, window, click(window, "Icono sparkles"))
	settle(window)
	if icon := tagNamed(test, program, "Pegatina").Icon(); icon != "sparkles" {
		test.Fatalf("the tag's icon is %q after choosing «sparkles»", icon)
	}
	steps(test, window, click(window, "Volver al icono automático"))
	settle(window)
	if icon := tagNamed(test, program, "Pegatina").Icon(); icon != catalogue.TagIcon("Pegatina") {
		test.Errorf("the tag's icon is %q after going back to the guessed one", icon)
	}
}
