package domain

import (
	"time"

	"github.com/FelipeSoft/backend-challenge-go/internal/domain"
	"github.com/google/uuid"
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
	id                             string
	externalID                     *string
	provider                       *string
	idempotencyKey                 *string
	payloadHash                    *string
	walletID                       string
	playerID                       string
	roundID                        *string
	gameID                         *string
	txType                         TransactionType
	amount                         domain.Money
	referenceExternalTransactionID *string
	resolvedInternalRef            *string
	state                          TransactionState
	failureCode                    *string
	providerResult                 *string
	createdAt                      time.Time
	updatedAt                      time.Time
}

func NewInternalOpening(walletID, playerID string, amount domain.Money, now time.Time) (WagerTransaction, error) {
	if walletID == "" || playerID == "" {
		return WagerTransaction{}, ErrOpeningFieldsRequired
	}

	id, err := uuid.NewV7()
	if err != nil {
		return WagerTransaction{}, ErrOpeningFieldsRequired
	}

	return WagerTransaction{
		id:        id.String(),
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
	referenceExternalTransactionID *string,
	now time.Time,
) (WagerTransaction, error) {
	if txType == TypeOpening {
		return WagerTransaction{}, ErrOpeningNotAllowedExternal
	}
	if id == "" || externalID == "" || provider == "" || idempotencyKey == "" || payloadHash == "" || walletID == "" || playerID == "" {
		return WagerTransaction{}, ErrMissingExternalFields
	}
	zeroMoney, err := domain.NewMoneyFromInt(0, amount.Currency())
	if err != nil {
		return WagerTransaction{}, ErrOpeningNotAllowedExternal
	}
	if txType == TypeLoss && !amount.Equals(zeroMoney) {
		return WagerTransaction{}, ErrKindLossWithoutZeroAmount
	}
	return WagerTransaction{
		id:                             id,
		externalID:                     &externalID,
		provider:                       &provider,
		idempotencyKey:                 &idempotencyKey,
		payloadHash:                    &payloadHash,
		walletID:                       walletID,
		playerID:                       playerID,
		roundID:                        roundID,
		gameID:                         gameID,
		txType:                         txType,
		amount:                         amount,
		referenceExternalTransactionID: referenceExternalTransactionID,
		state:                          StatePending,
		createdAt:                      now,
		updatedAt:                      now,
	}, nil
}

func RehydrateTransaction(
	id string,
	externalID, provider, idempotencyKey, payloadHash *string,
	walletID, playerID string,
	roundID, gameID *string,
	txType TransactionType,
	amount domain.Money,
	referenceExternalTransactionID, resolvedInternalRef *string,
	state TransactionState,
	failureCode, providerResult *string,
	createdAt, updatedAt time.Time,
) WagerTransaction {
	return WagerTransaction{
		id:                             id,
		externalID:                     externalID,
		provider:                       provider,
		idempotencyKey:                 idempotencyKey,
		payloadHash:                    payloadHash,
		walletID:                       walletID,
		playerID:                       playerID,
		roundID:                        roundID,
		gameID:                         gameID,
		txType:                         txType,
		amount:                         amount,
		referenceExternalTransactionID: referenceExternalTransactionID,
		resolvedInternalRef:            resolvedInternalRef,
		state:                          state,
		failureCode:                    failureCode,
		providerResult:                 providerResult,
		createdAt:                      createdAt,
		updatedAt:                      updatedAt,
	}
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
			return &ErrInvalidTransition{CurrentState: t.state, TargetState: newState}
		}
	case StatePendingReference:
		if newState != StateProcessed && newState != StateRejected && newState != StateFailed {
			return &ErrInvalidTransition{CurrentState: t.state, TargetState: newState}
		}
	default:
		return ErrUnexpectedOrTerminalState
	}

	t.state = newState
	t.failureCode = failureCode
	t.updatedAt = now
	return nil
}

func (t *WagerTransaction) SetResolvedInternalRef(internalRef string, now time.Time) {
	t.resolvedInternalRef = &internalRef
	t.updatedAt = now
}

func (t *WagerTransaction) SetProviderResult(result string, now time.Time) {
	t.providerResult = &result
	t.updatedAt = now
}

func (t WagerTransaction) ID() string              { return t.id }
func (t WagerTransaction) ExternalID() *string     { return t.externalID }
func (t WagerTransaction) Provider() *string       { return t.provider }
func (t WagerTransaction) IdempotencyKey() *string { return t.idempotencyKey }
func (t WagerTransaction) PayloadHash() *string    { return t.payloadHash }
func (t WagerTransaction) WalletID() string        { return t.walletID }
func (t WagerTransaction) PlayerID() string        { return t.playerID }
func (t WagerTransaction) RoundID() *string        { return t.roundID }
func (t WagerTransaction) GameID() *string         { return t.gameID }
func (t WagerTransaction) Type() TransactionType   { return t.txType }
func (t WagerTransaction) Amount() domain.Money    { return t.amount }
func (t WagerTransaction) ReferenceExternalTransactionID() *string {
	return t.referenceExternalTransactionID
}
func (t WagerTransaction) ResolvedInternalRef() *string { return t.resolvedInternalRef }
func (t WagerTransaction) State() TransactionState      { return t.state }
func (t WagerTransaction) FailureCode() *string         { return t.failureCode }
func (t WagerTransaction) ProviderResult() *string      { return t.providerResult }
func (t WagerTransaction) CreatedAt() time.Time         { return t.createdAt }
func (t WagerTransaction) UpdatedAt() time.Time         { return t.updatedAt }
