package worker

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/FeshLig/gophermart/internal/accrual"
	"github.com/FeshLig/gophermart/internal/model"
	"github.com/FeshLig/gophermart/internal/repository"
)

type OrderWorker struct {
	repo     repository.OrderRepository
	client   *accrual.Client
	interval time.Duration
}

func NewOrderWorker(repo repository.OrderRepository, client *accrual.Client) *OrderWorker {
	return &OrderWorker{
		repo:     repo,
		client:   client,
		interval: 2 * time.Second,
	}
}

func (w *OrderWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *OrderWorker) process(ctx context.Context) {
	orders, err := w.repo.GetOrdersByStatus(ctx, model.StatusNew, model.StatusProcessing)
	if err != nil {
		log.Println("worker: get orders:", err)
		return
	}

	for _, order := range orders {
		w.handleOrder(ctx, order)
	}
}

func (w *OrderWorker) handleOrder(ctx context.Context, order model.Order) {
	number := strconv.FormatInt(order.Number, 10)

	resp, statusCode, err := w.client.GetOrder(ctx, number)
	if err != nil {
		log.Println("worker: accrual request error:", err)
		return
	}

	switch statusCode {
	case 204:
		return

	case 429:
		time.Sleep(time.Second)
		return
	}

	if resp == nil {
		return
	}

	newStatus := model.OrderStatus(resp.Status)

	err = w.repo.UpdateOrderStatus(ctx, order.Number, newStatus, resp.Accrual)
	if err != nil {
		log.Println("worker: update status error:", err)
	}
}
