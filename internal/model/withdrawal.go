package model

import "time"

type Withdrawal struct {
	ID          int64
	UserID      int64
	OrderNumber int64
	Sum         float64
	ProcessedAt time.Time
}
