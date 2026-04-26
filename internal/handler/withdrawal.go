package handler

import (
	"net/http"

	"github.com/FeshLig/gophermart/internal/dto"
	"github.com/FeshLig/gophermart/internal/middleware"
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

func (h *WithdrawalHandler) Withdraw(c *gin.Context) {
	userID, exists := middleware.UserIDFromContext(c.Request.Context())
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

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

func (h *WithdrawalHandler) GetWithdrawals(c *gin.Context) {
	userID, exists := middleware.UserIDFromContext(c.Request.Context())
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	withdrawals, err := h.withdrawalService.GetWithdrawals(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToWithdrawalsDTO(withdrawals))
}
