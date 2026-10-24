package projects

import (
	"context"
	"strings"
	"time"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/catalogue"
	"github.com/Itros97/MUSA/customers"
	"github.com/Itros97/MUSA/sales"
	"github.com/Itros97/MUSA/storage"
)

// Store keeps the projects in the database. Deleted projects are never returned.
type Store struct {
	database  *storage.Database
	catalogue *catalogue.Store
	customers *customers.Store
	now       func() time.Time
	newID     func() string
}

// NewStore keeps projects in database, finding their products in products and their customers in
// people; tests pass a fixed clock and fixed ids.
func NewStore(database *storage.Database, products *catalogue.Store, people *customers.Store, now func() time.Time, newID func() string) *Store {
	return &Store{database: database, catalogue: products, customers: people, now: now, newID: newID}
}

// All reads every project, the soonest due first; those without a deadline last.
func (store *Store) All() ([]Project, error) { return readProjects(store.database.Rows(), "") }

// Find reads one project; found is false when it does not exist or was deleted.
func (store *Store) Find(id string) (project Project, found bool, err error) {
	projects, err := readProjects(store.database.Rows(), id)
	if err != nil || len(projects) == 0 {
		return Project{}, false, err
	}
	return projects[0], true, nil
}

// Save checks a project and adds it, or changes it; a new one with no stage starts to be done.
// It gives back the saved project, or the problems that stop it. Its stage is changed with Move.
func (store *Store) Save(project Project) (Project, []Code, error) {
	project.Title = strings.TrimSpace(project.Title)
	problems, err := store.problems(project)
	if err != nil || len(problems) > 0 {
		return Project{}, problems, err
	}
	if project.ID == "" {
		project.ID = store.newID()
	}
	if project.Stage == "" {
		project.Stage = ToStart
	}
	at := storage.Milliseconds(store.now())
	_, err = store.database.Rows().ExecContext(context.Background(), `INSERT INTO project(id, title, customer_id, stage, deadline,
		price_cents, variant_id, sale_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, '', ?, ?)
		ON CONFLICT(id) DO UPDATE SET title = excluded.title, customer_id = excluded.customer_id, deadline = excluded.deadline,
		price_cents = excluded.price_cents, variant_id = excluded.variant_id, updated_at = excluded.updated_at`,
		project.ID, project.Title, project.CustomerID, project.Stage, deadlineText(project.Deadline), project.Price, project.VariantID, at, at)
	if err != nil {
		return Project{}, nil, storageFailed("save project", err)
	}
	return project, nil, nil
}

func (store *Store) problems(project Project) ([]Code, error) {
	var problems []Code
	if project.Title == "" {
		problems = append(problems, CodeBlankTitle)
	}
	if project.CustomerID == "" {
		problems = append(problems, CodeNoCustomer)
	}
	if project.Price < 0 {
		problems = append(problems, CodeNegativePrice)
	}
	if project.VariantID == "" {
		return problems, nil
	}
	product, found, err := store.catalogue.FindByVariant(project.VariantID)
	if err != nil {
		return nil, err
	}
	if !found || product.Kind != catalogue.Digital {
		problems = append(problems, CodeProductNotDigital)
	}
	return problems, nil
}

// Move puts a project in another column. Moving it to Delivered records its sale; moving it out
// of Delivered takes that sale back, giving back what it took from the catalogue. Both happen in
// the same transaction as the move.
func (store *Store) Move(id string, stage Stage) ([]Code, error) {
	project, found, err := store.Find(id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, projectNotFound(id)
	}
	if project.Stage == stage {
		return nil, nil
	}
	recorded, taken, err := store.salesOfMove(project, stage)
	if err != nil {
		return nil, err
	}
	return nil, store.database.InTransaction(func(rows storage.Querier) error {
		return store.moveIn(rows, project, stage, recorded, taken)
	})
}

// salesOfMove are the sale a move records (when it delivers) and the sale it takes back (when it
// leaves Delivered), read before the transaction: the database has one connection, and the
// transaction holds it.
func (store *Store) salesOfMove(project Project, stage Stage) (recorded, taken sales.Sale, err error) {
	if stage == Delivered {
		recorded, err = store.saleOf(project)
		return recorded, taken, err
	}
	if project.Stage == Delivered && project.SaleID != "" {
		taken, _, err = sales.FindIn(store.database.Rows(), project.SaleID) // a sale deleted from the history is not found: nothing to take back
	}
	return recorded, taken, err
}

// moveIn records or takes back the sale of a move, and puts the project in its new stage.
func (store *Store) moveIn(rows storage.Querier, project Project, stage Stage, recorded, taken sales.Sale) error {
	now := store.now()
	if recorded.ID != "" {
		if err := sales.InsertIn(rows, recorded); err != nil {
			return err
		}
	}
	if taken.ID != "" {
		if err := sales.DeleteIn(rows, taken, now); err != nil {
			return err
		}
	}
	_, err := rows.ExecContext(context.Background(), "UPDATE project SET stage = ?, sale_id = ?, updated_at = ? WHERE id = ?",
		stage, recorded.ID, storage.Milliseconds(now), project.ID)
	if err != nil {
		return storageFailed("move project", err)
	}
	return nil
}

// saleOf is the sale a delivered project records: one line of its product, or of its title.
func (store *Store) saleOf(project Project) (sales.Sale, error) {
	customer, _, err := store.customers.Find(project.CustomerID)
	if err != nil {
		return sales.Sale{}, err
	}
	line := sales.Line{ID: store.newID(), VariantID: project.VariantID, ProductName: project.Title, Quantity: 1, UnitPrice: project.Price}
	if project.VariantID != "" {
		product, found, err := store.catalogue.FindByVariant(project.VariantID)
		if err != nil {
			return sales.Sale{}, err
		}
		variant, _ := product.Variant(project.VariantID) // a missing product costs nothing
		if found {
			line.UnitCost = variant.Cost
		}
	}
	return sales.Sale{ID: store.newID(), Channel: sales.Commission, Status: sales.Delivered, Customer: customer.Name,
		Note: project.Title, SoldAt: store.now(), Lines: []sales.Line{line}}, nil
}

// Delete removes a project that is not delivered; a delivered one keeps its sale in the history.
func (store *Store) Delete(id string) ([]Code, error) {
	project, found, err := store.Find(id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, projectNotFound(id)
	}
	if project.Stage == Delivered {
		return []Code{CodeAlreadyDelivered}, nil
	}
	at := storage.Milliseconds(store.now())
	_, err = store.database.Rows().ExecContext(context.Background(), "UPDATE project SET deleted_at = ?, updated_at = ? WHERE id = ?", at, at, id)
	if err != nil {
		return nil, storageFailed("delete project", err)
	}
	return nil, nil
}

func readProjects(rows storage.Querier, onlyID string) ([]Project, error) {
	result, err := rows.QueryContext(context.Background(), `SELECT p.id, p.title, p.customer_id, p.stage, p.deadline, p.price_cents,
			p.variant_id, p.sale_id, p.created_at, coalesce(s.sold_at, 0), p.rowid
		FROM project p LEFT JOIN sale s ON s.id = p.sale_id
		WHERE p.deleted_at IS NULL AND (?1 = '' OR p.id = ?1) ORDER BY p.deadline = '', p.deadline, p.created_at`, onlyID)
	if err != nil {
		return nil, storageFailed("read projects", err)
	}
	defer result.Close()
	var projects []Project
	for result.Next() {
		var project Project
		var deadline string
		var created, delivered int64
		err := result.Scan(&project.ID, &project.Title, &project.CustomerID, &project.Stage, &deadline, &project.Price, &project.VariantID,
			&project.SaleID, &created, &delivered, &project.sequence)
		if err != nil {
			return nil, storageFailed("read projects", err)
		}
		project.CreatedAt = storage.Moment(created)
		if delivered > 0 {
			project.DeliveredAt = storage.Moment(delivered)
		}
		if deadline != "" {
			project.Deadline, err = calendar.Parse(deadline)
			if err != nil {
				return nil, storageFailed("read project deadline", err)
			}
		}
		projects = append(projects, project)
	}
	return projects, result.Err()
}

func deadlineText(deadline calendar.Date) string {
	if deadline.Zero() {
		return ""
	}
	return deadline.String()
}
