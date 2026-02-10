package service

import (
	"context"

	"github.com/makhtech/management/internal/domain/models"
)

// PlanService интерфейс для работы с планами
type PlanService interface {
	Create(ctx context.Context, req *models.CreatePlanRequest) (*models.Plan, error)
	GetByID(ctx context.Context, id int32) (*models.Plan, error)
	Update(ctx context.Context, req *models.UpdatePlanRequest) (*models.Plan, error)
	Delete(ctx context.Context, id int32) error
	List(ctx context.Context, activeOnly bool) ([]*models.Plan, error)
}
