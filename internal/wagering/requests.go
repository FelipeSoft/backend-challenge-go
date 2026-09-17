package wagering

type CreateWagerTransactionRequest struct {
	ProviderID                     string  `json:"providerId" binding:"required"`
	ExternalTransactionID          string  `json:"externalTransactionId" binding:"required"`
	PlayerID                       string  `json:"playerId" binding:"required"`
	WalletID                       string  `json:"walletId" binding:"required"`
	RoundID                        string  `json:"roundId" binding:"required"`
	GameID                         string  `json:"gameId" binding:"required"`
	Kind                           string  `json:"kind" binding:"required"`
	ReferenceExternalTransactionId *string `json:"referenceExternalTransactionId"`
	Money                          struct {
		Amount   string `json:"amount" binding:"required"`
		Currency string `json:"currency" binding:"required"`
	} `json:"money" binding:"required"`
}
