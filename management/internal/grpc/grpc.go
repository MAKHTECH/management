package grpc

import (
	"github.com/makhtech/management/internal/service"
	managementv1 "github.com/makhtech/proto/gen/go/management"
)

type ServerAPI struct {
	managementv1.UnimplementedManagementServer

	planService service.PlanService
}

// NewServerAPI создает новый ServerAPI с зависимостями
func NewServerAPI(planSvc service.PlanService) *ServerAPI {
	return &ServerAPI{
		planService: planSvc,
	}
}
