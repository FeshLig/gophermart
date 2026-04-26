package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FeshLig/gophermart/internal/handler"
	"github.com/FeshLig/gophermart/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, login, password string) (string, error) {
	args := m.Called(ctx, login, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, login, password string) (string, error) {
	args := m.Called(ctx, login, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GetJWTSecret() string {
	return "secret"
}

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		handler := handler.NewAuthHandler(mockSvc)

		mockSvc.On("Register", mock.Anything, "user", "pass").
			Return("token123", nil).Once()

		r := gin.New()
		r.POST("/register", handler.Register)

		req := httptest.NewRequest(http.MethodPost, "/register",
			strings.NewReader(`{"login":"user","password":"pass"}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		require.Equal(t, "token", cookies[0].Name)

		mockSvc.AssertExpectations(t)
	})

	t.Run("bad json", func(t *testing.T) {
		handler := handler.NewAuthHandler(new(MockAuthService))

		r := gin.New()
		r.POST("/register", handler.Register)

		req := httptest.NewRequest(http.MethodPost, "/register",
			strings.NewReader(`invalid json`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("user exists", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		handler := handler.NewAuthHandler(mockSvc)

		mockSvc.On("Register", mock.Anything, "user", "pass").
			Return("", service.ErrUserAlreadyExists).Once()

		r := gin.New()
		r.POST("/register", handler.Register)

		req := httptest.NewRequest(http.MethodPost, "/register",
			strings.NewReader(`{"login":"user","password":"pass"}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusConflict, w.Code)

		mockSvc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		handler := handler.NewAuthHandler(mockSvc)

		mockSvc.On("Register", mock.Anything, "user", "pass").
			Return("", errors.New("db error")).Once()

		r := gin.New()
		r.POST("/register", handler.Register)

		req := httptest.NewRequest(http.MethodPost, "/register",
			strings.NewReader(`{"login":"user","password":"pass"}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)

		mockSvc.AssertExpectations(t)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		handler := handler.NewAuthHandler(mockSvc)

		mockSvc.On("Login", mock.Anything, "user", "pass").
			Return("token123", nil).Once()

		r := gin.New()
		r.POST("/login", handler.Login)

		req := httptest.NewRequest(http.MethodPost, "/login",
			strings.NewReader(`{"login":"user","password":"pass"}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.NotEmpty(t, w.Result().Cookies())

		mockSvc.AssertExpectations(t)
	})

	t.Run("bad json", func(t *testing.T) {
		handler := handler.NewAuthHandler(new(MockAuthService))

		r := gin.New()
		r.POST("/login", handler.Login)

		req := httptest.NewRequest(http.MethodPost, "/login",
			strings.NewReader(`bad json`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		handler := handler.NewAuthHandler(mockSvc)

		mockSvc.On("Login", mock.Anything, "user", "pass").
			Return("", service.ErrInvalidCredentials).Once()

		r := gin.New()
		r.POST("/login", handler.Login)

		req := httptest.NewRequest(http.MethodPost, "/login",
			strings.NewReader(`{"login":"user","password":"pass"}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)

		mockSvc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		mockSvc := new(MockAuthService)
		handler := handler.NewAuthHandler(mockSvc)

		mockSvc.On("Login", mock.Anything, "user", "pass").
			Return("", errors.New("db error")).Once()

		r := gin.New()
		r.POST("/login", handler.Login)

		req := httptest.NewRequest(http.MethodPost, "/login",
			strings.NewReader(`{"login":"user","password":"pass"}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)

		mockSvc.AssertExpectations(t)
	})
}
