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

func (a *API) FetchAsyncSearchResult(
	ctx context.Context,
	req *seqapi.FetchAsyncSearchResultRequest,
) (*seqapi.FetchAsyncSearchResultResponse, error) {
	if a.asyncSearches == nil {
		return nil, status.Error(codes.Unimplemented, types.ErrAsyncSearchesDisabled.Error())
	}

	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_fetch_async_search_result")
	defer span.End()

	spanAttributes := []attribute.KeyValue{
		{
			Key:   "search_id",
			Value: attribute.StringValue(req.SearchId),
		},
		{
			Key:   "offset",
			Value: attribute.IntValue(int(req.Offset)),
		},
		{
			Key:   "limit",
			Value: attribute.IntValue(int(req.Limit)),
		},
		{
			Key:   "order",
			Value: attribute.StringValue(string(req.Order)),
		},
	}
	span.SetAttributes(spanAttributes...)

	if _, err := uuid.FromString(req.SearchId); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid search_id")
	}

	resp, err := a.asyncSearches.FetchAsyncSearchResult(ctx, req)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return resp, nil
}
