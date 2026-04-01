package main

import (
	"context"
	"fmt"
	"log"

	"github.com/FeshLig/gophermart/internal/accrual"
	"github.com/FeshLig/gophermart/internal/config"
	"github.com/FeshLig/gophermart/internal/handler"
	"github.com/FeshLig/gophermart/internal/http"
	"github.com/FeshLig/gophermart/internal/repository"
	"github.com/FeshLig/gophermart/internal/service"
	"github.com/FeshLig/gophermart/internal/worker"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func run() error {
	ctx := context.Background()

	// Загружаем конфиг
	cfg := config.GetConfig()

	// Репозиторий
	repo, err := repository.NewPostgres(ctx, string(cfg.DatabaseURI))
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
	accrualClient := accrual.NewClient(accrualURL)

	worker := worker.NewOrderWorker(repo, accrualClient)

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
	router := http.NewRouter(handlers, services.Auth.GetJWTSecret)

	addr := fmt.Sprintf("%s:%d", cfg.Address.Host, cfg.Address.Port)
	log.Printf("starting server at %s", addr)

	return router.Run(addr)
}
