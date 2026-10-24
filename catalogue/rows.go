package catalogue

import (
	"context"

	"github.com/Itros97/MUSA/storage"
)

// readProducts reads the products that are not deleted, or only the one with onlyID when it is not
// empty, with their variants and tags.
func readProducts(rows storage.Querier, onlyID string) ([]Product, error) {
	products, order, err := readProductRows(rows, onlyID)
	if err != nil {
		return nil, err
	}
	if err := readVariants(rows, products); err != nil {
		return nil, err
	}
	if err := readProductTags(rows, products); err != nil {
		return nil, err
	}
	all := make([]Product, 0, len(order))
	for _, id := range order {
		all = append(all, *products[id])
	}
	return all, nil
}

func readProductRows(rows storage.Querier, onlyID string) (map[string]*Product, []string, error) {
	result, err := rows.QueryContext(context.Background(), `
		SELECT p.id, p.kind, p.name, p.description, p.status, p.created_at, p.updated_at,
			coalesce(pp.material, ''), coalesce(dp.delivery_format, ''), coalesce(dp.turnaround_days, 0)
		FROM product p
		LEFT JOIN physical_product pp ON pp.product_id = p.id
		LEFT JOIN digital_product dp ON dp.product_id = p.id
		WHERE p.deleted_at IS NULL AND (? = '' OR p.id = ?)
		ORDER BY p.created_at, p.rowid`, onlyID, onlyID)
	if err != nil {
		return nil, nil, storageFailed("read products", err)
	}
	defer result.Close()
	products := map[string]*Product{}
	var order []string
	for result.Next() {
		var product Product
		var created, updated int64
		err := result.Scan(&product.ID, &product.Kind, &product.Name, &product.Description, &product.Status,
			&created, &updated, &product.Material, &product.DeliveryFormat, &product.TurnaroundDays)
		if err != nil {
			return nil, nil, storageFailed("read products", err)
		}
		product.CreatedAt, product.UpdatedAt = storage.Moment(created), storage.Moment(updated)
		products[product.ID] = &product
		order = append(order, product.ID)
	}
	return products, order, result.Err()
}

func readVariants(rows storage.Querier, products map[string]*Product) error {
	result, err := rows.QueryContext(context.Background(), `
		SELECT v.product_id, v.id, v.name, v.price_cents, v.cost_cents,
			coalesce(pv.sku, ''), coalesce(pv.stock, 0), coalesce(pv.min_stock, 0), coalesce(pv.weight_grams, 0),
			coalesce(pv.width_mm, 0), coalesce(pv.height_mm, 0), coalesce(pv.depth_mm, 0)
		FROM product_variant v
		LEFT JOIN physical_variant pv ON pv.variant_id = v.id
		WHERE v.deleted_at IS NULL
		ORDER BY v.position`)
	if err != nil {
		return storageFailed("read variants", err)
	}
	defer result.Close()
	for result.Next() {
		var productID string
		var variant Variant
		err := result.Scan(&productID, &variant.ID, &variant.Name, &variant.Price, &variant.Cost,
			&variant.SKU, &variant.Stock, &variant.MinimumStock, &variant.WeightGrams,
			&variant.WidthMillimetres, &variant.HeightMillimetres, &variant.DepthMillimetres)
		if err != nil {
			return storageFailed("read variants", err)
		}
		if product, found := products[productID]; found {
			product.Variants = append(product.Variants, variant)
		}
	}
	return result.Err()
}

func readProductTags(rows storage.Querier, products map[string]*Product) error {
	result, err := rows.QueryContext(context.Background(), `
		SELECT pt.product_id, t.id, t.name, t.color, t.icon
		FROM product_tag pt JOIN tag t ON t.id = pt.tag_id
		WHERE t.deleted_at IS NULL
		ORDER BY pt.position`)
	if err != nil {
		return storageFailed("read tags", err)
	}
	defer result.Close()
	for result.Next() {
		var productID string
		var tag Tag
		if err := result.Scan(&productID, &tag.ID, &tag.Name, &tag.Color, &tag.ChosenIcon); err != nil {
			return storageFailed("read tags", err)
		}
		if product, found := products[productID]; found {
			product.Tags = append(product.Tags, tag)
		}
	}
	return result.Err()
}

// writeProduct inserts or updates a product with its details, variants and tag links; variants it
// no longer has are marked deleted.
func writeProduct(rows storage.Querier, product Product) error {
	if err := writeProductRow(rows, product); err != nil {
		return err
	}
	updated := storage.Milliseconds(product.UpdatedAt)
	for position, variant := range product.Variants {
		if err := writeVariant(rows, product, variant, position); err != nil {
			return err
		}
	}
	statements := []statement{
		{"UPDATE product_variant SET deleted_at = ?1, updated_at = ?1 WHERE product_id = ?2 AND deleted_at IS NULL" +
			" AND id NOT IN (SELECT value FROM json_each(?3))", []any{updated, product.ID, variantIDsJSON(product.Variants)}},
		{"DELETE FROM product_tag WHERE product_id = ?", []any{product.ID}},
	}
	for position, tag := range product.Tags {
		statements = append(statements, statement{
			"INSERT INTO product_tag(product_id, tag_id, position) VALUES (?, ?, ?)", []any{product.ID, tag.ID, position}})
	}
	return execute(rows, "write product", statements...)
}

func writeProductRow(rows storage.Querier, product Product) error {
	created, updated := storage.Milliseconds(product.CreatedAt), storage.Milliseconds(product.UpdatedAt)
	details := statement{"INSERT OR REPLACE INTO physical_product(product_id, material) VALUES (?, ?)",
		[]any{product.ID, product.Material}}
	if product.Kind == Digital {
		details = statement{"INSERT OR REPLACE INTO digital_product(product_id, delivery_format, turnaround_days) VALUES (?, ?, ?)",
			[]any{product.ID, product.DeliveryFormat, zeroIsNull(product.TurnaroundDays)}}
	}
	return execute(rows, "write product",
		statement{`INSERT INTO product(id, kind, name, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET name = excluded.name, description = excluded.description,
			status = excluded.status, updated_at = excluded.updated_at`,
			[]any{product.ID, product.Kind, product.Name, product.Description, product.Status, created, updated}},
		details)
}

func writeVariant(rows storage.Querier, product Product, variant Variant, position int) error {
	updated := storage.Milliseconds(product.UpdatedAt)
	statements := []statement{{`INSERT INTO product_variant(id, product_id, name, price_cents, cost_cents, position, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, price_cents = excluded.price_cents, cost_cents = excluded.cost_cents,
		position = excluded.position, updated_at = excluded.updated_at, deleted_at = NULL`,
		[]any{variant.ID, product.ID, variant.Name, variant.Price, variant.Cost, position, updated}}}
	if product.Kind == Physical {
		statements = append(statements, statement{`INSERT OR REPLACE INTO physical_variant
			(variant_id, sku, stock, min_stock, weight_grams, width_mm, height_mm, depth_mm) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			[]any{variant.ID, emptyIsNull(variant.SKU), variant.Stock, zeroIsNull(variant.MinimumStock), zeroIsNull(variant.WeightGrams),
				zeroIsNull(variant.WidthMillimetres), zeroIsNull(variant.HeightMillimetres), zeroIsNull(variant.DepthMillimetres)}})
	}
	return execute(rows, "write variant", statements...)
}

func variantIDsJSON(variants []Variant) string {
	ids := make([]string, len(variants))
	for index, variant := range variants {
		ids[index] = variant.ID
	}
	return jsonList(ids)
}
