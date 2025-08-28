package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) Status(ctx context.Context, req *seqapi.StatusRequest) (*seqapi.StatusResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_status")
	defer span.End()

	return a.seqDB.Status(ctx, req)
}
