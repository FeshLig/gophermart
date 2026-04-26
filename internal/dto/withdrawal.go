package dto

import (
	"strconv"
	"time"

	"github.com/FeshLig/gophermart/internal/model"
)

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type Withdrawal struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func ToWithdrawalDTO(w model.Withdrawal) Withdrawal {
	return Withdrawal{
		Order:       strconv.FormatInt(w.OrderNumber, 10),
		Sum:         w.Sum,
		ProcessedAt: w.ProcessedAt.Format(time.RFC3339),
	}
}

func ToWithdrawalsDTO(ws []model.Withdrawal) []Withdrawal {
	result := make([]Withdrawal, 0, len(ws))

	for _, w := range ws {
		result = append(result, ToWithdrawalDTO(w))
	}

	return result
}
