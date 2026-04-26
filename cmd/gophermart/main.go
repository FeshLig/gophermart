package main

import (
	"context"
	"fmt"
	"time"

	"github.com/FeshLig/gophermart/internal/accrual"
	"github.com/FeshLig/gophermart/internal/config"
	"github.com/FeshLig/gophermart/internal/handler"
	"github.com/FeshLig/gophermart/internal/http"
	"github.com/FeshLig/gophermart/internal/repository"
	"github.com/FeshLig/gophermart/internal/service"
	"github.com/FeshLig/gophermart/internal/worker"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err := run(logger); err != nil {
		logger.Fatal("server failed", zap.Error(err))
	}
}

func run(logger *zap.Logger) error {
	ctx := context.Background()

	// Загружаем конфиг
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Репозиторий
	repo, err := repository.NewPostgres(ctx, string(cfg.DatabaseURI), logger)
	if err != nil {
		return fmt.Errorf("failed to init repository: %w", err)
	}
	defer repo.Close()

	err = repository.RunMigrations(string(cfg.DatabaseURI))
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	accrualURL := cfg.AccuralAddress.URL()
	if accrualURL == "" {
		return fmt.Errorf("accrual system address is not set")
	}
	accrualClient := accrual.NewClient(accrualURL, 5*time.Second)

	worker := worker.NewOrderWorker(repo, accrualClient, logger)

	go worker.Start(ctx)

	// Сервисы
	services := service.NewServices(repo)

	// Хендлеры
	handlers := handler.NewHandlers(
		services.Auth,
		services.Order,
		services.Balance,
		services.Withdrawal,
	)

	// Роутер
	router := http.NewRouter(handlers, services.Auth.GetJWTSecret, logger)

	addr := fmt.Sprintf("%s:%d", cfg.Address.Host, cfg.Address.Port)
	logger.Info("starting server",
		zap.String("addr", addr),
	)

	return router.Run(addr)
}
