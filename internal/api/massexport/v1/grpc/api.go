package grpc

import (
	"github.com/ozontech/seq-ui/internal/pkg/service/massexport"
	massexport_v1 "github.com/ozontech/seq-ui/pkg/massexport/v1"
)

type API struct {
	massexport_v1.UnimplementedMassExportServiceServer

	exporter massexport.Service
}

func New(exporter massexport.Service) *API {
	return &API{
		exporter: exporter,
	}
}
