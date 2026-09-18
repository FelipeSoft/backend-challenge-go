package domain

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyPlayerID     = errors.New("player id cannot be empty")
	ErrInvalidVersion    = errors.New("invalid wallet version: must be greater than or equal to 1")
	ErrRehydrationReq    = errors.New("wallet id and player id are required for rehydration")
	ErrInsufficientFunds = errors.New("insufficient funds: debit would result in a negative wallet balance")
)

type ErrCurrencyMismatch struct {
	WalletCurrency    string
	AttemptedCurrency string
}

func (e *ErrCurrencyMismatch) Error() string {
	return fmt.Sprintf("currency mismatch: wallet currency is %s, attempted operation in %s", e.WalletCurrency, e.AttemptedCurrency)
}

func (e *ErrCurrencyMismatch) Is(target error) bool {
	_, ok := target.(*ErrCurrencyMismatch)
	return ok
}