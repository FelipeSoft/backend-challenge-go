package domain

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyMoneyString     = errors.New("money string cannot be empty")
	ErrInvalidMoneyFormat   = errors.New("invalid format: scientific notation or positive sign not allowed")
	ErrNegativeExternal     = errors.New("negative values are not allowed in external financial inputs")
	ErrMultipleSeparators   = errors.New("invalid format: multiple decimal separators")
	ErrOverflowAddition     = errors.New("overflow error during addition")
	ErrOverflowSubtraction  = errors.New("overflow error during subtraction")
	ErrOverflowNegation     = errors.New("overflow error during negation of math.MinInt64")
)

type ErrInvalidMoneyDecimalPlaces struct {
	Expected int
	Got      int
}

func (e *ErrInvalidMoneyDecimalPlaces) Error() string {
	return fmt.Sprintf("invalid scale: expected exactly %d decimal places, got %d", e.Expected, e.Got)
}

func (e *ErrInvalidMoneyDecimalPlaces) Is(target error) bool {
	_, ok := target.(*ErrInvalidMoneyDecimalPlaces)
	return ok
}

type ErrMoneyOverflow struct {
	Err error
}

func (e *ErrMoneyOverflow) Error() string {
	return fmt.Sprintf("overflow error: value is too large or invalid: %v", e.Err)
}

func (e *ErrMoneyOverflow) Unwrap() error {
	return e.Err
}

func (e *ErrMoneyOverflow) Is(target error) bool {
	_, ok := target.(*ErrMoneyOverflow)
	return ok
}

type ErrMoneyCurrencyMismatch struct {
	Operation string
	CurrencyA string
	CurrencyB string
}

func (e *ErrMoneyCurrencyMismatch) Error() string {
	return fmt.Sprintf("currency mismatch: cannot %s %s and %s", e.Operation, e.CurrencyA, e.CurrencyB)
}

func (e *ErrMoneyCurrencyMismatch) Is(target error) bool {
	_, ok := target.(*ErrMoneyCurrencyMismatch)
	return ok
}

type ErrInvalidCurrency struct {
	Currency string
	Reason   string
}

func (e *ErrInvalidCurrency) Error() string {
	return fmt.Sprintf("invalid currency %q: %s", e.Currency, e.Reason)
}

func (e *ErrInvalidCurrency) Is(target error) bool {
	_, ok := target.(*ErrInvalidCurrency)
	return ok
}