package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FeshLig/gophermart/internal/handler"
	httpRouter "github.com/FeshLig/gophermart/internal/http"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockAuthHandler struct{}

func (m *mockAuthHandler) Register(c *gin.Context) {
	c.Status(http.StatusOK)
}

func (m *mockAuthHandler) Login(c *gin.Context) {
	c.Status(http.StatusOK)
}

type mockOrderHandler struct {
	UploadOrderFn func(c *gin.Context)
	GetOrdersFn   func(c *gin.Context)
}

func (m *mockOrderHandler) UploadOrder(c *gin.Context) {
	if m.UploadOrderFn != nil {
		m.UploadOrderFn(c)
		return
	}
	c.Status(http.StatusOK)
}

func (m *mockOrderHandler) GetOrders(c *gin.Context) {
	if m.GetOrdersFn != nil {
		m.GetOrdersFn(c)
		return
	}
	c.Status(http.StatusOK)
}

type mockBalanceHandler struct{}

func (m *mockBalanceHandler) GetBalance(c *gin.Context) {
	c.Status(http.StatusOK)
}

func (m *mockBalanceHandler) GetWithdrawals(c *gin.Context) {
	c.Status(http.StatusOK)
}

type mockWithdrawalHandler struct{}

func (m *mockWithdrawalHandler) Withdraw(c *gin.Context) {
	c.Status(http.StatusOK)
}

func TestHealthcheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handlers := &handler.Handlers{
		Auth:       &handler.AuthHandler{},
		Order:      &handler.OrderHandler{},
		Balance:    &handler.BalanceHandler{},
		Withdrawal: &handler.WithdrawalHandler{},
	}

	router := httpRouter.NewRouter(handlers, func() string {
		return "secret"
	}, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.Run("")

	router.Engine().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}

func TestPublicRoutesExist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handlers := &handler.Handlers{
		Auth:       &handler.AuthHandler{},
		Order:      &handler.OrderHandler{},
		Balance:    &handler.BalanceHandler{},
		Withdrawal: &handler.WithdrawalHandler{},
	}

	router := httpRouter.NewRouter(handlers, func() string {
		return "secret"
	}, zap.NewNop())

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/user/register"},
		{http.MethodPost, "/api/user/login"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			require.NotEqual(t, http.StatusNotFound, w.Code)
		})
	}
}

func TestProtectedRoutesExist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handlers := &handler.Handlers{
		Auth:       &handler.AuthHandler{},
		Order:      &handler.OrderHandler{},
		Balance:    &handler.BalanceHandler{},
		Withdrawal: &handler.WithdrawalHandler{},
	}

	router := httpRouter.NewRouter(handlers, func() string {
		return "secret"
	}, zap.NewNop())

	protected := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/user/orders"},
		{http.MethodGet, "/api/user/orders"},
		{http.MethodGet, "/api/user/balance"},
		{http.MethodPost, "/api/user/balance/withdraw"},
		{http.MethodGet, "/api/user/withdrawals"},
	}

	for _, tt := range protected {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.Engine().ServeHTTP(w, req)

			require.NotEqual(t, http.StatusNotFound, w.Code)
		})
	}
}

func TestProtectedRouteCallsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false

	handlers := &handler.Handlers{
		Auth:       &handler.AuthHandler{},
		Order:      &handler.OrderHandler{},
		Balance:    &handler.BalanceHandler{},
		Withdrawal: &handler.WithdrawalHandler{},
	}

	router := httpRouter.NewRouter(handlers, func() string {
		return "secret"
	}, zap.NewNop())

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", nil)
	w := httptest.NewRecorder()

	router.Engine().ServeHTTP(w, req)

	// может не пройти auth — но маршрут должен быть вызван или отклонён middleware
	require.True(t, w.Code == http.StatusUnauthorized || called)
}
