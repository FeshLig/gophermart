package handler

import (
	"net/http"

	"github.com/FeshLig/gophermart/internal/dto"
	"github.com/FeshLig/gophermart/internal/middleware"
	"github.com/FeshLig/gophermart/internal/service"
	"github.com/gin-gonic/gin"
)

type BalanceHandler struct {
	balanceService    service.BalanceService
	withdrawalService service.WithdrawalService
}

func NewBalanceHandler(balanceService service.BalanceService, withdrawalService service.WithdrawalService) *BalanceHandler {
	return &BalanceHandler{
		balanceService:    balanceService,
		withdrawalService: withdrawalService,
	}
}

// GetBalance возвращает текущий баланс пользователя
func (h *BalanceHandler) GetBalance(c *gin.Context) {
	userID, exists := middleware.UserIDFromContext(c.Request.Context())
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	balance, err := h.balanceService.GetBalance(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, balance)
}

// GetWithdrawals возвращает историю снятий пользователя
func (h *BalanceHandler) GetWithdrawals(c *gin.Context) {
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
