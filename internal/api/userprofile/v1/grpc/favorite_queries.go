package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/internal/api/grpcutil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/userprofile/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// GetFavoriteQueries returns user's favorite queries.
func (a *API) GetFavoriteQueries(ctx context.Context, _ *userprofile.GetFavoriteQueriesRequest) (*userprofile.GetFavoriteQueriesResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "userprofile_v1_get_favorite_queries")
	defer span.End()

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	request := types.GetFavoriteQueriesRequest{
		ProfileID: profileID,
	}
	favoriteQueries, err := a.service.GetFavoriteQueries(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &userprofile.GetFavoriteQueriesResponse{
		Queries: favoriteQueries.ToProto(),
	}, nil
}

// CreateFavoriteQuery creates user's favorite query if it doesn't exist.
func (a *API) CreateFavoriteQuery(ctx context.Context, req *userprofile.CreateFavoriteQueryRequest) (*userprofile.CreateFavoriteQueryResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "userprofile_v1_create_favorite_query")
	defer span.End()

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "query",
			Value: attribute.StringValue(req.GetQuery()),
		},
		attribute.KeyValue{
			Key:   "name",
			Value: attribute.StringValue(req.GetName()),
		},
		attribute.KeyValue{
			Key:   "relative_from",
			Value: attribute.Int64Value(int64(req.GetRelativeFrom())),
		},
	)

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	request := types.GetOrCreateFavoriteQueryRequest{
		ProfileID: profileID,
		Query:     req.Query,
	}
	if req.RelativeFrom != nil {
		request.RelativeFrom = *req.RelativeFrom
	}
	if req.Name != nil {
		request.Name = *req.Name
	}
	fqID, err := a.service.GetOrCreateFavoriteQuery(ctx, request)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &userprofile.CreateFavoriteQueryResponse{
		Id: fqID,
	}, nil
}

// DeleteFavoriteQuery deletes user's favorite query.
func (a *API) DeleteFavoriteQuery(ctx context.Context, req *userprofile.DeleteFavoriteQueryRequest) (*userprofile.DeleteFavoriteQueryResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "userprofile_v1_delete_favorite_query")
	defer span.End()

	span.SetAttributes(attribute.KeyValue{
		Key:   "id",
		Value: attribute.Int64Value(req.GetId()),
	})

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	request := types.DeleteFavoriteQueryRequest{
		ID:        req.Id,
		ProfileID: profileID,
	}
	if err = a.service.DeleteFavoriteQuery(ctx, request); err != nil {
		return nil, grpcutil.ProcessError(err)
	}

	return &userprofile.DeleteFavoriteQueryResponse{}, nil
}
