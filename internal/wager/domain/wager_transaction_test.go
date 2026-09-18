package domain

import (
	"testing"
	"time"

	"github.com/FelipeSoft/backend-challenge-go/internal/domain"
)

func TestNewInternalOpening_Success(t *testing.T) {
	now := time.Now()
	amount, _ := domain.NewMoneyFromString("100.00", "BRL")
	tx, err := NewInternalOpening("w-123", "p-456", amount, now)
	if err != nil {
		t.Fatalf("unexpected error creating internal opening: %v", err)
	}
	if tx.Type() != TypeOpening {
		t.Errorf("expected type OPENING, got %s", tx.Type())
	}
	if tx.State() != StatePending {
		t.Errorf("expected initial state PENDING, got %s", tx.State())
	}
	if tx.IdempotencyKey() != nil {
		t.Error("expected idempotency key to be nil for internal opening")
	}
}

func TestNewExternalTransaction_Success(t *testing.T) {
	now := time.Now()
	amount, _ := domain.NewMoneyFromString("50.00", "BRL")
	roundID := "round-01"
	gameID := "game-slot-01"
	tx, err := NewExternalTransaction(
		"tx-ext-1",
		"ext-id-999",
		"provider-xyz",
		"idem-key-123",
		"hash-abc-789",
		"w-123",
		"p-456",
		&roundID,
		&gameID,
		TypeBet,
		amount,
		nil,
		now,
	)
	if err != nil {
		t.Fatalf("unexpected error creating external transaction: %v", err)
	}
	if tx.Type() != TypeBet {
		t.Errorf("expected type BET, got %s", tx.Type())
	}
	if tx.State() != StatePending {
		t.Errorf("expected initial state PENDING, got %s", tx.State())
	}
}

func TestNewExternalTransaction_RejectsOpeningType(t *testing.T) {
	now := time.Now()
	amount, _ := domain.NewMoneyFromString("50.00", "BRL")
	_, err := NewExternalTransaction(
		"tx-ext-1",
		"ext-id-999",
		"provider-xyz",
		"idem-key-123",
		"hash-abc-789",
		"w-123",
		"p-456",
		nil,
		nil,
		TypeOpening,
		amount,
		nil,
		now,
	)
	if err == nil {
		t.Error("expected error when trying to create OPENING via external channel, got nil")
	}
}

func TestWagerTransaction_StateMachineTransitions(t *testing.T) {
	now := time.Now()
	amount, _ := domain.NewMoneyFromString("10.00", "BRL")
	roundID := "r-1"
	gameID := "g-1"
	tx, _ := NewExternalTransaction(
		"tx-1", "ext-1", "prov", "idem-1", "hash-1",
		"w-1", "p-1", &roundID, &gameID, TypeBet, amount, nil, now,
	)
	err := tx.TransitionTo(StatePendingReference, nil, now)
	if err != nil {
		t.Errorf("failed to transition to PENDING_REFERENCE: %v", err)
	}
	err = tx.TransitionTo(StateProcessed, nil, now)
	if err != nil {
		t.Errorf("failed to transition to PROCESSED: %v", err)
	}
	failureCode := "RULE_VIOLATION"
	err = tx.TransitionTo(StateRejected, &failureCode, now)
	if err == nil {
		t.Error("expected error when attempting to transition from a terminal state, got nil")
	}
}

func TestWagerTransaction_DirectToFailedTerminalState(t *testing.T) {
	now := time.Now()
	amount, _ := domain.NewMoneyFromString("10.00", "BRL")
	roundID := "r-1"
	gameID := "g-1"
	tx, _ := NewExternalTransaction(
		"tx-2", "ext-2", "prov", "idem-2", "hash-2",
		"w-1", "p-1", &roundID, &gameID, TypeWin, amount, nil, now,
	)
	failureCode := "INFRA_TIMEOUT"
	err := tx.TransitionTo(StateFailed, &failureCode, now)
	if err != nil {
		t.Errorf("unexpected error transitioning directly to FAILED: %v", err)
	}
	if !tx.IsTerminal() {
		t.Error("expected transaction to be in a terminal state")
	}
}
