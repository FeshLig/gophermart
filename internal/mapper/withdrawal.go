package mapper

import (
	"strconv"
	"time"

	"github.com/FeshLig/gophermart/internal/dto"
	"github.com/FeshLig/gophermart/internal/model"
)

func ToWithdrawalDTO(w model.Withdrawal) dto.Withdrawal {
	return dto.Withdrawal{
		Order:       strconv.FormatInt(w.OrderNumber, 10),
		Sum:         w.Sum,
		ProcessedAt: w.ProcessedAt.Format(time.RFC3339),
	}
}

func ToWithdrawalsDTO(ws []model.Withdrawal) []dto.Withdrawal {
	result := make([]dto.Withdrawal, 0, len(ws))

	for _, w := range ws {
		result = append(result, ToWithdrawalDTO(w))
	}

	return result
}
