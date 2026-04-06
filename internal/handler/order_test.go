package handler_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FeshLig/gophermart/internal/handler"
	"github.com/FeshLig/gophermart/internal/mapper"
	"github.com/FeshLig/gophermart/internal/model"
	"github.com/FeshLig/gophermart/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) UploadOrder(ctx context.Context, userID int64, number string) error {
	args := m.Called(ctx, userID, number)
	return args.Error(0)
}

func (m *MockOrderService) GetOrders(ctx context.Context, userID int64) ([]model.Order, error) {
	args := m.Called(ctx, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]model.Order), args.Error(1)
}

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) UploadOrder(c *gin.Context) {
	userID := c.GetInt64("userID")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	orderNumber := string(body)

	err = h.orderService.UploadOrder(c.Request.Context(), userID, orderNumber)
	if err != nil {
		switch err {
		case service.ErrInvalidOrderNumber:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case service.ErrOrderAlreadyExists:
			c.JSON(http.StatusOK, gin.H{"message": "order already exists"})
		case service.ErrOrderConflict:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusAccepted)
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
	userID := c.GetInt64("userID")

	orders, err := h.orderService.GetOrders(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapper.ToOrdersDTO(orders))
}

func addUser(c *gin.Context, id int64) {
	c.Set("userID", id)
}

func TestOrderHandler_UploadOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := int64(1)

	t.Run("empty body", func(t *testing.T) {
		handler := handler.NewOrderHandler(new(MockOrderService))

		r := gin.New()
		r.POST("/", func(c *gin.Context) {
			addUser(c, userID)
			handler.UploadOrder(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		mockSvc := new(MockOrderService)
		handler := handler.NewOrderHandler(mockSvc)

		mockSvc.On("UploadOrder", mock.Anything, userID, "12345").
			Return(nil).Once()

		r := gin.New()
		r.POST("/", func(c *gin.Context) {
			addUser(c, userID)
			handler.UploadOrder(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("12345"))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusAccepted, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("invalid order number", func(t *testing.T) {
		mockSvc := new(MockOrderService)
		handler := handler.NewOrderHandler(mockSvc)

		mockSvc.On("UploadOrder", mock.Anything, userID, "bad").
			Return(service.ErrInvalidOrderNumber).Once()

		r := gin.New()
		r.POST("/", func(c *gin.Context) {
			addUser(c, userID)
			handler.UploadOrder(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("bad"))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnprocessableEntity, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("already exists", func(t *testing.T) {
		mockSvc := new(MockOrderService)
		handler := handler.NewOrderHandler(mockSvc)

		mockSvc.On("UploadOrder", mock.Anything, userID, "123").
			Return(service.ErrOrderAlreadyExists).Once()

		r := gin.New()
		r.POST("/", func(c *gin.Context) {
			addUser(c, userID)
			handler.UploadOrder(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("123"))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("conflict", func(t *testing.T) {
		mockSvc := new(MockOrderService)
		handler := handler.NewOrderHandler(mockSvc)

		mockSvc.On("UploadOrder", mock.Anything, userID, "123").
			Return(service.ErrOrderConflict).Once()

		r := gin.New()
		r.POST("/", func(c *gin.Context) {
			addUser(c, userID)
			handler.UploadOrder(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("123"))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusConflict, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		mockSvc := new(MockOrderService)
		handler := handler.NewOrderHandler(mockSvc)

		mockSvc.On("UploadOrder", mock.Anything, userID, "123").
			Return(errors.New("db error")).Once()

		r := gin.New()
		r.POST("/", func(c *gin.Context) {
			addUser(c, userID)
			handler.UploadOrder(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("123"))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestOrderHandler_GetOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := int64(1)

	t.Run("success", func(t *testing.T) {
		mockSvc := new(MockOrderService)
		handler := handler.NewOrderHandler(mockSvc)

		orders := []model.Order{
			{Number: 123, UserID: userID},
		}

		mockSvc.On("GetOrders", mock.Anything, userID).
			Return(orders, nil).Once()

		r := gin.New()
		r.GET("/", func(c *gin.Context) {
			addUser(c, userID)
			handler.GetOrders(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), "123")

		mockSvc.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc := new(MockOrderService)
		handler := handler.NewOrderHandler(mockSvc)

		mockSvc.On("GetOrders", mock.Anything, userID).
			Return(nil, errors.New("fail")).Once()

		r := gin.New()
		r.GET("/", func(c *gin.Context) {
			addUser(c, userID)
			handler.GetOrders(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)

		mockSvc.AssertExpectations(t)
	})
}
