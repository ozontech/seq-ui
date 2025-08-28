package massexport_v1

import (
	"github.com/go-chi/chi/v5"
	grpc_api "github.com/ozontech/seq-ui/internal/api/massexport/v1/grpc"
	http_api "github.com/ozontech/seq-ui/internal/api/massexport/v1/http"
	"github.com/ozontech/seq-ui/internal/pkg/service/massexport"
)

type MassExport struct {
	grpcAPI *grpc_api.API
	httpAPI *http_api.API
}

func New(service massexport.Service) *MassExport {
	return &MassExport{
		grpcAPI: grpc_api.New(service),
		httpAPI: http_api.New(service),
	}
}

func (m *MassExport) GRPCServer() *grpc_api.API {
	return m.grpcAPI
}

func (m *MassExport) HTTPRouter() chi.Router {
	return m.httpAPI.Router()
}
