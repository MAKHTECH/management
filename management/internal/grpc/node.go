package grpc

import (
	"context"

	managementv1 "github.com/makhtech/proto/gen/go/management"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ServerAPI) CreateNode(ctx context.Context, req *managementv1.CreateNodeRequest) (*managementv1.Node, error) {
	panic("implement me")
}
func (s *ServerAPI) GetNode(ctx context.Context, req *managementv1.GetNodeRequest) (*managementv1.Node, error) {
	panic("implement me")
}
func (s *ServerAPI) UpdateNode(ctx context.Context, req *managementv1.UpdateNodeRequest) (*managementv1.Node, error) {
	panic("implement me")
}
func (s *ServerAPI) ListNodes(ctx context.Context, req *managementv1.ListNodesRequest) (*managementv1.ListNodesResponse, error) {
	panic("implement me")
}
func (s *ServerAPI) DeleteNode(ctx context.Context, req *managementv1.GetNodeRequest) (*emptypb.Empty, error) {
	panic("implement me")
}
func (s *ServerAPI) GetNodeUtilization(ctx context.Context, req *managementv1.GetNodeRequest) (*managementv1.NodeUtilization, error) {
	panic("implement me")
}
