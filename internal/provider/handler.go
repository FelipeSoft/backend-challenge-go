package provider

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	getProviderWagerTransaction *GetProviderWagerTransaction
}

func NewHandler(getProviderWagerTransaction *GetProviderWagerTransaction) *Handler {
	return &Handler{
		getProviderWagerTransaction: getProviderWagerTransaction,
	}
}

func (s *Handler) GetExternalTransaction(c *gin.Context) {
	externalTransactionId := c.Param("externalTransactionId")
	if externalTransactionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "externalTransactionId is required"})
		return
	}
	providerId := c.Param("providerId")
	if providerId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "providerId is required"})
		return
	}
	externalTransaction, err := s.getProviderWagerTransaction.Execute(c.Request.Context(), externalTransactionId, providerId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, externalTransaction)
}
