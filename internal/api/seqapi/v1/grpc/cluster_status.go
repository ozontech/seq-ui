package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) Status(ctx context.Context, req *seqapi.StatusRequest) (*seqapi.StatusResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_status")
	defer span.End()

	env := a.GetEnvFromContext(ctx)
	params, err := a.GetParams(env)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return params.client.Status(ctx, req)
}
