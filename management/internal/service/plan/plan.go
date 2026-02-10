package plan

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/makhtech/management/internal/domain/models"
	"github.com/makhtech/management/internal/repository"
)

// Service - сервис для работы с планами
type Service struct {
	planRepo repository.PlanRepository
	log      *slog.Logger
}

// New создает новый сервис планов
func New(planRepo repository.PlanRepository, log *slog.Logger) *Service {
	return &Service{
		planRepo: planRepo,
		log:      log,
	}
}

// Create создает новый план
func (s *Service) Create(ctx context.Context, req *models.CreatePlanRequest) (*models.Plan, error) {
	const op = "service.plan.Create"

	log := s.log.With(slog.String("op", op), slog.String("name", req.Name))
	log.Info("creating new plan")

	// Валидация
	if req.Name == "" {
		return nil, fmt.Errorf("%s: name is required", op)
	}
	if req.CPU <= 0 {
		return nil, fmt.Errorf("%s: cpu must be positive", op)
	}
	if req.RAMMB <= 0 {
		return nil, fmt.Errorf("%s: ram_mb must be positive", op)
	}
	if req.DiskGB <= 0 {
		return nil, fmt.Errorf("%s: disk_gb must be positive", op)
	}
	if req.PriceMonth < 0 {
		return nil, fmt.Errorf("%s: price_month must be non-negative", op)
	}

	plan, err := s.planRepo.Create(ctx, req)
	if err != nil {
		log.Error("failed to create plan", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("plan created successfully", slog.Int("id", int(plan.ID)))
	return plan, nil
}

// GetByID получает план по ID
func (s *Service) GetByID(ctx context.Context, id int32) (*models.Plan, error) {
	const op = "service.plan.GetByID"

	log := s.log.With(slog.String("op", op), slog.Int("id", int(id)))
	log.Debug("getting plan by id")

	plan, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrPlanNotFound) {
			log.Warn("plan not found")
			return nil, repository.ErrPlanNotFound
		}
		log.Error("failed to get plan", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return plan, nil
}

// Update обновляет существующий план
func (s *Service) Update(ctx context.Context, req *models.UpdatePlanRequest) (*models.Plan, error) {
	const op = "service.plan.Update"

	log := s.log.With(slog.String("op", op), slog.Int("id", int(req.ID)))
	log.Info("updating plan")

	// Валидация ID
	if req.ID <= 0 {
		return nil, fmt.Errorf("%s: invalid plan id", op)
	}

	// Валидация опциональных полей
	if req.Name != nil && *req.Name == "" {
		return nil, fmt.Errorf("%s: name cannot be empty", op)
	}
	if req.CPU != nil && *req.CPU <= 0 {
		return nil, fmt.Errorf("%s: cpu must be positive", op)
	}
	if req.RAMMB != nil && *req.RAMMB <= 0 {
		return nil, fmt.Errorf("%s: ram_mb must be positive", op)
	}
	if req.DiskGB != nil && *req.DiskGB <= 0 {
		return nil, fmt.Errorf("%s: disk_gb must be positive", op)
	}
	if req.PriceMonth != nil && *req.PriceMonth < 0 {
		return nil, fmt.Errorf("%s: price_month must be non-negative", op)
	}

	plan, err := s.planRepo.Update(ctx, req)
	if err != nil {
		if errors.Is(err, repository.ErrPlanNotFound) {
			log.Warn("plan not found for update")
			return nil, repository.ErrPlanNotFound
		}
		log.Error("failed to update plan", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("plan updated successfully")
	return plan, nil
}

// Delete удаляет план по ID
func (s *Service) Delete(ctx context.Context, id int32) error {
	const op = "service.plan.Delete"

	log := s.log.With(slog.String("op", op), slog.Int("id", int(id)))
	log.Info("deleting plan")

	if id <= 0 {
		return fmt.Errorf("%s: invalid plan id", op)
	}

	err := s.planRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrPlanNotFound) {
			log.Warn("plan not found for deletion")
			return repository.ErrPlanNotFound
		}
		log.Error("failed to delete plan", slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("plan deleted successfully")
	return nil
}

// List возвращает список планов
func (s *Service) List(ctx context.Context, activeOnly bool) ([]*models.Plan, error) {
	const op = "service.plan.List"

	log := s.log.With(slog.String("op", op), slog.Bool("activeOnly", activeOnly))
	log.Debug("listing plans")

	plans, err := s.planRepo.List(ctx, activeOnly)
	if err != nil {
		log.Error("failed to list plans", slog.String("error", err.Error()))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("plans listed successfully", slog.Int("count", len(plans)))
	return plans, nil
}
