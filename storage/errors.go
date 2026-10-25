package storage

import "fmt"

// Code is a stable name for what went wrong, for programs to check and people to search for.
type Code string

// Error codes of this package.
const (
	CodeOpenFailed        Code = "storage.open-failed"
	CodeTransactionFailed Code = "storage.transaction-failed"
)

// Error is a storage failure: its code, and what lay underneath.
type Error struct {
	Code   Code
	Detail string
	Cause  error
}

// Error is the code, the detail and the cause underneath, for logs.
func (failure *Error) Error() string {
	return fmt.Sprintf("%s: %s: %v", failure.Code, failure.Detail, failure.Cause)
}

// Unwrap is the cause underneath, for errors.Is and errors.As.
func (failure *Error) Unwrap() error { return failure.Cause }

func openFailed(path string, cause error) error {
	return &Error{Code: CodeOpenFailed, Detail: path, Cause: cause}
}

func transactionFailed(cause error) error {
	return &Error{Code: CodeTransactionFailed, Detail: "commit or begin", Cause: cause}
}
