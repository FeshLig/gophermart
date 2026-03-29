package handler

import (
	"net/http"

	"github.com/FeshLig/gophermart/internal/dto"
	"github.com/FeshLig/gophermart/internal/mapper"
	"github.com/FeshLig/gophermart/internal/service"
	"github.com/gin-gonic/gin"
)

type WithdrawalHandler struct {
	withdrawalService service.WithdrawalService
}

func NewWithdrawalHandler(ws service.WithdrawalService) *WithdrawalHandler {
	return &WithdrawalHandler{
		withdrawalService: ws,
	}
}

// POST /api/withdraw
func (h *WithdrawalHandler) Withdraw(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(int64)

	var req dto.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Sum <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sum must be positive"})
		return
	}

	err := h.withdrawalService.Withdraw(c.Request.Context(), userID, req.Order, req.Sum)
	if err != nil {
		switch err {
		case service.ErrInsufficientFunds:
			c.JSON(http.StatusPaymentRequired, gin.H{"error": err.Error()})
		case service.ErrInvalidOrderNumber:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusOK)
}

// GET /api/withdrawals
func (h *WithdrawalHandler) GetWithdrawals(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(int64)

	withdrawals, err := h.withdrawalService.GetWithdrawals(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapper.ToWithdrawalsDTO(withdrawals))
}
