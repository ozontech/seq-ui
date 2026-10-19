package grpc

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) GetAsyncSearchesList(ctx context.Context, req *seqapi.GetAsyncSearchesListRequest) (*seqapi.GetAsyncSearchesListResponse, error) {
	if a.asyncSearchesByEnv == nil {
		return nil, status.Error(codes.Unimplemented, types.ErrAsyncSearchesDisabled.Error())
	}

	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_get_async_searches_list")
	defer span.End()

	env := types.GetEnvKey(ctx)
	svc := a.asyncSearchesByEnv[env]

	attributes := []attribute.KeyValue{
		{
			Key:   "offset",
			Value: attribute.IntValue(int(req.Offset)),
		},
		{
			Key:   "limit",
			Value: attribute.IntValue(int(req.Limit)),
		},
	}
	if req.Status != nil {
		attributes = append(attributes, attribute.KeyValue{
			Key:   "status",
			Value: attribute.StringValue(string(*req.Status)),
		})
	}
	if req.OwnerName != nil {
		attributes = append(attributes, attribute.KeyValue{
			Key:   "owner",
			Value: attribute.StringValue(*req.OwnerName),
		})
	}
	if env != "" {
		attributes = append(attributes, attribute.String("env", env))
	}
	span.SetAttributes(attributes...)

	if err := checkLimitOffset(req.Limit, req.Offset); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := svc.GetAsyncSearchesList(ctx, req)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return resp, nil
}
