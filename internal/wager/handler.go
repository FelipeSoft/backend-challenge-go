package wager

import (
	"net/http"

	"github.com/FelipeSoft/backend-challenge-go/internal/wager/query"
	"github.com/FelipeSoft/backend-challenge-go/internal/wager/usecase"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	getWagerTransaction    *query.GetWagerTransaction
	createWagerTransaction *usecase.CreateWagerTransaction
}

func NewHandler(
	getWagerTransaction *query.GetWagerTransaction,
	createWagerTransaction *usecase.CreateWagerTransaction,
) *Handler {
	return &Handler{
		getWagerTransaction:    getWagerTransaction,
		createWagerTransaction: createWagerTransaction,
	}
}

func (s *Handler) GetTransaction(c *gin.Context) {
	wagerTransactionId := c.Param("transactionId")
	if wagerTransactionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "transactionId is required",
		})
		return
	}
	wagerTransaction, err := s.getWagerTransaction.Execute(c.Request.Context(), wagerTransactionId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, wagerTransaction)
}

func (s *Handler) CreateWagerTransaction(c *gin.Context) {
	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Idempotency-Key header is required",
		})
		return
	}
	var req CreateWagerTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	output, err := s.createWagerTransaction.Execute(c.Request.Context(), usecase.CreateWagerTransactionInput{
		ProviderID:                     req.ProviderID,
		ExternalTransactionID:          req.ExternalTransactionID,
		PlayerID:                       req.PlayerID,
		WalletID:                       req.WalletID,
		RoundID:                        &req.RoundID,
		GameID:                         &req.GameID,
		Kind:                           req.Kind,
		MoneyAmount:                    req.Money.Amount,
		MoneyCurrency:                  req.Money.Currency,
		IdempotencyKey:                 idempotencyKey,
		ReferenceExternalTransactionId: req.ReferenceExternalTransactionId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, output)
}
