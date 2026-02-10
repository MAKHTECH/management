package grpc

import (
	"context"
	"errors"
	"log/slog"

	"github.com/makhtech/management/internal/domain/models"
	"github.com/makhtech/management/internal/repository"
	managementv1 "github.com/makhtech/proto/gen/go/management"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// for admins:
func (s *ServerAPI) CreatePlan(ctx context.Context, req *managementv1.CreatePlanRequest) (*managementv1.Plan, error) {
	domainReq := &models.CreatePlanRequest{
		Name:       req.GetName(),
		CPU:        req.GetCpu(),
		RAMMB:      req.GetRamMb(),
		DiskGB:     req.GetDiskGb(),
		PriceMonth: req.GetPriceMonth(),
	}

	plan, err := s.planService.Create(ctx, domainReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create plan: %v", err)
	}

	return planToProto(plan), nil
}

func (s *ServerAPI) UpdatePlan(ctx context.Context, req *managementv1.UpdatePlanRequest) (*managementv1.Plan, error) {
	domainReq := &models.UpdatePlanRequest{
		ID: req.GetId(),
	}

	if req.Name != nil {
		domainReq.Name = req.Name
	}
	if req.Cpu != nil {
		domainReq.CPU = req.Cpu
	}
	if req.RamMb != nil {
		domainReq.RAMMB = req.RamMb
	}
	if req.DiskGb != nil {
		domainReq.DiskGB = req.DiskGb
	}
	if req.PriceMonth != nil {
		domainReq.PriceMonth = req.PriceMonth
	}
	if req.IsActive != nil {
		domainReq.IsActive = req.IsActive
	}

	plan, err := s.planService.Update(ctx, domainReq)
	if err != nil {
		if errors.Is(err, repository.ErrPlanNotFound) {
			return nil, status.Errorf(codes.NotFound, "plan not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update plan: %v", err)
	}

	return planToProto(plan), nil
}

func (s *ServerAPI) DeletePlan(ctx context.Context, req *managementv1.GetPlanRequest) (*emptypb.Empty, error) {
	err := s.planService.Delete(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, repository.ErrPlanNotFound) {
			return nil, status.Errorf(codes.NotFound, "plan not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete plan: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// for users:
func (s *ServerAPI) GetPlan(ctx context.Context, req *managementv1.GetPlanRequest) (*managementv1.Plan, error) {
	plan, err := s.planService.GetByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, repository.ErrPlanNotFound) {
			return nil, status.Errorf(codes.NotFound, "plan not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get plan: %v", err)
	}

	return planToProto(plan), nil
}

func (s *ServerAPI) ListPlans(ctx context.Context, req *managementv1.ListPlansRequest) (*managementv1.ListPlansResponse, error) {
	slog.Info("ListPlans called", slog.Bool("active_only", req.GetActiveOnly()))

	plans, err := s.planService.List(ctx, req.GetActiveOnly())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list plans: %v", err)
	}

	protoPlans := make([]*managementv1.Plan, 0, len(plans))
	for _, plan := range plans {
		protoPlans = append(protoPlans, planToProto(plan))
	}

	return &managementv1.ListPlansResponse{
		Plans: protoPlans,
	}, nil
}

// planToProto конвертирует domain модель в proto
func planToProto(plan *models.Plan) *managementv1.Plan {
	return &managementv1.Plan{
		Id:         plan.ID,
		Name:       plan.Name,
		Cpu:        plan.CPU,
		RamMb:      plan.RAMMB,
		DiskGb:     plan.DiskGB,
		PriceMonth: plan.PriceMonth,
		IsActive:   plan.IsActive,
		CreatedAt:  timestamppb.New(plan.CreatedAt),
	}
}
