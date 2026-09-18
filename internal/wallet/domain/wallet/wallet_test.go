package domain

import (
	"testing"
	"time"

	"github.com/FelipeSoft/backend-challenge-go/internal/domain"
)

func TestNewWallet_Success(t *testing.T) {
	now := time.Now()
	wallet, err := NewWallet("p-456", "BRL", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wallet.Version() != 1 {
		t.Errorf("expected initial version to be 1, got %d", wallet.Version())
	}
	if wallet.Balance().Amount() != 0 {
		t.Errorf("expected initial balance to be 0, got %d", wallet.Balance().Amount())
	}
	if wallet.Currency() != "BRL" {
		t.Errorf("expected currency BRL, got %s", wallet.Currency())
	}
}

func TestWallet_CreditAndDebit(t *testing.T) {
	now := time.Now()
	wallet, _ := NewWallet("p-456", "BRL", now)
	deposit, _ := domain.NewMoneyFromString("100.00", "BRL")
	err := wallet.Credit(deposit, now)
	if err != nil {
		t.Fatalf("failed to credit: %v", err)
	}
	if wallet.Balance().Amount() != 10000 {
		t.Errorf("expected balance 10000 cents, got %d", wallet.Balance().Amount())
	}
	if wallet.Version() != 2 {
		t.Errorf("expected version to increment to 2, got %d", wallet.Version())
	}
	withdraw, _ := domain.NewMoneyFromString("40.00", "BRL")
	err = wallet.Debit(withdraw, now)
	if err != nil {
		t.Fatalf("failed to debit: %v", err)
	}
	if wallet.Balance().Amount() != 6000 {
		t.Errorf("expected balance 6000 cents, got %d", wallet.Balance().Amount())
	}
	if wallet.Version() != 3 {
		t.Errorf("expected version to increment to 3, got %d", wallet.Version())
	}
}

func TestWallet_DebitInsufficientFunds(t *testing.T) {
	now := time.Now()
	wallet, _ := NewWallet("p-456", "BRL", now)
	withdraw, _ := domain.NewMoneyFromString("50.00", "BRL")
	err := wallet.Debit(withdraw, now)
	if err == nil {
		t.Error("expected error due to insufficient funds, got nil")
	}
	if wallet.Version() != 1 {
		t.Errorf("version should not change on failed operation, got %d", wallet.Version())
	}
}

func TestWallet_CurrencyMismatchOperations(t *testing.T) {
	now := time.Now()
	wallet, _ := NewWallet("p-456", "BRL", now)
	usdMoney, _ := domain.NewMoneyFromString("50.00", "USD")
	errCredit := wallet.Credit(usdMoney, now)
	if errCredit == nil {
		t.Error("expected error when crediting different currency, got nil")
	}
	errDebit := wallet.Debit(usdMoney, now)
	if errDebit == nil {
		t.Error("expected error when debiting different currency, got nil")
	}
}
