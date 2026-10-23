package events

import (
	"context"
	"strings"
	"time"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/money"
	"github.com/Itros97/MUSA/sales"
	"github.com/Itros97/MUSA/storage"
)

// Store keeps the fairs in the database; moving stock changes the catalogue in the same
// transaction. Deleted fairs are never returned.
type Store struct {
	database *storage.Database
	sales    *sales.Store
	now      func() time.Time
	newID    func() string
}

// NewStore keeps fairs in database, beside the sales they make; tests pass a fixed clock and ids.
func NewStore(database *storage.Database, sold *sales.Store, now func() time.Time, newID func() string) *Store {
	return &Store{database: database, sales: sold, now: now, newID: newID}
}

// All reads every fair: open ones first, then the most recent first.
func (store *Store) All() ([]Event, error) { return readEvents(store.database.Rows(), "") }

// Find reads one fair; found is false when it does not exist or was deleted.
func (store *Store) Find(id string) (event Event, found bool, err error) {
	events, err := readEvents(store.database.Rows(), id)
	if err != nil || len(events) == 0 {
		return Event{}, false, err
	}
	return events[0], true, nil
}

// Detail reads a fair with its stock, sales and expenses.
func (store *Store) Detail(id string) (Detail, error) {
	event, found, err := store.Find(id)
	if err != nil {
		return Detail{}, err
	}
	if !found {
		return Detail{}, eventNotFound(id)
	}
	detail := Detail{Event: event}
	if detail.Stock, err = readStock(store.database.Rows(), id); err != nil {
		return Detail{}, err
	}
	if detail.Sales, err = store.sales.OfEvent(id); err != nil {
		return Detail{}, err
	}
	detail.Expenses, err = readExpenses(store.database.Rows(), id)
	return detail, err
}

// Save checks a draft and stores it; a new fair starts open.
func (store *Store) Save(draft Draft) ([]Refusal, error) {
	draft.Name, draft.Location = strings.TrimSpace(draft.Name), strings.TrimSpace(draft.Location)
	if problems := Validate(draft); len(problems) > 0 {
		return refusals(problems...), nil
	}
	now := store.now()
	existing, found, err := store.Find(draft.ID)
	if err != nil {
		return nil, err
	}
	event := Event{ID: draft.ID, Name: draft.Name, Location: draft.Location, StartDate: draft.StartDate,
		EndDate: draft.EndDate, Status: Open, CreatedAt: now, UpdatedAt: now}
	if found {
		event.Status, event.CreatedAt = existing.Status, existing.CreatedAt
	}
	return nil, execute(store.database.Rows(), "save event", statement{
		`INSERT OR REPLACE INTO event(id, name, location, start_date, end_date, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		[]any{event.ID, event.Name, event.Location, event.StartDate.String(), event.EndDate.String(), event.Status,
			storage.Milliseconds(event.CreatedAt), storage.Milliseconds(event.UpdatedAt)}})
}

// MoveStock takes quantity units of a physical variant from the catalogue to an open fair.
func (store *Store) MoveStock(eventID, variantID string, quantity int) ([]Refusal, error) {
	if quantity <= 0 {
		return refusals(CodeInvalidQuantity), nil
	}
	if _, refused, err := store.openDetail(eventID); err != nil || refused != nil {
		return refused, err
	}
	available, found, err := catalogue.StockIn(store.database.Rows(), variantID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, notTaken(variantID)
	}
	if quantity > available {
		return []Refusal{{Code: CodeNotEnoughStock, Available: available}}, nil
	}
	return nil, store.database.InTransaction(func(rows storage.Querier) error {
		return moveIn(rows, eventID, variantID, quantity, store.now())
	})
}

// ReturnStock brings quantity unsold units of a variant back from an open fair to the catalogue.
func (store *Store) ReturnStock(eventID, variantID string, quantity int) ([]Refusal, error) {
	if quantity <= 0 {
		return refusals(CodeInvalidQuantity), nil
	}
	detail, refused, err := store.openDetail(eventID)
	if err != nil || refused != nil {
		return refused, err
	}
	line, _ := detail.LineOf(variantID) // a variant not taken has nothing remaining
	if quantity > line.Remaining() {
		return []Refusal{{Code: CodeExceedsRemaining, Available: line.Remaining()}}, nil
	}
	return nil, store.database.InTransaction(func(rows storage.Querier) error {
		return moveIn(rows, eventID, variantID, -quantity, store.now())
	})
}

// Sell records quantity units of a variant on an open fair's stand sold at unitPrice, as a sale of
// one line; the cost is copied from the variant as it is now.
func (store *Store) Sell(eventID, variantID string, quantity int, unitPrice money.Money) ([]Refusal, error) {
	var codes []Code
	if quantity <= 0 {
		codes = append(codes, CodeInvalidQuantity)
	}
	if unitPrice < 0 {
		codes = append(codes, CodeNegativePrice)
	}
	if len(codes) > 0 {
		return refusals(codes...), nil
	}
	detail, refused, err := store.openDetail(eventID)
	if err != nil || refused != nil {
		return refused, err
	}
	line, found := detail.LineOf(variantID)
	if !found {
		return nil, notTaken(variantID)
	}
	if quantity > line.Remaining() {
		return []Refusal{{Code: CodeExceedsRemaining, Available: line.Remaining()}}, nil
	}
	sale := sales.Sale{ID: store.newID(), Channel: sales.AtEvent, EventID: eventID, SoldAt: store.now(),
		Lines: []sales.Line{{ID: store.newID(), VariantID: variantID, ProductName: line.ProductName,
			VariantName: line.VariantName, Quantity: quantity, UnitPrice: unitPrice, UnitCost: line.Cost}}}
	return nil, store.database.InTransaction(func(rows storage.Querier) error { return sales.InsertIn(rows, sale) })
}

// Undo deletes one of a fair's sales, so its units are on the stand again; a closed fair keeps them.
func (store *Store) Undo(saleID string) ([]Refusal, error) {
	sale, found, err := sales.FindIn(store.database.Rows(), saleID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, saleNotFound(saleID)
	}
	if event, found, err := store.Find(sale.EventID); err != nil || (found && event.Status == Closed) {
		return refusalIfClosed(err)
	}
	return nil, store.database.InTransaction(func(rows storage.Querier) error { return sales.DeleteIn(rows, sale, store.now()) })
}

// Close finishes a fair: unsold units go back to the catalogue and it becomes read-only.
func (store *Store) Close(id string) ([]Refusal, error) {
	return store.finish(id, "UPDATE event SET status = 'CLOSED', updated_at = ?1 WHERE id = ?2")
}

// Delete deletes a fair without sales, bringing its stock back; a fair with sales is closed instead.
func (store *Store) Delete(id string) ([]Refusal, error) {
	detail, err := store.Detail(id)
	if err != nil {
		return nil, err
	}
	if len(detail.Sales) > 0 {
		return refusals(CodeHasSales), nil
	}
	return nil, store.database.InTransaction(func(rows storage.Querier) error {
		return store.giveBackAndRun(rows, detail, "UPDATE event SET deleted_at = ?1, updated_at = ?1 WHERE id = ?2")
	})
}

// AddExpense records what attending an open fair cost.
func (store *Store) AddExpense(eventID, concept string, amount money.Money) ([]Refusal, error) {
	var codes []Code
	if strings.TrimSpace(concept) == "" {
		codes = append(codes, CodeBlankConcept)
	}
	if amount < 0 {
		codes = append(codes, CodeNegativeAmount)
	}
	if len(codes) > 0 {
		return refusals(codes...), nil
	}
	if _, refused, err := store.openDetail(eventID); err != nil || refused != nil {
		return refused, err
	}
	at := storage.Milliseconds(store.now())
	return nil, execute(store.database.Rows(), "add expense", statement{
		"INSERT INTO event_expense(id, event_id, concept, amount_cents, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		[]any{store.newID(), eventID, strings.TrimSpace(concept), amount, at, at}})
}

// DeleteExpense takes an expense off an open fair.
func (store *Store) DeleteExpense(eventID, expenseID string) ([]Refusal, error) {
	if _, refused, err := store.openDetail(eventID); err != nil || refused != nil {
		return refused, err
	}
	at := storage.Milliseconds(store.now())
	return nil, execute(store.database.Rows(), "delete expense", statement{
		"UPDATE event_expense SET deleted_at = ?, updated_at = ? WHERE id = ?", []any{at, at, expenseID}})
}

func (store *Store) finish(id, update string) ([]Refusal, error) {
	detail, refused, err := store.openDetail(id)
	if err != nil || refused != nil {
		return refused, err
	}
	return nil, store.database.InTransaction(func(rows storage.Querier) error {
		return store.giveBackAndRun(rows, detail, update)
	})
}

// giveBackAndRun brings every unsold unit back to the catalogue, then runs update on the fair's row.
func (store *Store) giveBackAndRun(rows storage.Querier, detail Detail, update string) error {
	now := store.now()
	for _, line := range detail.Stock {
		if line.Remaining() <= 0 {
			continue
		}
		if err := moveIn(rows, detail.Event.ID, line.VariantID, -line.Remaining(), now); err != nil {
			return err
		}
	}
	return execute(rows, "finish event", statement{update, []any{storage.Milliseconds(now), detail.Event.ID}})
}

// openDetail reads a fair that must be open; refused says so when it is closed.
func (store *Store) openDetail(id string) (detail Detail, refused []Refusal, err error) {
	detail, err = store.Detail(id)
	if err != nil {
		return Detail{}, nil, err
	}
	if detail.Event.Status != Open {
		return Detail{}, refusals(CodeEventClosed), nil
	}
	return detail, nil, nil
}

func refusalIfClosed(err error) ([]Refusal, error) {
	if err != nil {
		return nil, err
	}
	return refusals(CodeEventClosed), nil
}

// moveIn moves quantity units of a variant from the catalogue to a fair; a negative quantity
// moves them back.
func moveIn(rows storage.Querier, eventID, variantID string, quantity int, at time.Time) error {
	if err := catalogue.AdjustStockIn(rows, variantID, -quantity, at); err != nil {
		return err
	}
	return execute(rows, "move stock", statement{`INSERT INTO event_stock(event_id, variant_id, taken, updated_at) VALUES (?1, ?2, ?3, ?4)
		ON CONFLICT(event_id, variant_id) DO UPDATE SET taken = taken + ?3, updated_at = ?4`,
		[]any{eventID, variantID, quantity, storage.Milliseconds(at)}})
}

func readEvents(rows storage.Querier, onlyID string) ([]Event, error) {
	result, err := rows.QueryContext(context.Background(), `
		SELECT id, name, location, start_date, end_date, status, created_at, updated_at FROM event
		WHERE deleted_at IS NULL AND (?1 = '' OR id = ?1)
		ORDER BY status = 'CLOSED', start_date DESC, name`, onlyID)
	if err != nil {
		return nil, storageFailed("read events", err)
	}
	defer result.Close()
	var events []Event
	for result.Next() {
		var event Event
		var start, end string
		var created, updated int64
		if err := result.Scan(&event.ID, &event.Name, &event.Location, &start, &end, &event.Status, &created, &updated); err != nil {
			return nil, storageFailed("read events", err)
		}
		if event.StartDate, err = calendar.Parse(start); err != nil {
			return nil, storageFailed("read events", err)
		}
		if event.EndDate, err = calendar.Parse(end); err != nil {
			return nil, storageFailed("read events", err)
		}
		event.CreatedAt, event.UpdatedAt = storage.Moment(created), storage.Moment(updated)
		events = append(events, event)
	}
	return events, result.Err()
}

// readStock keeps lines with nothing taken as rows, so a future sync sees the change, but does not
// list them.
func readStock(rows storage.Querier, eventID string) ([]StockLine, error) {
	result, err := rows.QueryContext(context.Background(), `
		SELECT es.variant_id, p.name, v.name, v.price_cents, v.cost_cents, es.taken,
			(SELECT coalesce(sum(sl.quantity), 0) FROM sale_line sl JOIN sale s ON s.id = sl.sale_id
			 WHERE s.event_id = es.event_id AND sl.variant_id = es.variant_id AND s.deleted_at IS NULL)
		FROM event_stock es
		JOIN product_variant v ON v.id = es.variant_id
		JOIN product p ON p.id = v.product_id
		WHERE es.event_id = ? AND es.taken > 0
		ORDER BY p.name, v.position`, eventID)
	if err != nil {
		return nil, storageFailed("read event stock", err)
	}
	defer result.Close()
	var lines []StockLine
	for result.Next() {
		var line StockLine
		if err := result.Scan(&line.VariantID, &line.ProductName, &line.VariantName, &line.Price, &line.Cost, &line.Taken, &line.Sold); err != nil {
			return nil, storageFailed("read event stock", err)
		}
		lines = append(lines, line)
	}
	return lines, result.Err()
}

func readExpenses(rows storage.Querier, eventID string) ([]Expense, error) {
	result, err := rows.QueryContext(context.Background(), `
		SELECT id, event_id, concept, amount_cents, created_at FROM event_expense
		WHERE event_id = ? AND deleted_at IS NULL ORDER BY created_at`, eventID)
	if err != nil {
		return nil, storageFailed("read expenses", err)
	}
	defer result.Close()
	var expenses []Expense
	for result.Next() {
		var expense Expense
		var created int64
		if err := result.Scan(&expense.ID, &expense.EventID, &expense.Concept, &expense.Amount, &created); err != nil {
			return nil, storageFailed("read expenses", err)
		}
		expense.CreatedAt = storage.Moment(created)
		expenses = append(expenses, expense)
	}
	return expenses, result.Err()
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
