package http

import (
	"net/http"

	"github.com/FeshLig/gophermart/internal/handler"
	"github.com/FeshLig/gophermart/internal/middleware"
	"github.com/gin-gonic/gin"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter(handlers *handler.Handlers, getJWTSecret func() string) *Router {
	r := gin.New()

	// Глобальные middleware
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.Gzip())

	// Healthcheck
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// --- API ---
	api := r.Group("/api")

	// --- Публичные маршруты (без auth) ---
	api.POST("/user/register", handlers.Auth.Register)
	api.POST("/user/login", handlers.Auth.Login)

	// --- Middleware авторизации ---
	authMiddleware := middleware.AuthMiddleware(getJWTSecret)

	// --- Защищённые маршруты ---
	auth := api.Group("/user")
	auth.Use(authMiddleware)

	// Orders
	auth.POST("/orders", handlers.Order.UploadOrder) // /api/user/orders
	auth.GET("/orders", handlers.Order.GetOrders)    // /api/user/orders

	// Balance
	auth.GET("/balance", handlers.Balance.GetBalance)            // /api/user/balance
	auth.POST("/balance/withdraw", handlers.Withdrawal.Withdraw) // /api/user/balance/withdraw

	// Withdrawals
	auth.GET("/withdrawals", handlers.Balance.GetWithdrawals) // /api/user/withdrawals

	return &Router{
		engine: r,
	}
}

func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
