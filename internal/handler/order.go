package handler

import (
	"io"
	"net/http"

	"github.com/FeshLig/gophermart/internal/mapper"
	"github.com/FeshLig/gophermart/internal/service"
	"github.com/gin-gonic/gin"
)

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
