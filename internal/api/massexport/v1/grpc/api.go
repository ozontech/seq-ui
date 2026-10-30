package grpc

import (
	"github.com/ozontech/seq-ui/internal/pkg/service/massexport"
	api "github.com/ozontech/seq-ui/pkg/massexport/v1"
)

type API struct {
	api.UnimplementedMassExportServiceServer

	serviceByEnv map[string]massexport.Service
}

func New(serviceByEnv map[string]massexport.Service) *API {
	return &API{
		serviceByEnv: serviceByEnv,
	}
}
