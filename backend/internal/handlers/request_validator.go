package handlers

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

// newDecimalAwareValidator returns a *validator.Validate that understands
// decimal.Decimal fields carrying gt=0/gte=0 constraints (see
// models.AuctionRequest's StartPrice/MinIncrement/InsuranceAmount).
//
// go-playground/validator's built-in gt/gte tags only understand native
// numeric kinds via reflection and fail with "Bad field type decimal.Decimal"
// (surfaced to the client as an HTTP 500) on any other struct type unless
// it's taught how to read a comparable value out of it.
//
// Every decimal.Decimal + validate tag combination reachable through this
// handler's validator is a pure sign-vs-zero comparison (gt=0 or gte=0) — no
// lt/lte or comparison against a non-zero bound exists anywhere in this
// package. decimal.Decimal.Sign() returns exactly -1/0/1, derived directly
// from the underlying big.Int coefficient's sign with no float conversion,
// so comparing that sign against 0 via the built-in gt/gte tags is exact —
// unlike a float64 view (e.g. InexactFloat64()), it cannot mis-round an
// arbitrarily small non-zero decimal to 0 or vice versa.
func newDecimalAwareValidator() *validator.Validate {
	v := validator.New()
	v.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		if d, ok := field.Interface().(decimal.Decimal); ok {
			return d.Sign()
		}
		return nil
	}, decimal.Decimal{})
	return v
}
