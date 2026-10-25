package calendar

import "fmt"

// Code is a stable name for what went wrong, for programs to check and people to search for.
type Code string

// CodeInvalidDate is the code of text that is not a stored date.
const CodeInvalidDate Code = "calendar.invalid-date"

// Error is a failure of the calendar: its code, the text it was about and the cause underneath.
type Error struct {
	Code   Code
	Detail string
	Cause  error
}

// Error is the code, the detail and the cause underneath, for logs.
func (failure *Error) Error() string {
	return fmt.Sprintf("%s: %q: %v", failure.Code, failure.Detail, failure.Cause)
}

// Unwrap is the cause underneath, for errors.Is and errors.As.
func (failure *Error) Unwrap() error { return failure.Cause }

func invalidDate(text string, cause error) error {
	return &Error{Code: CodeInvalidDate, Detail: text, Cause: cause}
}
