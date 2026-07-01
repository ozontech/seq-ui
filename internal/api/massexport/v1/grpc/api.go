package grpc

import (
	"github.com/ozontech/seq-ui/internal/pkg/service/massexport"
	api "github.com/ozontech/seq-ui/pkg/massexport/v1"
)

type API struct {
	api.UnimplementedMassExportServiceServer

	exporter massexport.Service
}

func New(exporter massexport.Service) *API {
	return &API{
		exporter: exporter,
	}
}
