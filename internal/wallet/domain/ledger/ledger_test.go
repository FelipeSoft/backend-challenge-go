package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/FelipeSoft/backend-challenge-go/internal/domain"
)

func TestNewWalletLedgerEntry_Success(t *testing.T) {
	now := time.Now()
	amount, err := domain.NewMoneyFromString("50.00", "BRL")
	if err != nil {
		t.Fatalf("failed to create amount: %v", err)
	}
	balanceBefore, err := domain.NewMoneyFromString("150.00", "BRL")
	if err != nil {
		t.Fatalf("failed to create balanceBefore: %v", err)
	}

	entry, err := NewWalletLedgerEntry(
		"ledger-1",
		"wallet-1",
		"tx-1",
		DirectionDebit,
		amount,
		balanceBefore,
		now,
	)
	if err != nil {
		t.Fatalf("unexpected error creating ledger entry: %v", err)
	}

	if entry.ID() != "ledger-1" {
		t.Errorf("expected id 'ledger-1', got '%s'", entry.ID())
	}
	if entry.WalletID() != "wallet-1" {
		t.Errorf("expected walletID 'wallet-1', got '%s'", entry.WalletID())
	}
	if entry.TransactionId() != "tx-1" {
		t.Errorf("expected transactionId 'tx-1', got '%s'", entry.TransactionId())
	}
	if entry.Direction() != DirectionDebit {
		t.Errorf("expected direction DEBIT, got '%s'", entry.Direction())
	}
	if entry.BalanceAfter().Amount() != 10000 { // 150.00 - 50.00 = 100.00 (10000 centavos)
		t.Errorf("expected balanceAfter 10000 centavos, got %d", entry.BalanceAfter().Amount())
	}
}

func TestNewWalletLedgerEntry_Credit(t *testing.T) {
	now := time.Now()
	amount, _ := domain.NewMoneyFromString("25.00", "BRL")
	balanceBefore, _ := domain.NewMoneyFromString("100.00", "BRL")

	entry, err := NewWalletLedgerEntry(
		"ledger-2",
		"wallet-1",
		"tx-2",
		DirectionCredit,
		amount,
		balanceBefore,
		now,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.BalanceAfter().Amount() != 12500 { // 100.00 + 25.00 = 125.00
		t.Errorf("expected balanceAfter 12500 centavos, got %d", entry.BalanceAfter().Amount())
	}
}

func TestNewWalletLedgerEntry_ValidationErrors(t *testing.T) {
	now := time.Now()
	amount, _ := domain.NewMoneyFromString("10.00", "BRL")
	balanceBefore, _ := domain.NewMoneyFromString("50.00", "BRL")
	usdAmount, _ := domain.NewMoneyFromString("10.00", "USD")

	t.Run("Empty ID", func(t *testing.T) {
		_, err := NewWalletLedgerEntry("", "w-1", "tx-1", DirectionDebit, amount, balanceBefore, now)
		if err != ErrEmptyLedgerID {
			t.Errorf("expected ErrEmptyLedgerID, got %v", err)
		}
	})

	t.Run("Empty Wallet ID", func(t *testing.T) {
		_, err := NewWalletLedgerEntry("ledger-1", "", "tx-1", DirectionDebit, amount, balanceBefore, now)
		if err != ErrEmptyLedgerWallet {
			t.Errorf("expected ErrEmptyLedgerWallet, got %v", err)
		}
	})

	t.Run("Empty Transaction ID", func(t *testing.T) {
		_, err := NewWalletLedgerEntry("ledger-1", "w-1", "", DirectionDebit, amount, balanceBefore, now)
		if err != ErrEmptyLedgerTx {
			t.Errorf("expected ErrEmptyLedgerTx, got %v", err)
		}
	})

	t.Run("Invalid Direction", func(t *testing.T) {
		_, err := NewWalletLedgerEntry("ledger-1", "w-1", "tx-1", Direction("INVALID"), amount, balanceBefore, now)
		if err == nil {
			t.Error("expected error for invalid direction, got nil")
		}
	})

	t.Run("Currency Mismatch", func(t *testing.T) {
		_, err := NewWalletLedgerEntry("ledger-1", "w-1", "tx-1", DirectionDebit, usdAmount, balanceBefore, now)
		if err == nil {
			t.Error("expected currency mismatch error, got nil")
		}
		var currencyErr *ErrLedgerCurrencyMismatch
		if !errors.As(err, &currencyErr) {
			t.Errorf("expected ErrLedgerCurrencyMismatch, got %T", err)
		}
	})
}
