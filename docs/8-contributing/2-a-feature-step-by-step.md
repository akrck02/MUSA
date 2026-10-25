# 🧭 A feature, step by step
---

This page is for whoever builds their first feature in MUSA: by the end you have followed a real
one from the idea to the pull request, through every layer, in the order that keeps the app
working at each step.

The feature is small on purpose: **a note on a customer**, where the artist writes what to
remember about them (tastes, sizes, how they pay). It touches the model, the database, the store,
the screen, the tests and the docs, and nothing else. Its two commits are the code below.

## 0. Decide whether it needs a proposal

A small fix is a change, a test and a commit. A change that spans several packages, changes what
the artist sees in a new way, or needs a decision from someone else starts as a proposal in
`docs/proposals/` ([working together](6-working-together.md) says how). A note on a customer is
small enough to skip it; most features are not.

## 1. Make a branch

```bash
git switch -c feature/customer-note
```

One feature, one branch, named after what it does.

## 2. Write the test first, and see it fail

Start where the behaviour is promised: the store. The test says what must be true, in its name:

```go
// customers/customers_test.go
func TestACustomerKeepsANoteWithoutItsSurroundingSpaces(test *testing.T) {
	store := testStore(test)
	saved, problem, err := store.Save(Customer{Name: "Ana", Note: "  Prefers pastel colours  "})
	if err != nil || problem != "" {
		test.Fatal(problem, err)
	}
	customer, _, err := store.Find(saved.ID)
	if err != nil || customer.Note != "Prefers pastel colours" {
		test.Errorf("note %q, %v", customer.Note, err)
	}
}
```

`go test ./customers` fails: `unknown field Note`. Good: the test can fail, so it proves
something when it passes.

## 3. The model

The model is a plain Go struct. Add the field, with a comment that says what it is for:

```go
// customers/customer.go
type Customer struct {
	ID       string
	Name     string
	Contact  string // an email or a phone, as typed
	Note     string // what the artist wants to remember about them: tastes, sizes, how they pay
	Profiles []Profile
}
```

## 4. The migration

Artists already have databases without the column. A **migration** brings them up to date when
the app opens. Add one step at the **end** of the list in `storage/migrations.go`, with a comment
in the artist's terms:

```go
	// A customer can carry a note: what the artist wants to remember about them.
	"ALTER TABLE customer ADD COLUMN note TEXT NOT NULL DEFAULT ''",
```

Two rules, because the database counts the steps it has taken:

- **Never edit or reorder a step** once it is on `main`: databases out there have already run it.
- **Your step is a draft until you merge.** Write it now, so the feature works; when you rebase
  before merging, check that it is still the last step. If someone else's step was merged in the
  meantime, move yours after it in that rebase (see [working together](6-working-together.md)).

`DEFAULT ''` lets the old rows read as "no note", so nothing else has to change.

## 5. The store

The store is the only place that knows SQL. Read and write the new column, and clean the text the
way the other fields are cleaned:

```go
// customers/customers.go, in Save
customer.Note = strings.TrimSpace(customer.Note)

// in write
`INSERT INTO customer(id, name, contact, note, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET name = excluded.name, contact = excluded.contact, note = excluded.note,
	updated_at = excluded.updated_at`

// in read
"SELECT id, name, contact, note FROM customer …"
… result.Scan(&customer.ID, &customer.Name, &customer.Contact, &customer.Note)
```

Now `go test ./customers` passes. Commit: `customers: keep a note on each customer`.

## 6. The screen: test first again

Customers are edited from a project's editor. Write the flow test as a person would use it:
open a project, type the note, save, and check what the store kept:

```go
// app/projects_flow_test.go
func TestANoteOnTheCustomerIsSavedWithTheProject(test *testing.T) {
	program, window := openProject(test, "Retrato de familia")
	steps(test, window, typeInto(window, "Gustos, tallas, cómo paga…", "Prefiere colores pastel"), click(window, "Guardar"))
	settle(window)
	project, _ := projectTitled(test, program, "Retrato de familia")
	person, _, err := program.stores.Customers.Find(project.CustomerID)
	if err != nil || person.Note != "Prefiere colores pastel" {
		test.Errorf("the note is %q after saving (%v)", person.Note, err)
	}
}
```

It fails with `no element "Gustos, tallas, cómo paga…"`, and lists what the frame showed.

## 7. The screen

The editor already keeps the customer it edits (`editing.person`) and saves it whole, so the field
only has to point at the new value:

```go
// projects/editor.go, in customerFields
look.Field(context, "Nota", "customer-note", &editing.person.Note, "Gustos, tallas, cómo paga…", "")
```

The label and the hint are Spanish, for the artist; everything around them is English. `look.Field`
gives the field Bubble 4's look and a label that tests and screen readers find. The comment of
`customerFields` now mentions the note too: comments must keep telling the truth.

## 8. The docs

What the artist can do changed, so the guide changes in the same commit:
`docs/4-guides/5-projects.md` gains a line for **Nota**.

## 9. Check everything

```bash
gofmt -l .
go vet ./...
go test ./...
```

On Linux, add `-tags=x11` to these commands and to `go run`. `go test ./...` runs the style test and
the code map test too. Then look at it once: run
`go run ./cmd/musa -demo`, open a project, write a note, close and open it again.

Commit: `projects: write a note on the customer from the project's editor`.

## 10. Open the pull request

Push the branch and open a pull request against `feature/go-ui` (or `main`, once it is the base).
The template asks what changed, how it was tested, and what was looked at on screen. Another
person reviews it before it is merged.

## The order, in one line

**Test → model → migration → store → test passes → flow test → screen → docs → checks → pull
request.** Each step leaves the app building and its tests passing, so you can stop anywhere and
someone else can pick it up.
