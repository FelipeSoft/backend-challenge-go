package domain

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyLedgerID      = errors.New("ledger entry id cannot be empty")
	ErrEmptyLedgerWallet  = errors.New("wallet id cannot be empty")
	ErrEmptyLedgerTx      = errors.New("transaction id cannot be empty")
	ErrInvalidDirection   = errors.New("invalid ledger entry direction")
)

type ErrLedgerCurrencyMismatch struct {
	AmountCurrency   string
	BalanceCurrency  string
}

func (e *ErrLedgerCurrencyMismatch) Error() string {
	return fmt.Sprintf("currency mismatch between amount (%s) and balanceBefore (%s)", e.AmountCurrency, e.BalanceCurrency)
}

func (e *ErrLedgerCurrencyMismatch) Is(target error) bool {
	_, ok := target.(*ErrLedgerCurrencyMismatch)
	return ok
}