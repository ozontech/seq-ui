package grpc

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
)

type fieldsCache struct {
	cache cache.Cache
}

func newFieldsCache(c cache.Cache) *fieldsCache {
	return &fieldsCache{cache: c}
}

func (c *fieldsCache) get(ctx context.Context, key string) ([]*seqapi.Field, bool) {
	data, err := c.cache.Get(ctx, key)
	if err != nil {
		return nil, false
	}

	var resp seqapi.GetFieldsResponse
	if err := proto.Unmarshal([]byte(data), &resp); err != nil {
		logger.Error("can't unmarshal cached fields", zap.Error(err))
		return nil, false
	}

	return resp.GetFields(), true
}

func (c *fieldsCache) set(ctx context.Context, key string, fields []*seqapi.Field, ttl time.Duration) {
	data, err := proto.Marshal(&seqapi.GetFieldsResponse{Fields: fields})
	if err != nil {
		logger.Error("can't marshal fields for cache", zap.Error(err))
		return
	}

	if err := c.cache.SetWithTTL(ctx, key, string(data), ttl); err != nil {
		logger.Error("can't set fields to cache", zap.Error(err))
	}
}

func (a *API) GetFields(ctx context.Context, req *seqapi.GetFieldsRequest) (*seqapi.GetFieldsResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "seqapi_v1_get_fields")
	defer span.End()

	env := types.GetEnvKey(ctx)
	params, err := a.GetParams(env)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if params.fieldsCache == nil {
		resp, err := params.client.GetFields(ctx, req)
		if err != nil {
			return nil, err
		}

		resp.SystemFields = params.systemFields
		resp.PinnedFields = params.pinnedFields
		return resp, nil
	}

	if cachedFields, ok := params.fieldsCache.get(ctx, params.fieldsCacheKey); ok {
		return &seqapi.GetFieldsResponse{
			Fields:       cachedFields,
			SystemFields: params.systemFields,
			PinnedFields: params.pinnedFields,
		}, nil
	}

	resp, err := params.client.GetFields(ctx, req)
	if err != nil {
		return nil, err
	}

	params.fieldsCache.set(ctx, params.fieldsCacheKey, resp.GetFields(), params.options.Caches.Fields.TTL)
	resp.SystemFields = params.systemFields
	resp.PinnedFields = params.pinnedFields
	return resp, nil
}

func (a *API) GetPinnedFields(ctx context.Context, _ *seqapi.GetFieldsRequest) (*seqapi.GetFieldsResponse, error) {
	env := types.GetEnvKey(ctx)
	params, err := a.GetParams(env)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &seqapi.GetFieldsResponse{
		Fields: params.pinnedFields,
	}, nil
}
