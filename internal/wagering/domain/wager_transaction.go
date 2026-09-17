package domain

import (
	"time"

	"github.com/FelipeSoft/jungle-gaming/internal/domain"
)

type TransactionType string

const (
	TypeOpening  TransactionType = "OPENING"
	TypeBet      TransactionType = "BET"
	TypeWin      TransactionType = "WIN"
	TypeLoss     TransactionType = "LOSS"
	TypeRefund   TransactionType = "REFUND"
	TypeRollback TransactionType = "ROLLBACK"
)

type TransactionState string

const (
	StatePending          TransactionState = "PENDING"
	StatePendingReference TransactionState = "PENDING_REFERENCE"
	StateProcessed        TransactionState = "PROCESSED"
	StateRejected         TransactionState = "REJECTED"
	StateFailed           TransactionState = "FAILED"
)

type WagerTransaction struct {
	id                  string
	externalID          *string
	provider            *string
	idempotencyKey      *string
	payloadHash         *string
	walletID            string
	playerID            string
	roundID             *string
	gameID              *string
	txType              TransactionType
	amount              domain.Money
	externalReference   *string
	resolvedInternalRef *string
	state               TransactionState
	failureCode         *string
	providerResult      *string
	createdAt           time.Time
	updatedAt           time.Time
}

func NewInternalOpening(id, walletID, playerID string, amount domain.Money, now time.Time) (WagerTransaction, error) {
	if id == "" || walletID == "" || playerID == "" {
		return WagerTransaction{}, ErrOpeningFieldsRequired
	}
	return WagerTransaction{
		id:        id,
		walletID:  walletID,
		playerID:  playerID,
		txType:    TypeOpening,
		amount:    amount,
		state:     StatePending,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func NewExternalTransaction(
	id string,
	externalID, provider, idempotencyKey, payloadHash string,
	walletID, playerID string,
	roundID, gameID *string,
	txType TransactionType,
	amount domain.Money,
	externalReference *string,
	now time.Time,
) (WagerTransaction, error) {
	if txType == TypeOpening {
		return WagerTransaction{}, ErrOpeningNotAllowedExternal
	}
	if id == "" || externalID == "" || provider == "" || idempotencyKey == "" || payloadHash == "" || walletID == "" || playerID == "" {
		return WagerTransaction{}, ErrMissingExternalFields
	}
	return WagerTransaction{
		id:                id,
		externalID:        &externalID,
		provider:          &provider,
		idempotencyKey:    &idempotencyKey,
		payloadHash:       &payloadHash,
		walletID:          walletID,
		playerID:          playerID,
		roundID:           roundID,
		gameID:            gameID,
		txType:            txType,
		amount:            amount,
		externalReference: externalReference,
		state:             StatePending,
		createdAt:         now,
		updatedAt:         now,
	}, nil
}

func (t WagerTransaction) IsTerminal() bool {
	return t.state == StateProcessed || t.state == StateRejected || t.state == StateFailed
}

func (t *WagerTransaction) TransitionTo(newState TransactionState, failureCode *string, now time.Time) error {
	if t.IsTerminal() {
		return &ErrTerminalTransition{
			CurrentState: t.state,
			TargetState:  newState,
		}
	}
	switch t.state {
	case StatePending:
		if newState != StatePendingReference && newState != StateProcessed && newState != StateRejected && newState != StateFailed {
			return &ErrInvalidTransition{
				CurrentState: t.state,
				TargetState:  newState,
			}
		}
	case StatePendingReference:
		if newState != StateProcessed && newState != StateRejected && newState != StateFailed {
			return &ErrInvalidTransition{
				CurrentState: t.state,
				TargetState:  newState,
			}
		}
	default:
		return ErrUnexpectedOrTerminalState
	}
	t.state = newState
	t.failureCode = failureCode
	t.updatedAt = now
	return nil
}

func (t WagerTransaction) ID() string                 { return t.id }
func (t WagerTransaction) State() TransactionState    { return t.state }
func (t WagerTransaction) Type() TransactionType      { return t.txType }
func (t WagerTransaction) WalletID() string           { return t.walletID }
func (t WagerTransaction) PlayerID() string           { return t.playerID }
func (t WagerTransaction) Amount() domain.Money       { return t.amount }
func (t WagerTransaction) IdempotencyKey() *string    { return t.idempotencyKey }