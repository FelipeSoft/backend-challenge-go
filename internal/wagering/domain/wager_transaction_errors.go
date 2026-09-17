package domain

import (
	"errors"
	"fmt"
)

var (
	ErrOpeningFieldsRequired     = errors.New("id, walletID and playerID are required for opening transaction")
	ErrOpeningNotAllowedExternal = errors.New("OPENING type is reserved for internal wallet creation and cannot be created via external channels")
	ErrMissingExternalFields     = errors.New("missing mandatory fields for external transaction")
	ErrUnexpectedOrTerminalState = errors.New("current state is unexpected or terminal")
)

type ErrTerminalTransition struct {
	CurrentState TransactionState
	TargetState  TransactionState
}

func (e *ErrTerminalTransition) Error() string {
	return fmt.Sprintf("cannot transition from terminal state %s to %s", e.CurrentState, e.TargetState)
}

func (e *ErrTerminalTransition) Is(target error) bool {
	_, ok := target.(*ErrTerminalTransition)
	return ok
}

type ErrInvalidTransition struct {
	CurrentState TransactionState
	TargetState  TransactionState
}

func (e *ErrInvalidTransition) Error() string {
	return fmt.Sprintf("invalid state transition from %s to %s", e.CurrentState, e.TargetState)
}

func (e *ErrInvalidTransition) Is(target error) bool {
	_, ok := target.(*ErrInvalidTransition)
	return ok
}