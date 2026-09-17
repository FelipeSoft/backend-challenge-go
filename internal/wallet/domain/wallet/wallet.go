package domain

import (
	"time"

	"github.com/FelipeSoft/jungle-gaming/internal/domain"
	"github.com/google/uuid"
)

type Wallet struct {
	id        string
	playerId  string
	balance   domain.Money
	version   int64
	createdAt time.Time
	updatedAt time.Time
}

func NewWallet(playerId string, currency string, now time.Time) (Wallet, error) {
	if playerId == "" {
		return Wallet{}, ErrEmptyPlayerID
	}
	zeroMoney, err := domain.Zero(currency)
	if err != nil {
		return Wallet{}, err
	}
	walletId, err := uuid.NewV7()
	if err != nil {
		return Wallet{}, err
	}
	return Wallet{
		id:        walletId.String(),
		playerId:  playerId,
		balance:   zeroMoney,
		version:   1,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func RehydrateWallet(id string, playerId string, balance domain.Money, version int64, createdAt, updatedAt time.Time) (Wallet, error) {
	if id == "" || playerId == "" {
		return Wallet{}, ErrRehydrationReq
	}
	if version < 1 {
		return Wallet{}, ErrInvalidVersion
	}
	return Wallet{
		id:        id,
		playerId:  playerId,
		balance:   balance,
		version:   version,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}, nil
}

func (w *Wallet) Debit(amount domain.Money, now time.Time) error {
	if amount.Currency() != w.Currency() {
		return &ErrCurrencyMismatch{
			WalletCurrency:    w.Currency(),
			AttemptedCurrency: amount.Currency(),
		}
	}
	newBalance, err := w.balance.Sub(amount)
	if err != nil {
		return err
	}
	if newBalance.Amount() < 0 {
		return ErrInsufficientFunds
	}
	if amount.Amount() == 0 {
		return nil
	}
	w.balance = newBalance
	w.version++
	w.updatedAt = now
	return nil
}

func (w *Wallet) Credit(amount domain.Money, now time.Time) error {
	if amount.Currency() != w.Currency() {
		return &ErrCurrencyMismatch{
			WalletCurrency:    w.Currency(),
			AttemptedCurrency: amount.Currency(),
		}
	}
	newBalance, err := w.balance.Add(amount)
	if err != nil {
		return err
	}
	if amount.Amount() == 0 {
		return nil
	}
	w.balance = newBalance
	w.version++
	w.updatedAt = now
	return nil
}

func (w Wallet) ID() string            { return w.id }
func (w Wallet) PlayerID() string      { return w.playerId }
func (w Wallet) Currency() string      { return w.balance.Currency() }
func (w Wallet) Balance() domain.Money { return w.balance }
func (w Wallet) Version() int64        { return w.version }
func (w Wallet) CreatedAt() time.Time  { return w.createdAt }
func (w Wallet) UpdatedAt() time.Time  { return w.updatedAt }
