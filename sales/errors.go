package sales

import "fmt"

// Code is a stable name for what went wrong, for programs to check and people to search for.
type Code string

// Error codes of this package. The first ones are problems of a draft, which the editor shows;
// the last ones are failures.
const (
	CodeNoLines          Code = "sales.no-lines"
	CodeEventChannel     Code = "sales.event-channel"
	CodeNegativeShipping Code = "sales.negative-shipping"
	CodeManagedByEvent   Code = "sales.managed-by-event"
	CodeInvalidQuantity  Code = "sales.invalid-quantity"
	CodeNegativePrice    Code = "sales.negative-price"
	CodeNotEnoughStock   Code = "sales.not-enough-stock"
	CodeSaleNotFound     Code = "sales.sale-not-found"
	CodeVariantNotFound  Code = "sales.variant-not-found"
	CodeStorageFailed    Code = "sales.storage-failed"
)

// Problem is one reason a sale cannot be recorded or deleted. LineID points at the line it is
// about; VariantID and Available tell how much stock there was.
type Problem struct {
	Code      Code
	LineID    string
	VariantID string
	Available int
}

// Error is a failure of the sales: its code and what it was about.
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

func saleNotFound(id string) error { return &Error{Code: CodeSaleNotFound, Detail: id} }

func variantNotFound(id string) error { return &Error{Code: CodeVariantNotFound, Detail: id} }

func storageFailed(doing string, cause error) error {
	return &Error{Code: CodeStorageFailed, Detail: doing, Cause: cause}
}
