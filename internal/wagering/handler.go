package wagering

import (
	"net/http"

	"github.com/FelipeSoft/jungle-gaming/internal/wagering/usecase"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service                Service
	createWagerTransaction usecase.CreateWagerTransaction
}

func NewHandler() *Handler {
	return &Handler{}
}

func (s *Handler) GetTransaction(c *gin.Context) {
	wagerTransactionId := c.Param("transactionId")
	if wagerTransactionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "transactionId is required",
		})
		return
	}
	wagerTransaction, err := s.service.GetWagerTransaction(wagerTransactionId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, wagerTransaction)
}

func (s *Handler) CreateWagerTransaction(c *gin.Context) {
	var req CreateWagerTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.createWagerTransaction.Execute(usecase.CreateWagerTransactionInput{
		ProviderID:                     req.ProviderID,
		ExternalTransactionID:          req.ExternalTransactionID,
		PlayerID:                       req.PlayerID,
		WalletID:                       req.WalletID,
		RoundID:                        req.RoundID,
		GameID:                         req.GameID,
		Kind:                           req.Kind,
		MoneyAmount:                    req.Money.Amount,
		MoneyCurrency:                  req.Money.Currency,
		ReferenceExternalTransactionId: req.ReferenceExternalTransactionId,
	})
	c.JSON(http.StatusOK, gin.H{
		"transactionId": "0192f298-345e-7e38-af88-e43f851a819d",
		"status":        "PROCESSED",
		"balance": gin.H{
			"amount":   "975.00",
			"currency": "BRL",
		},
		"idempotentReplay": false,
	})
}
