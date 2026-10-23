package events

import "fmt"

// Code is a stable name for what went wrong, for programs to check and people to search for.
type Code string

// Error codes of this package. The first ones are why an action on a fair is turned down, which
// the screens show; the last ones are failures.
const (
	CodeBlankName        Code = "events.blank-name"
	CodeEndBeforeStart   Code = "events.end-before-start"
	CodeInvalidQuantity  Code = "events.invalid-quantity"
	CodeNegativePrice    Code = "events.negative-price"
	CodeBlankConcept     Code = "events.blank-concept"
	CodeNegativeAmount   Code = "events.negative-amount"
	CodeNothingToSell    Code = "events.nothing-to-sell"
	CodeEventClosed      Code = "events.event-closed"
	CodeHasSales         Code = "events.has-sales"
	CodeNotEnoughStock   Code = "events.not-enough-stock"
	CodeExceedsRemaining Code = "events.exceeds-remaining"
	CodeEventNotFound    Code = "events.event-not-found"
	CodeNotTaken         Code = "events.variant-not-taken"
	CodeSaleNotFound     Code = "events.sale-not-found"
	CodeStorageFailed    Code = "events.storage-failed"
)

// Refusal is why an action on a fair was turned down. Available is how many units there were, for
// CodeNotEnoughStock (in the catalogue) and CodeExceedsRemaining (on the stand).
type Refusal struct {
	Code      Code
	Available int
}

// Error is a failure of the events: its code and what it was about.
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

func eventNotFound(id string) error { return &Error{Code: CodeEventNotFound, Detail: id} }

func notTaken(variantID string) error { return &Error{Code: CodeNotTaken, Detail: variantID} }

func saleNotFound(id string) error { return &Error{Code: CodeSaleNotFound, Detail: id} }

func storageFailed(doing string, cause error) error {
	return &Error{Code: CodeStorageFailed, Detail: doing, Cause: cause}
}

func refusals(codes ...Code) []Refusal {
	all := make([]Refusal, len(codes))
	for index, code := range codes {
		all[index] = Refusal{Code: code}
	}
	return all
}
