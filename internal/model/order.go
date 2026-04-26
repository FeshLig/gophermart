package model

import "time"

type OrderStatus string

const (
	StatusNew        OrderStatus = "NEW"
	StatusProcessing OrderStatus = "PROCESSING"
	StatusInvalid    OrderStatus = "INVALID"
	StatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	Number     int64
	UserID     int64
	Status     OrderStatus
	Accrual    *float64
	UploadedAt time.Time
}
