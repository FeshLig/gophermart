package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FeshLig/gophermart/internal/handler"
	"github.com/FeshLig/gophermart/internal/model"
	"github.com/FeshLig/gophermart/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWithdrawalHandler_Withdraw(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := int64(1)

	t.Run("success withdraw", func(t *testing.T) {
		mockSvc := new(MockWithdrawalService)
		h := handler.NewWithdrawalHandler(mockSvc)

		mockSvc.On("Withdraw", mock.Anything, userID, "12345", 100.0).Return(nil).Once()

		r := gin.New()
		r.POST("/", func(c *gin.Context) {
			addUser(c, userID)
			h.Withdraw(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"order":"12345","sum":100}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		mockSvc := new(MockWithdrawalService)
		h := handler.NewWithdrawalHandler(mockSvc)

		mockSvc.On("Withdraw", mock.Anything, userID, "12345", 100.0).
			Return(service.ErrInsufficientFunds).Once()

		r := gin.New()
		r.POST("/", func(c *gin.Context) {
			addUser(c, userID)
			h.Withdraw(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"order":"12345","sum":100}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusPaymentRequired, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("invalid order number", func(t *testing.T) {
		mockSvc := new(MockWithdrawalService)
		h := handler.NewWithdrawalHandler(mockSvc)

		mockSvc.On("Withdraw", mock.Anything, userID, "bad", 100.0).
			Return(service.ErrInvalidOrderNumber).Once()

		r := gin.New()
		r.POST("/", func(c *gin.Context) {
			addUser(c, userID)
			h.Withdraw(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"order":"bad","sum":100}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusUnprocessableEntity, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mockSvc := new(MockWithdrawalService)
		h := handler.NewWithdrawalHandler(mockSvc)

		r := gin.New()
		r.POST("/", func(c *gin.Context) {
			addUser(c, userID)
			h.Withdraw(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{bad json}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("sum <= 0", func(t *testing.T) {
		mockSvc := new(MockWithdrawalService)
		h := handler.NewWithdrawalHandler(mockSvc)

		r := gin.New()
		r.POST("/", func(c *gin.Context) {
			addUser(c, userID)
			h.Withdraw(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"order":"123","sum":0}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestWithdrawalHandler_GetWithdrawals(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := int64(1)

	t.Run("success", func(t *testing.T) {
		mockSvc := new(MockWithdrawalService)
		h := handler.NewWithdrawalHandler(mockSvc)

		withdrawals := []model.Withdrawal{
			{OrderNumber: 123, UserID: userID, Sum: 100},
		}

		mockSvc.On("GetWithdrawals", mock.Anything, userID).Return(withdrawals, nil).Once()

		r := gin.New()
		r.GET("/", func(c *gin.Context) {
			addUser(c, userID)
			h.GetWithdrawals(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "123")

		mockSvc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		mockSvc := new(MockWithdrawalService)
		h := handler.NewWithdrawalHandler(mockSvc)

		mockSvc.On("GetWithdrawals", mock.Anything, userID).Return(nil, errors.New("fail")).Once()

		r := gin.New()
		r.GET("/", func(c *gin.Context) {
			addUser(c, userID)
			h.GetWithdrawals(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}
