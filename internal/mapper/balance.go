package mapper

import "github.com/FeshLig/gophermart/internal/dto"

func ToBalanceDTO(current, withdrawn float64) dto.Balance {
	return dto.Balance{
		Current:   current,
		Withdrawn: withdrawn,
	}
}
