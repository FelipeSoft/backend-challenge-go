package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type CanonicalPayload struct {
	ProviderID                     string         `json:"providerId"`
	ExternalTransactionID          string         `json:"externalTransactionId"`
	PlayerID                       string         `json:"playerID"`
	WalletID                       string         `json:"walletID"`
	RoundID                        *string        `json:"roundId,omitempty"`
	GameID                         *string        `json:"gameId,omitempty"`
	Kind                           string         `json:"kind"`
	Money                          CanonicalMoney `json:"money"`
	ReferenceExternalTransactionID *string        `json:"referenceExternalTransactionId,omitempty"`
}

type CanonicalMoney struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

func ComputePayloadHash(
	providerID, externalTransactionID, playerID, walletID, kind string,
	roundID, gameID, referenceExternalTransactionID *string,
	amountAmount int64,
	amountCurrency string,
) (string, error) {
	canonical := CanonicalPayload{
		ProviderID:            providerID,
		ExternalTransactionID: externalTransactionID,
		PlayerID:              playerID,
		WalletID:              walletID,
		RoundID:               roundID,
		GameID:                gameID,
		Kind:                  kind,
		Money: CanonicalMoney{
			Amount:   amountAmount,
			Currency: amountCurrency,
		},
		ReferenceExternalTransactionID: referenceExternalTransactionID,
	}
	bytesData, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(bytesData)
	return hex.EncodeToString(hash[:]), nil
}
