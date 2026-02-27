package grpc

import (
	"context"

	"github.com/gofrs/uuid"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
)

func (a *API) CancelAsyncSearch(
	ctx context.Context,
	req *seqapi.CancelAsyncSearchRequest,
) (*seqapi.CancelAsyncSearchResponse, error) {
	if a.asyncSearches == nil {
		return nil, status.Error(codes.Unimplemented, types.ErrAsyncSearchesDisabled.Error())
	}

	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_cancel_async_search")
	defer span.End()

	if _, err := uuid.FromString(req.SearchId); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid search_id")
	}

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "search_id",
			Value: attribute.StringValue(req.SearchId),
		},
	)

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	resp, err := a.asyncSearches.CancelAsyncSearch(ctx, profileID, req)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return resp, nil
}
