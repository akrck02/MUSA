package demo

import (
	"errors"
	"fmt"
)

// Code is a stable name for what went wrong, for programs to check and people to search for.
type Code string

// Error codes of this package.
const (
	// CodeNotEmpty is the code of filling a database that already has data: its SKUs would clash.
	CodeNotEmpty Code = "demo.not-empty"
	// CodeRefused is the code of demo data a store turned down, which is a bug in the demo.
	CodeRefused Code = "demo.refused"
)

// Error is a failure of the demo: its code and what it was about.
type Error struct {
	Code   Code
	Detail string
}

// Error is the code and the detail, for logs.
func (failure *Error) Error() string { return fmt.Sprintf("%s: %s", failure.Code, failure.Detail) }

// CodeOf is the code of an error the demo returned, or "" for any other error.
func CodeOf(err error) Code {
	var failure *Error
	if !errors.As(err, &failure) {
		return ""
	}
	return failure.Code
}

func notEmpty() error {
	return &Error{Code: CodeNotEmpty, Detail: "the database already has products or events"}
}

func refused(refusals any) error { return &Error{Code: CodeRefused, Detail: fmt.Sprint(refusals)} }
