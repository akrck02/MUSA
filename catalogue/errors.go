package catalogue

import "fmt"

// Code is a stable name for what went wrong, for programs to check and people to search for.
type Code string

// Error codes of this package. The first ones are problems of a draft, which the editor shows
// beside the field; the last ones are failures.
const (
	CodeBlankName            Code = "catalogue.blank-name"
	CodeNoVariants           Code = "catalogue.no-variants"
	CodeNegativeTurnaround   Code = "catalogue.negative-turnaround"
	CodeBlankVariantName     Code = "catalogue.blank-variant-name"
	CodeDuplicateVariantName Code = "catalogue.duplicate-variant-name"
	CodeNegativePrice        Code = "catalogue.negative-price"
	CodeNegativeCost         Code = "catalogue.negative-cost"
	CodeNegativeStock        Code = "catalogue.negative-stock"
	CodeNegativeMinimumStock Code = "catalogue.negative-minimum-stock"
	CodeInvalidMeasure       Code = "catalogue.invalid-measure"
	CodeDuplicateSKU         Code = "catalogue.duplicate-sku"
	CodeProductNotFound      Code = "catalogue.product-not-found"
	CodeVariantNotFound      Code = "catalogue.variant-not-found"
	CodeStorageFailed        Code = "catalogue.storage-failed"
)

// Problem is one reason a draft cannot be saved. VariantID names the variant it is about, when it
// is about one; OwnerName is the other product that already uses a SKU.
type Problem struct {
	Code      Code
	VariantID string
	OwnerName string
}

// Error is a failure of the catalogue: its code and what it was about.
type Error struct {
	Code   Code
	Detail string
	Cause  error
}

// Error is the code, the detail and the cause underneath, for logs.
func (failure *Error) Error() string {
	if failure.Cause == nil {
		return fmt.Sprintf("%s: %s", failure.Code, failure.Detail)
	}
	return fmt.Sprintf("%s: %s: %v", failure.Code, failure.Detail, failure.Cause)
}

// Unwrap is the cause underneath, for errors.Is and errors.As.
func (failure *Error) Unwrap() error { return failure.Cause }

func productNotFound(id string) error {
	return &Error{Code: CodeProductNotFound, Detail: id}
}

func variantNotFound(id string) error {
	return &Error{Code: CodeVariantNotFound, Detail: id}
}

func storageFailed(doing string, cause error) error {
	return &Error{Code: CodeStorageFailed, Detail: doing, Cause: cause}
}
