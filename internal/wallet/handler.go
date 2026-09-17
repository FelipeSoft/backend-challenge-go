package wallet

import (
	"net/http"
	"time"

	"github.com/FelipeSoft/jungle-gaming/internal/wallet/query"
	"github.com/FelipeSoft/jungle-gaming/internal/wallet/usecase"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	getWallet       query.GetWallet
	getWalletLedger query.GetWalletLedger
	createWallet    usecase.CreateWallet
	reconcileWallet usecase.ReconcileWallet
}

func NewHandler() *Handler {
	return &Handler{}
}

func (s *Handler) GetWallet(c *gin.Context) {
	walletId := c.Param("walletId")
	if walletId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "walletId is required",
		})
		return
	}
	output, err := s.getWallet.Execute(c.Request.Context(), walletId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, output)
}

func (s *Handler) GetWalletLedger(c *gin.Context) {
	walletId := c.Param("walletId")
	if walletId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "walletId is required",
		})
		return
	}
	output, err := s.getWalletLedger.Execute(
		c.Request.Context(),
		walletId, 
		c.Query("cursor"),
		c.Query("limit"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, output)
}

func (s *Handler) CreateWallet(c *gin.Context) {
	requestArrivalTime := time.Now().UTC()
	var req CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	output, err := s.createWallet.Execute(c.Request.Context(), usecase.CreateWalletInput{
		PlayerID:               req.PlayerID,
		InitialBalanceAmount:   req.InitialBalance.Amount,
		InitialBalanceCurrency: req.InitialBalance.Currency,
		Now:                    requestArrivalTime,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":       output.WalletID,
		"playerId": output.PlayerID,
		"balance": gin.H{
			"amount":   output.BalanceAmount,
			"currency": output.BalanceCurrency,
		},
		"version": 1,
	})
}

func (s *Handler) ReconcileWallet(c *gin.Context) {
	walletId := c.Param("walletId")
	if walletId == "" {
		c.JSON(400, gin.H{
			"error": "walletId is required",
		})
		return
	}
	output, err := s.reconcileWallet.Execute(usecase.ReconcileWalletInput{
		WalletID: walletId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"walletId": walletId,
		"storedBalance": gin.H{
			"amount":   output.StoredBalanceAmount,
			"currency": output.StoredBalanceCurrency,
		},
		"calculatedBalance": gin.H{
			"amount":   output.CalculatedBalanceAmount,
			"currency": output.CalculatedBalanceCurrency,
		},
		"difference": gin.H{
			"amount":   output.DifferenceAmount,
			"currency": output.DifferenceCurrency,
		},
		"consistent":     output.Consistent,
		"checkedEntries": output.CheckedEntries,
	})
}
