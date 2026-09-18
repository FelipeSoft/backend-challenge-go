package wallet

type CreateWalletRequest struct {
	PlayerID       string `json:"playerId" binding:"required"`
	InitialBalance struct {
		Amount   string `json:"amount" binding:"required"`
		Currency string `json:"currency" binding:"required"`
	} `json:"initialBalance" binding:"required"`
}