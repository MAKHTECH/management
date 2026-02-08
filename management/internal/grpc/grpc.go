package grpc

import (
	managementv1 "github.com/makhtech/proto/gen/go/management"
)

type ServerAPI struct {
	managementv1.UnimplementedManagementServer

	//auth auth.Auth todo: service
}
