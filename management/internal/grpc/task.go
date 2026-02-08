package grpc

import (
	"context"

	managementv1 "github.com/makhtech/proto/gen/go/management"
)

func (s *ServerAPI) CreateTask(ctx context.Context, req *managementv1.CreateTaskRequest) (*managementv1.Task, error) {
	panic("implement me")
}
func (s *ServerAPI) GetTask(ctx context.Context, req *managementv1.GetTaskRequest) (*managementv1.Task, error) {
	panic("implement me")
}
func (s *ServerAPI) ListTasksByVDS(ctx context.Context, req *managementv1.ListTasksByVDSRequest) (*managementv1.ListTasksResponse, error) {
	panic("implement me")
}
func (s *ServerAPI) UpdateTaskStatus(ctx context.Context, req *managementv1.UpdateTaskStatusRequest) (*managementv1.Task, error) {
	panic("implement me")
}
func (s *ServerAPI) GetPendingTasksCount(ctx context.Context, req *managementv1.GetPendingTasksCountRequest) (*managementv1.GetPendingTasksCountResponse, error) {
	panic("implement me")
}
