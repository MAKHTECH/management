package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/makhtech/management/internal/domain/models"
	"github.com/makhtech/management/internal/repository"
)

// PlanRepository - репозиторий для работы с планами
type PlanRepository struct {
	db *Database
}

// NewPlanRepository создает новый репозиторий планов
func NewPlanRepository(db *Database) *PlanRepository {
	return &PlanRepository{db: db}
}

// Create создает новый план
func (r *PlanRepository) Create(ctx context.Context, req *models.CreatePlanRequest) (*models.Plan, error) {
	const op = "repository.postgres.PlanRepository.Create"

	query := `
		INSERT INTO plans (name, cpu, ram_mb, disk_gb, price_month, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, true, $6)
		RETURNING id, name, cpu, ram_mb, disk_gb, price_month, is_active, created_at
	`

	var plan models.Plan
	now := time.Now()

	err := r.db.Pool.QueryRow(ctx, query,
		req.Name,
		req.CPU,
		req.RAMMB,
		req.DiskGB,
		req.PriceMonth,
		now,
	).Scan(
		&plan.ID,
		&plan.Name,
		&plan.CPU,
		&plan.RAMMB,
		&plan.DiskGB,
		&plan.PriceMonth,
		&plan.IsActive,
		&plan.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &plan, nil
}

// GetByID получает план по ID
func (r *PlanRepository) GetByID(ctx context.Context, id int32) (*models.Plan, error) {
	const op = "repository.postgres.PlanRepository.GetByID"

	query := `
		SELECT id, name, cpu, ram_mb, disk_gb, price_month, is_active, created_at
		FROM plans
		WHERE id = $1
	`

	var plan models.Plan
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&plan.ID,
		&plan.Name,
		&plan.CPU,
		&plan.RAMMB,
		&plan.DiskGB,
		&plan.PriceMonth,
		&plan.IsActive,
		&plan.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrPlanNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &plan, nil
}

// Update обновляет существующий план
func (r *PlanRepository) Update(ctx context.Context, req *models.UpdatePlanRequest) (*models.Plan, error) {
	const op = "repository.postgres.PlanRepository.Update"

	// Строим динамический запрос
	var setClauses []string
	var args []interface{}
	argIndex := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}
	if req.CPU != nil {
		setClauses = append(setClauses, fmt.Sprintf("cpu = $%d", argIndex))
		args = append(args, *req.CPU)
		argIndex++
	}
	if req.RAMMB != nil {
		setClauses = append(setClauses, fmt.Sprintf("ram_mb = $%d", argIndex))
		args = append(args, *req.RAMMB)
		argIndex++
	}
	if req.DiskGB != nil {
		setClauses = append(setClauses, fmt.Sprintf("disk_gb = $%d", argIndex))
		args = append(args, *req.DiskGB)
		argIndex++
	}
	if req.PriceMonth != nil {
		setClauses = append(setClauses, fmt.Sprintf("price_month = $%d", argIndex))
		args = append(args, *req.PriceMonth)
		argIndex++
	}
	if req.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, req.ID)
	}

	args = append(args, req.ID)

	query := fmt.Sprintf(`
		UPDATE plans
		SET %s
		WHERE id = $%d
		RETURNING id, name, cpu, ram_mb, disk_gb, price_month, is_active, created_at
	`, strings.Join(setClauses, ", "), argIndex)

	var plan models.Plan
	err := r.db.Pool.QueryRow(ctx, query, args...).Scan(
		&plan.ID,
		&plan.Name,
		&plan.CPU,
		&plan.RAMMB,
		&plan.DiskGB,
		&plan.PriceMonth,
		&plan.IsActive,
		&plan.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrPlanNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &plan, nil
}

// Delete удаляет план по ID
func (r *PlanRepository) Delete(ctx context.Context, id int32) error {
	const op = "repository.postgres.PlanRepository.Delete"

	query := `DELETE FROM plans WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrPlanNotFound
	}

	return nil
}

// List возвращает список планов
func (r *PlanRepository) List(ctx context.Context, activeOnly bool) ([]*models.Plan, error) {
	const op = "repository.postgres.PlanRepository.List"

	var query string
	var args []interface{}

	if activeOnly {
		query = `
			SELECT id, name, cpu, ram_mb, disk_gb, price_month, is_active, created_at
			FROM plans
			WHERE is_active = true
			ORDER BY id
		`
	} else {
		query = `
			SELECT id, name, cpu, ram_mb, disk_gb, price_month, is_active, created_at
			FROM plans
			ORDER BY id
		`
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var plans []*models.Plan
	for rows.Next() {
		var plan models.Plan
		if err := rows.Scan(
			&plan.ID,
			&plan.Name,
			&plan.CPU,
			&plan.RAMMB,
			&plan.DiskGB,
			&plan.PriceMonth,
			&plan.IsActive,
			&plan.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		plans = append(plans, &plan)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return plans, nil
}
