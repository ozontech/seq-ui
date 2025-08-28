package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/pkg/massexport/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) GetAll(ctx context.Context, req *massexport.GetAllRequest) (*massexport.GetAllResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "massexport_v1_jobs")
	defer span.End()

	exports, err := a.exporter.GetAll(ctx)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	result := make([]*massexport.CheckResponse, 0, len(exports))
	for i := range exports {
		result = append(result, convertExportInfo(exports[i]))
	}

	return &massexport.GetAllResponse{Exports: result}, nil
}
