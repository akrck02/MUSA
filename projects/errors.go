package projects

import "fmt"

// Code is a stable name for what went wrong, for programs to check and people to search for.
type Code string

// Error codes of this package. The first ones are problems the artist fixes in the editor; the
// last ones are failures.
const (
	CodeBlankTitle        Code = "projects.blank-title"
	CodeNoCustomer        Code = "projects.no-customer"
	CodeNegativePrice     Code = "projects.negative-price"
	CodeProductNotDigital Code = "projects.product-not-digital"
	CodeAlreadyDelivered  Code = "projects.already-delivered"
	CodeProjectNotFound   Code = "projects.project-not-found"
	CodeStorageFailed     Code = "projects.storage-failed"
)

// Error is a failure of the projects: its code, what it was about and what lay underneath.
type Error struct {
	Code   Code
	Detail string
	Cause  error
}

func (failure *Error) Error() string {
	if failure.Cause == nil {
		return fmt.Sprintf("%s: %s", failure.Code, failure.Detail)
	}
	return fmt.Sprintf("%s: %s: %v", failure.Code, failure.Detail, failure.Cause)
}

func (failure *Error) Unwrap() error { return failure.Cause }

func projectNotFound(id string) error { return &Error{Code: CodeProjectNotFound, Detail: id} }

func storageFailed(doing string, cause error) error {
	return &Error{Code: CodeStorageFailed, Detail: doing, Cause: cause}
}
