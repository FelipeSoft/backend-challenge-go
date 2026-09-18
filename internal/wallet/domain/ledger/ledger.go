package domain

import (
	"fmt"
	"time"

	"github.com/FelipeSoft/backend-challenge-go/internal/domain"
)

type Direction string

const (
	DirectionDebit  Direction = "DEBIT"
	DirectionCredit Direction = "CREDIT"
)

type WalletLedgerEntry struct {
	id            string
	walletID      string
	transactionId string
	direction     Direction
	amount        domain.Money
	balanceBefore domain.Money
	balanceAfter  domain.Money
	createdAt     time.Time
}

func NewWalletLedgerEntry(
	id string,
	walletID string,
	transactionId string,
	direction Direction,
	amount domain.Money,
	balanceBefore domain.Money,
	now time.Time,
) (WalletLedgerEntry, error) {
	if id == "" {
		return WalletLedgerEntry{}, ErrEmptyLedgerID
	}
	if walletID == "" {
		return WalletLedgerEntry{}, ErrEmptyLedgerWallet
	}
	if transactionId == "" {
		return WalletLedgerEntry{}, ErrEmptyLedgerTx
	}
	if direction != DirectionDebit && direction != DirectionCredit {
		return WalletLedgerEntry{}, fmt.Errorf("%w: %s", ErrInvalidDirection, direction)
	}
	if amount.Currency() != balanceBefore.Currency() {
		return WalletLedgerEntry{}, &ErrLedgerCurrencyMismatch{
			AmountCurrency:  amount.Currency(),
			BalanceCurrency: balanceBefore.Currency(),
		}
	}
	var balanceAfter domain.Money
	var err error
	switch direction {
	case DirectionCredit:
		balanceAfter, err = balanceBefore.Add(amount)
		if err != nil {
			return WalletLedgerEntry{}, fmt.Errorf("failed to calculate balanceAfter for credit: %w", err)
		}
	case DirectionDebit:
		balanceAfter, err = balanceBefore.Sub(amount)
		if err != nil {
			return WalletLedgerEntry{}, fmt.Errorf("failed to calculate balanceAfter for debit: %w", err)
		}
	}
	return WalletLedgerEntry{
		id:            id,
		walletID:      walletID,
		transactionId: transactionId,
		direction:     direction,
		amount:        amount,
		balanceBefore: balanceBefore,
		balanceAfter:  balanceAfter,
		createdAt:     now,
	}, nil
}

func (e WalletLedgerEntry) ID() string                  { return e.id }
func (e WalletLedgerEntry) WalletID() string            { return e.walletID }
func (e WalletLedgerEntry) TransactionId() string       { return e.transactionId }
func (e WalletLedgerEntry) Direction() Direction        { return e.direction }
func (e WalletLedgerEntry) Amount() domain.Money        { return e.amount }
func (e WalletLedgerEntry) BalanceBefore() domain.Money { return e.balanceBefore }
func (e WalletLedgerEntry) BalanceAfter() domain.Money  { return e.balanceAfter }
func (e WalletLedgerEntry) CreatedAt() time.Time        { return e.createdAt }
