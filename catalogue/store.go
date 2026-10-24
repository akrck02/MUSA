package catalogue

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Itros97/MUSA/storage"
)

// Store keeps the catalogue in the database. Deleted products are never returned.
type Store struct {
	database *storage.Database
	now      func() time.Time
	newID    func() string
}

// NewStore keeps the catalogue in database, telling the time with now and making ids with newID;
// tests pass fixed ones.
func NewStore(database *storage.Database, now func() time.Time, newID func() string) *Store {
	return &Store{database: database, now: now, newID: newID}
}

// All reads every product that is not deleted, with its variants and tags.
func (store *Store) All() ([]Product, error) {
	return readProducts(store.database.Rows(), "")
}

// Find reads one product; found is false when it does not exist or was deleted.
func (store *Store) Find(id string) (product Product, found bool, err error) {
	products, err := readProducts(store.database.Rows(), id)
	if err != nil || len(products) == 0 {
		return Product{}, false, err
	}
	return products[0], true, nil
}

// FindByVariant reads the product one of whose variants is variantID.
func (store *Store) FindByVariant(variantID string) (product Product, found bool, err error) {
	var productID string
	err = store.database.Rows().QueryRowContext(context.Background(),
		"SELECT product_id FROM product_variant WHERE id = ? AND deleted_at IS NULL", variantID).Scan(&productID)
	if errors.Is(err, sql.ErrNoRows) {
		return Product{}, false, nil
	}
	if err != nil {
		return Product{}, false, storageFailed("find by variant", err)
	}
	return store.Find(productID)
}

// Stock is the units in stock of a physical variant; found is false when it is not physical or was deleted.
func (store *Store) Stock(variantID string) (stock int, found bool, err error) {
	return StockIn(store.database.Rows(), variantID)
}

// StockIn is Stock, read inside a transaction other features are running.
func StockIn(rows storage.Querier, variantID string) (stock int, found bool, err error) {
	err = rows.QueryRowContext(context.Background(), `
		SELECT pv.stock FROM physical_variant pv
		JOIN product_variant v ON v.id = pv.variant_id
		JOIN product p ON p.id = v.product_id
		WHERE pv.variant_id = ? AND v.deleted_at IS NULL AND p.deleted_at IS NULL`, variantID).Scan(&stock)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, storageFailed("read stock", err)
	}
	return stock, true, nil
}

// Tags reads every tag, sorted by name.
func (store *Store) Tags() ([]Tag, error) {
	rows, err := store.database.Rows().QueryContext(context.Background(),
		"SELECT id, name, color, icon FROM tag WHERE deleted_at IS NULL ORDER BY normalized_name")
	if err != nil {
		return nil, storageFailed("read tags", err)
	}
	defer rows.Close()
	var tags []Tag
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.ChosenIcon); err != nil {
			return nil, storageFailed("read tags", err)
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// SetTagIcon gives the tag with this id an icon; an empty icon goes back to the one guessed from
// its name.
func (store *Store) SetTagIcon(id, icon string) error {
	return execute(store.database.Rows(), "set tag icon", statement{
		"UPDATE tag SET icon = ?, updated_at = ? WHERE id = ?", []any{icon, storage.Milliseconds(store.now()), id}})
}

// Save checks a draft and stores it as a product, making the tags it names that do not exist yet.
// It gives back the problems that stop it, if any, and the product as stored otherwise.
func (store *Store) Save(draft ProductDraft) (Product, []Problem, error) {
	product := trimmed(draft.Product)
	problems := Validate(product)
	taken, err := store.takenSKUProblems(product)
	if err != nil {
		return Product{}, nil, err
	}
	if problems = append(problems, taken...); len(problems) > 0 {
		return Product{}, problems, nil
	}
	now := store.now()
	existing, found, err := store.Find(product.ID)
	if err != nil {
		return Product{}, nil, err
	}
	product.CreatedAt, product.UpdatedAt = now, now
	if found {
		product.CreatedAt = existing.CreatedAt
		product.Variants = keepStockMovedElsewhere(product.Variants, existing, draft.StockBaseline)
	}
	err = store.database.InTransaction(func(rows storage.Querier) error {
		tags, err := store.resolveTags(rows, draft.TagNames, now)
		if err != nil {
			return err
		}
		product.Tags = tags
		return writeProduct(rows, product)
	})
	return product, nil, err
}

// Duplicate copies a product as a new draft named name, with new ids, no stock and no SKUs (they
// must be unique), and gives back the copy's id.
func (store *Store) Duplicate(id, name string) (string, error) {
	original, found, err := store.Find(id)
	if err != nil {
		return "", err
	}
	if !found {
		return "", productNotFound(id)
	}
	now := store.now()
	copied := original
	copied.ID, copied.Name, copied.Status = store.newID(), name, Draft
	copied.CreatedAt, copied.UpdatedAt = now, now
	copied.Variants = make([]Variant, len(original.Variants))
	for index, variant := range original.Variants {
		variant.ID, variant.SKU, variant.Stock = store.newID(), "", 0
		copied.Variants[index] = variant
	}
	err = store.database.InTransaction(func(rows storage.Querier) error { return writeProduct(rows, copied) })
	return copied.ID, err
}

// AdjustStock adds delta units to a variant's stock, never going below zero.
func (store *Store) AdjustStock(variantID string, delta int) error {
	return store.database.InTransaction(func(rows storage.Querier) error {
		return AdjustStockIn(rows, variantID, delta, store.now())
	})
}

// AdjustStockIn is AdjustStock inside a transaction other features are running: events take
// stock to a fair, sales take it from the catalogue.
func AdjustStockIn(rows storage.Querier, variantID string, delta int, at time.Time) error {
	var productID string
	err := rows.QueryRowContext(context.Background(), "SELECT product_id FROM product_variant WHERE id = ?", variantID).Scan(&productID)
	if errors.Is(err, sql.ErrNoRows) {
		return variantNotFound(variantID)
	}
	if err != nil {
		return storageFailed("adjust stock", err)
	}
	milliseconds := storage.Milliseconds(at)
	return execute(rows, "adjust stock",
		statement{"UPDATE physical_variant SET stock = max(0, stock + ?) WHERE variant_id = ?", []any{delta, variantID}},
		statement{"UPDATE product_variant SET updated_at = ? WHERE id = ?", []any{milliseconds, variantID}},
		statement{"UPDATE product SET updated_at = ? WHERE id = ?", []any{milliseconds, productID}},
	)
}

// IsPhysicalIn tells whether a variant belongs to a physical product, inside a transaction.
func IsPhysicalIn(rows storage.Querier, variantID string) (bool, error) {
	var count int
	err := rows.QueryRowContext(context.Background(), "SELECT count(*) FROM physical_variant WHERE variant_id = ?", variantID).Scan(&count)
	if err != nil {
		return false, storageFailed("check physical", err)
	}
	return count > 0, nil
}

// SetArchived archives a product, or brings it back as active.
func (store *Store) SetArchived(id string, archived bool) error {
	status := Active
	if archived {
		status = Archived
	}
	return execute(store.database.Rows(), "archive",
		statement{"UPDATE product SET status = ?, updated_at = ? WHERE id = ?", []any{status, storage.Milliseconds(store.now()), id}})
}

// Delete takes a product out of the catalogue. The row stays, marked deleted, so a future sync can
// pass the deletion on.
func (store *Store) Delete(id string) error {
	at := storage.Milliseconds(store.now())
	return execute(store.database.Rows(), "delete",
		statement{"UPDATE product SET deleted_at = ?, updated_at = ? WHERE id = ?", []any{at, at, id}})
}

func (store *Store) takenSKUProblems(product Product) ([]Problem, error) {
	if product.Kind != Physical {
		return nil, nil
	}
	var problems []Problem
	for _, variant := range product.Variants {
		if variant.SKU == "" {
			continue
		}
		var owner string
		err := store.database.Rows().QueryRowContext(context.Background(), `
			SELECT p.name FROM physical_variant pv
			JOIN product_variant v ON v.id = pv.variant_id
			JOIN product p ON p.id = v.product_id
			WHERE lower(pv.sku) = lower(?) AND p.id != ? AND v.deleted_at IS NULL AND p.deleted_at IS NULL
			LIMIT 1`, variant.SKU, product.ID).Scan(&owner)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, storageFailed("check sku", err)
		}
		problems = append(problems, Problem{Code: CodeDuplicateSKU, VariantID: variant.ID, OwnerName: owner})
	}
	return problems, nil
}

func (store *Store) resolveTags(rows storage.Querier, names []string, now time.Time) ([]Tag, error) {
	var tags []Tag
	for _, name := range DistinctTagNames(names) {
		key := Fold(name)
		tag := Tag{}
		err := rows.QueryRowContext(context.Background(),
			"SELECT id, name, color, icon FROM tag WHERE normalized_name = ? AND deleted_at IS NULL", key).Scan(&tag.ID, &tag.Name, &tag.Color, &tag.ChosenIcon)
		if errors.Is(err, sql.ErrNoRows) {
			tag = Tag{ID: store.newID(), Name: name, Color: TagColorFor(key)}
			at := storage.Milliseconds(now)
			err = execute(rows, "make tag", statement{
				"INSERT INTO tag(id, name, normalized_name, color, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
				[]any{tag.ID, tag.Name, key, tag.Color, at, at}})
		}
		if err != nil {
			return nil, storageFailed("resolve tag", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// keepStockMovedElsewhere applies only the editor's own change to each variant's stock: units
// may have gone to a fair or been sold while the editor was open.
func keepStockMovedElsewhere(variants []Variant, stored Product, baseline map[string]int) []Variant {
	kept := make([]Variant, len(variants))
	for index, variant := range variants {
		before, hasBaseline := baseline[variant.ID]
		current, isStored := stored.Variant(variant.ID)
		if hasBaseline && isStored {
			variant.Stock = max(0, current.Stock+variant.Stock-before)
		}
		kept[index] = variant
	}
	return kept
}

func trimmed(product Product) Product {
	product.Name = strings.TrimSpace(product.Name)
	product.Description = strings.TrimSpace(product.Description)
	product.Material = strings.TrimSpace(product.Material)
	product.DeliveryFormat = strings.TrimSpace(product.DeliveryFormat)
	variants := make([]Variant, len(product.Variants))
	for index, variant := range product.Variants {
		variant.Name = strings.TrimSpace(variant.Name)
		variant.SKU = strings.TrimSpace(variant.SKU)
		variants[index] = variant
	}
	product.Variants = variants
	return product
}

// statement is one SQL statement and its arguments.
type statement struct {
	query     string
	arguments []any
}

func execute(rows storage.Querier, doing string, statements ...statement) error {
	for _, each := range statements {
		if _, err := rows.ExecContext(context.Background(), each.query, each.arguments...); err != nil {
			return storageFailed(doing, err)
		}
	}
	return nil
}

// zeroIsNull stores "not said" (zero) as NULL, as the Kotlin schema did.
func zeroIsNull(value int) any {
	if value == 0 {
		return nil
	}
	return value
}

func emptyIsNull(value string) any {
	if value == "" {
		return nil
	}
	return value
}
