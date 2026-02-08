package grpc

import (
	"context"

	managementv1 "github.com/makhtech/proto/gen/go/management"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ServerAPI) CreatePlan(ctx context.Context, req *managementv1.CreatePlanRequest) (*managementv1.Plan, error) {
	panic("implement me")

}
func (s *ServerAPI) GetPlan(ctx context.Context, req *managementv1.GetPlanRequest) (*managementv1.Plan, error) {
	panic("implement me")

}
func (s *ServerAPI) UpdatePlan(ctx context.Context, req *managementv1.UpdatePlanRequest) (*managementv1.Plan, error) {
	panic("implement me")

}
func (s *ServerAPI) ListPlans(ctx context.Context, req *managementv1.ListPlansRequest) (*managementv1.ListPlansResponse, error) {
	panic("implement me")

}
func (s *ServerAPI) DeletePlan(ctx context.Context, req *managementv1.GetPlanRequest) (*emptypb.Empty, error) {
	panic("implement me")
}
