package sales

import (
	"context"
	"strings"
	"time"

	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/storage"
)

// Store keeps the sales in the database. Deleted sales are never returned.
type Store struct {
	database  *storage.Database
	catalogue *catalogue.Store
	now       func() time.Time
	newID     func() string
}

// NewStore keeps sales in database, finding what they sell in products; tests pass a fixed clock
// and fixed ids.
func NewStore(database *storage.Database, products *catalogue.Store, now func() time.Time, newID func() string) *Store {
	return &Store{database: database, catalogue: products, now: now, newID: newID}
}

// All reads every sale with its lines, newest first.
func (store *Store) All() ([]Sale, error) { return readSales(store.database.Rows(), "", "") }

// OfEvent reads the sales made at one fair, newest first.
func (store *Store) OfEvent(eventID string) ([]Sale, error) {
	return readSales(store.database.Rows(), eventID, "")
}

// Find reads one sale; found is false when it does not exist or was deleted.
func (store *Store) Find(id string) (sale Sale, found bool, err error) {
	return FindIn(store.database.Rows(), id)
}

// FindIn is Find inside a transaction other features are running.
func FindIn(rows storage.Querier, id string) (sale Sale, found bool, err error) {
	sales, err := readSales(rows, "", id)
	if err != nil || len(sales) == 0 {
		return Sale{}, false, err
	}
	return sales[0], true, nil
}

// Record checks a draft and records it as a sale outside fairs: names and costs are copied from
// the catalogue, and its physical units leave it. It gives back the problems that stop it, if any.
func (store *Store) Record(draft Draft) (Sale, []Problem, error) {
	if problems := Validate(draft); len(problems) > 0 {
		return Sale{}, problems, nil
	}
	products, err := store.productsOf(draft.Lines)
	if err != nil {
		return Sale{}, nil, err
	}
	if problems := stockProblems(draft.Lines, products); len(problems) > 0 {
		return Sale{}, problems, nil
	}
	sale := Sale{ID: store.newID(), Channel: draft.Channel, Status: StatusOfNew(draft.Channel), PaidWith: draft.PaidWith, Customer: strings.TrimSpace(draft.Customer),
		Note: strings.TrimSpace(draft.Note), ShippingCharged: draft.ShippingCharged, ShippingCost: draft.ShippingCost, SoldAt: draft.SoldAt}
	for _, line := range draft.Lines {
		product := products[line.VariantID]
		variant, _ := product.Variant(line.VariantID) // productsOf found it
		sale.Lines = append(sale.Lines, Line{ID: line.ID, VariantID: line.VariantID, ProductName: product.Name,
			VariantName: variant.Name, Quantity: line.Quantity, UnitPrice: line.UnitPrice, UnitCost: variant.Cost})
	}
	err = store.database.InTransaction(func(rows storage.Querier) error { return InsertIn(rows, sale) })
	return sale, nil, err
}

// Delete deletes a sale made outside fairs and gives its physical units back; a fair's sales are
// undone from their event.
func (store *Store) Delete(id string) ([]Problem, error) {
	sale, found, err := store.Find(id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, saleNotFound(id)
	}
	if sale.EventID != "" {
		return []Problem{{Code: CodeManagedByEvent}}, nil
	}
	return nil, store.database.InTransaction(func(rows storage.Querier) error { return DeleteIn(rows, sale, store.now()) })
}

// Advance moves an order to its next status: to prepare, sent, delivered.
func (store *Store) Advance(id string) ([]Problem, error) {
	sale, found, err := store.Find(id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, saleNotFound(id)
	}
	next, found := sale.Status.Next()
	if !found {
		return []Problem{{Code: CodeAlreadyDelivered}}, nil
	}
	at := storage.Milliseconds(store.now())
	_, err = store.database.Rows().ExecContext(context.Background(), "UPDATE sale SET status = ?, updated_at = ? WHERE id = ?",
		next, at, id)
	if err != nil {
		return nil, storageFailed("advance order", err)
	}
	return nil, nil
}

// InsertIn writes a sale and its lines inside a transaction; outside fairs its physical units
// leave the catalogue. A sale with no status takes the one a new order of its channel starts with.
func InsertIn(rows storage.Querier, sale Sale) error {
	if sale.Status == "" {
		sale.Status = StatusOfNew(sale.Channel)
	}
	at := storage.Milliseconds(sale.SoldAt)
	_, err := rows.ExecContext(context.Background(), `INSERT INTO sale(id, channel, status, paid_with, event_id, customer, note,
		shipping_charged_cents, shipping_cost_cents, sold_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sale.ID, sale.Channel, sale.Status, sale.PaidWith, emptyIsNull(sale.EventID), sale.Customer, sale.Note, sale.ShippingCharged, sale.ShippingCost, at, at)
	if err != nil {
		return storageFailed("record sale", err)
	}
	for position, line := range sale.Lines {
		_, err := rows.ExecContext(context.Background(), `INSERT INTO sale_line(id, sale_id, variant_id, product_name,
			variant_name, quantity, unit_price_cents, unit_cost_cents, position) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			line.ID, sale.ID, line.VariantID, line.ProductName, line.VariantName, line.Quantity, line.UnitPrice, line.UnitCost, position)
		if err != nil {
			return storageFailed("record sale line", err)
		}
	}
	return moveCatalogueStock(rows, sale, -1, sale.SoldAt)
}

// DeleteIn marks a sale deleted inside a transaction; outside fairs its physical units go back.
func DeleteIn(rows storage.Querier, sale Sale, at time.Time) error {
	milliseconds := storage.Milliseconds(at)
	_, err := rows.ExecContext(context.Background(), "UPDATE sale SET deleted_at = ?, updated_at = ? WHERE id = ?",
		milliseconds, milliseconds, sale.ID)
	if err != nil {
		return storageFailed("delete sale", err)
	}
	return moveCatalogueStock(rows, sale, 1, at)
}

// moveCatalogueStock leaves a fair's sales alone: their units left the catalogue when they were
// taken to the fair.
func moveCatalogueStock(rows storage.Querier, sale Sale, direction int, at time.Time) error {
	if sale.EventID != "" {
		return nil
	}
	for _, line := range sale.Lines {
		physical, err := catalogue.IsPhysicalIn(rows, line.VariantID)
		if err != nil {
			return err
		}
		if !physical {
			continue
		}
		if err := catalogue.AdjustStockIn(rows, line.VariantID, direction*line.Quantity, at); err != nil {
			return err
		}
	}
	return nil
}

func (store *Store) productsOf(lines []LineDraft) (map[string]catalogue.Product, error) {
	products := map[string]catalogue.Product{}
	for _, line := range lines {
		product, found, err := store.catalogue.FindByVariant(line.VariantID)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, variantNotFound(line.VariantID)
		}
		products[line.VariantID] = product
	}
	return products, nil
}

// stockProblems adds up the lines that sell the same physical variant against its stock.
func stockProblems(lines []LineDraft, products map[string]catalogue.Product) []Problem {
	wanted := map[string]int{}
	var order []string
	for _, line := range lines {
		if wanted[line.VariantID] == 0 {
			order = append(order, line.VariantID)
		}
		wanted[line.VariantID] += line.Quantity
	}
	var problems []Problem
	for _, variantID := range order {
		product := products[variantID]
		variant, _ := product.Variant(variantID) // productsOf found it
		if product.Kind == catalogue.Physical && wanted[variantID] > variant.Stock {
			problems = append(problems, Problem{Code: CodeNotEnoughStock, VariantID: variantID, Available: variant.Stock})
		}
	}
	return problems
}

func readSales(rows storage.Querier, eventID, onlyID string) ([]Sale, error) {
	result, err := rows.QueryContext(context.Background(), `
		SELECT id, channel, status, paid_with, coalesce(event_id, ''), customer, note, shipping_charged_cents, shipping_cost_cents, sold_at
		FROM sale WHERE deleted_at IS NULL AND (?1 = '' OR event_id = ?1) AND (?2 = '' OR id = ?2)
		ORDER BY sold_at DESC`, eventID, onlyID)
	if err != nil {
		return nil, storageFailed("read sales", err)
	}
	var sales []Sale
	for result.Next() {
		var sale Sale
		var soldAt int64
		err := result.Scan(&sale.ID, &sale.Channel, &sale.Status, &sale.PaidWith, &sale.EventID, &sale.Customer, &sale.Note,
			&sale.ShippingCharged, &sale.ShippingCost, &soldAt)
		if err != nil {
			result.Close() // the scan error is the one worth telling
			return nil, storageFailed("read sales", err)
		}
		sale.SoldAt = storage.Moment(soldAt)
		sales = append(sales, sale)
	}
	result.Close() // all read: closing frees the one connection for the lines
	if err := result.Err(); err != nil {
		return nil, storageFailed("read sales", err)
	}
	for index := range sales {
		lines, err := readLines(rows, sales[index].ID)
		if err != nil {
			return nil, err
		}
		sales[index].Lines = lines
	}
	return sales, nil
}

func readLines(rows storage.Querier, saleID string) ([]Line, error) {
	result, err := rows.QueryContext(context.Background(), `
		SELECT id, variant_id, product_name, variant_name, quantity, unit_price_cents, unit_cost_cents
		FROM sale_line WHERE sale_id = ? ORDER BY position`, saleID)
	if err != nil {
		return nil, storageFailed("read sale lines", err)
	}
	defer result.Close()
	var lines []Line
	for result.Next() {
		var line Line
		err := result.Scan(&line.ID, &line.VariantID, &line.ProductName, &line.VariantName, &line.Quantity, &line.UnitPrice, &line.UnitCost)
		if err != nil {
			return nil, storageFailed("read sale lines", err)
		}
		lines = append(lines, line)
	}
	return lines, result.Err()
}

func emptyIsNull(value string) any {
	if value == "" {
		return nil
	}
	return value
}
