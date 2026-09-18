package queue

import (
	"time"

	"github.com/FelipeSoft/backend-challenge-go/internal/domain"
)

type SQSMessageData struct {
	ProviderID                     string       `json:"providerId"`
	ExternalTransactionID          string       `json:"externalTransactionId"`
	IdempotencyKey                 string       `json:"idempotencyKey"`
	PlayerID                       string       `json:"playerId"`
	WalletID                       string       `json:"walletId"`
	RoundID                        *string      `json:"roundId"`
	GameID                         *string      `json:"gameId"`
	Kind                           string       `json:"kind"`
	Money                          domain.Money `json:"money"`
	ReferenceExternalTransactionId *string      `json:"referenceExternalTransactionId,omitempty"`
}

type SQSMessageEnvelope struct {
	MessageID  string         `json:"messageId"`
	Type       string         `json:"type"`
	OccurredAt time.Time      `json:"occurredAt"`
	Data       SQSMessageData `json:"data"`
}
