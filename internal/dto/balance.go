package dto

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func ToBalanceDTO(current, withdrawn float64) Balance {
	return Balance{
		Current:   current,
		Withdrawn: withdrawn,
	}
}
