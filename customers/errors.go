package customers

import "fmt"

// Code is a stable name for what went wrong, for programs to check and people to search for.
type Code string

// Error codes of this package. The first ones are problems the artist fixes in a form; the last
// one is a failure.
const (
	CodeBlankName      Code = "customers.blank-name"
	CodeNetworkPattern Code = "customers.network-pattern"
	CodeStorageFailed  Code = "customers.storage-failed"
)

// Error is a failure of the customers: its code, what it was doing and what lay underneath.
type Error struct {
	Code   Code
	Detail string
	Cause  error
}

func (failure *Error) Error() string {
	return fmt.Sprintf("%s: %s: %v", failure.Code, failure.Detail, failure.Cause)
}

func (failure *Error) Unwrap() error { return failure.Cause }

func storageFailed(doing string, cause error) error {
	return &Error{Code: CodeStorageFailed, Detail: doing, Cause: cause}
}
