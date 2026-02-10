package models

import "time"

// Plan - доменная модель тарифного плана
type Plan struct {
	ID         int32
	Name       string
	CPU        int32
	RAMMB      int32
	DiskGB     int32
	PriceMonth float64
	IsActive   bool
	CreatedAt  time.Time
}

// CreatePlanRequest - запрос на создание плана
type CreatePlanRequest struct {
	Name       string
	CPU        int32
	RAMMB      int32
	DiskGB     int32
	PriceMonth float64
}

// UpdatePlanRequest - запрос на обновление плана
type UpdatePlanRequest struct {
	ID         int32
	Name       *string
	CPU        *int32
	RAMMB      *int32
	DiskGB     *int32
	PriceMonth *float64
	IsActive   *bool
}
