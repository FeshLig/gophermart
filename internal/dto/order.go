package dto

import (
	"strconv"
	"time"

	"github.com/FeshLig/gophermart/internal/model"
)

type Order struct {
	Number     string   `json:"number"`
	Status     string   `json:"status"`
	Accrual    *float64 `json:"accrual,omitempty"`
	UploadedAt string   `json:"uploaded_at"`
}

func ToOrderDTO(o model.Order) Order {
	return Order{
		Number:     strconv.FormatInt(o.Number, 10),
		Status:     string(o.Status),
		Accrual:    o.Accrual,
		UploadedAt: o.UploadedAt.Format(time.RFC3339),
	}
}

func ToOrdersDTO(orders []model.Order) []Order {
	result := make([]Order, 0, len(orders))

	for _, o := range orders {
		result = append(result, ToOrderDTO(o))
	}

	return result
}
