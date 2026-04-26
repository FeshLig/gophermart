package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FeshLig/gophermart/internal/dto"
	"github.com/FeshLig/gophermart/internal/handler"
	"github.com/FeshLig/gophermart/internal/middleware"
	"github.com/FeshLig/gophermart/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockBalanceService struct {
	mock.Mock
}

func (m *MockBalanceService) GetBalance(ctx context.Context, userID int64) (dto.Balance, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(dto.Balance), args.Error(1)
}

type MockWithdrawalService struct {
	mock.Mock
}

func (m *MockWithdrawalService) Withdraw(ctx context.Context, userID int64, order string, sum float64) error {
	args := m.Called(ctx, userID, order, sum)
	return args.Error(0)
}

func (m *MockWithdrawalService) GetWithdrawals(ctx context.Context, userID int64) ([]model.Withdrawal, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Withdrawal), args.Error(1)
}

func addUserToContext(c *gin.Context, userID int64) {
	ctx := middleware.WithUserID(c.Request.Context(), userID)
	c.Request = c.Request.WithContext(ctx)
}

func TestBalanceHandler_GetBalance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockBalance := new(MockBalanceService)
		mockWithdrawal := new(MockWithdrawalService)

		handler := handler.NewBalanceHandler(mockBalance, mockWithdrawal)

		expected := dto.Balance{
			Current:   100,
			Withdrawn: 50,
		}

		mockBalance.On("GetBalance", mock.Anything, int64(1)).
			Return(expected, nil).Once()

		r := gin.New()
		r.GET("/balance", func(c *gin.Context) {
			addUserToContext(c, 1)
			handler.GetBalance(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/balance", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), `"current":100`)
		require.Contains(t, w.Body.String(), `"withdrawn":50`)

		mockBalance.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		mockBalance := new(MockBalanceService)
		mockWithdrawal := new(MockWithdrawalService)

		handler := handler.NewBalanceHandler(mockBalance, mockWithdrawal)

		mockBalance.On("GetBalance", mock.Anything, int64(1)).
			Return(dto.Balance{}, errors.New("fail")).Once()

		r := gin.New()
		r.GET("/balance", func(c *gin.Context) {
			addUserToContext(c, 1)
			handler.GetBalance(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/balance", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "fail")

		mockBalance.AssertExpectations(t)
	})
}

func TestBalanceHandler_GetWithdrawals(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockBalance := new(MockBalanceService)
		mockWithdrawal := new(MockWithdrawalService)

		handler := handler.NewBalanceHandler(mockBalance, mockWithdrawal)

		data := []model.Withdrawal{
			{UserID: 1, OrderNumber: 123, Sum: 50},
		}

		mockWithdrawal.On("GetWithdrawals", mock.Anything, int64(1)).
			Return(data, nil).Once()

		r := gin.New()
		r.GET("/withdrawals", func(c *gin.Context) {
			addUserToContext(c, 1)
			handler.GetWithdrawals(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "123")

		mockWithdrawal.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		mockBalance := new(MockBalanceService)
		mockWithdrawal := new(MockWithdrawalService)

		handler := handler.NewBalanceHandler(mockBalance, mockWithdrawal)

		mockWithdrawal.On("GetWithdrawals", mock.Anything, int64(1)).
			Return(nil, errors.New("fail")).Once()

		r := gin.New()
		r.GET("/withdrawals", func(c *gin.Context) {
			addUserToContext(c, 1)
			handler.GetWithdrawals(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		require.Contains(t, w.Body.String(), "fail")

		mockWithdrawal.AssertExpectations(t)
	})
}
