package provider

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(Service Service) *Handler {
	return &Handler{
		service: Service,
	}
}

func (s *Handler) GetExternalTransaction(c *gin.Context) {
	externalTransactionId := c.Param("externalTransactionId")
	if externalTransactionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "externalTransactionId is required",
		})
		return
	}
	externalTransaction, err := s.service.GetExternalTransaction(externalTransactionId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, externalTransaction)
}
