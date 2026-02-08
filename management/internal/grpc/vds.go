package grpc

import (
	"context"

	managementv1 "github.com/makhtech/proto/gen/go/management"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ServerAPI) CreateVDS(ctx context.Context, req *managementv1.CreateVDSRequest) (*managementv1.VDS, error) {
	panic("implement me")
}
func (s *ServerAPI) GetVDS(ctx context.Context, req *managementv1.GetVDSRequest) (*managementv1.VDS, error) {
	panic("implement me")
}
func (s *ServerAPI) ListVDSByUser(ctx context.Context, req *managementv1.ListVDSByUserRequest) (*managementv1.ListVDSResponse, error) {
	panic("implement me")
}
func (s *ServerAPI) UpdateVDSStatus(ctx context.Context, req *managementv1.UpdateVDSStatusRequest) (*managementv1.VDS, error) {
	panic("implement me")
}
func (s *ServerAPI) AllocateIP(ctx context.Context, req *managementv1.AllocateIPRequest) (*managementv1.VDS, error) {
	panic("implement me")
}
func (s *ServerAPI) DeleteVDS(ctx context.Context, req *managementv1.DeleteVDSRequest) (*emptypb.Empty, error) {
	panic("implement me")
}
