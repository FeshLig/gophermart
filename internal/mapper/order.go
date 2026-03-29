package mapper

import (
	"strconv"
	"time"

	"github.com/FeshLig/gophermart/internal/dto"
	"github.com/FeshLig/gophermart/internal/model"
)

func ToOrderDTO(o model.Order) dto.Order {
	return dto.Order{
		Number:     strconv.FormatInt(o.Number, 10),
		Status:     string(o.Status),
		Accrual:    o.Accrual,
		UploadedAt: o.UploadedAt.Format(time.RFC3339),
	}
}

func ToOrdersDTO(orders []model.Order) []dto.Order {
	result := make([]dto.Order, 0, len(orders))

	for _, o := range orders {
		result = append(result, ToOrderDTO(o))
	}

	return result
}
