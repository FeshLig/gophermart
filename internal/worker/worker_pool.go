package worker

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/FeshLig/gophermart/internal/accrual"
	"github.com/FeshLig/gophermart/internal/model"
	"github.com/FeshLig/gophermart/internal/repository"
	"go.uber.org/zap"
)

type AccrualClient interface {
	GetOrder(ctx context.Context, number string) (*accrual.Response, int, error)
}

type OrderWorkerPool struct {
	repo   repository.OrderRepository
	client AccrualClient
	logger *zap.Logger

	workers  int
	interval time.Duration

	jobs    chan model.Order
	pauseCh chan time.Duration
}

func NewOrderWorker(
	repo repository.OrderRepository,
	client AccrualClient,
	logger *zap.Logger,
) *OrderWorkerPool {

	return &OrderWorkerPool{
		repo:     repo,
		client:   client,
		interval: 2 * time.Second,
		workers:  5,
		jobs:     make(chan model.Order, 100),
		pauseCh:  make(chan time.Duration, 1),
		logger:   logger,
	}
}

func (p *OrderWorkerPool) Start(ctx context.Context) {
	var wg sync.WaitGroup

	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.Worker(ctx)
		}()
	}

	go p.Produce(ctx)

	wg.Wait()
}

func (p *OrderWorkerPool) Produce(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			close(p.jobs)
			return

		case <-ticker.C:
			orders, err := p.repo.GetOrdersByStatus(ctx, model.StatusNew, model.StatusProcessing)
			if err != nil {
				p.logger.Error("worker: get orders failed", zap.Error(err))
				continue
			}

			for _, o := range orders {
				select {
				case p.jobs <- o:
				case <-ctx.Done():
					close(p.jobs)
					return
				}
			}
		}
	}
}

func (p *OrderWorkerPool) Worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case dur := <-p.pauseCh:
			timer := time.NewTimer(dur)

			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}

		case order, ok := <-p.jobs:
			if !ok {
				return
			}

			p.Handle(ctx, order)
		}
	}
}

func (p *OrderWorkerPool) Handle(ctx context.Context, order model.Order) {
	number := strconv.FormatInt(order.Number, 10)

	p.logger.Debug("processing order",
		zap.Int64("order_number", order.Number),
	)

	resp, statusCode, err := p.client.GetOrder(ctx, number)
	if err != nil {
		p.logger.Error("worker: accrual request error",
			zap.Int64("order_number", order.Number),
			zap.Error(err),
		)
		return
	}

	switch statusCode {
	case http.StatusNoContent:
		return

	case http.StatusTooManyRequests:
		p.logger.Warn("worker: rate limited by accrual",
			zap.Int64("order_number", order.Number),
		)

		select {
		case p.pauseCh <- time.Second:
		default:
		}

		return
	}

	if resp == nil {
		return
	}

	newStatus := model.OrderStatus(resp.Status)

	err = p.repo.UpdateOrderStatus(ctx, order.Number, newStatus, resp.Accrual)
	if err != nil {
		p.logger.Error("worker: update status failed",
			zap.Int64("order_number", order.Number),
			zap.String("status", string(newStatus)),
			zap.Error(err),
		)
		return
	}

	var accrual float64
	if resp.Accrual != nil {
		accrual = *resp.Accrual
	}

	p.logger.Info("order updated",
		zap.Int64("order_number", order.Number),
		zap.String("status", string(newStatus)),
		zap.Float64("accrual", accrual),
	)
}

func (p *OrderWorkerPool) PauseCh() <-chan time.Duration {
	return p.pauseCh
}

func (p *OrderWorkerPool) SetInterval(d time.Duration) {
	p.interval = d
}
