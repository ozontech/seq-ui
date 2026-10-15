package http

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
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

// serveGetFields go doc.
//
//	@Router		/seqapi/v1/fields [get]
//	@ID			seqapi_v1_getFields
//	@Tags		seqapi_v1
//	@Param		env		query		string				false	"Environment"
//	@Success	200		{object}	getFieldsResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error		"An unexpected error response"
func (a *API) serveGetFields(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_get_fields")
	defer span.End()

	wr := httputil.NewWriter(w)

	env := types.GetEnvKey(ctx)
	params, err := a.GetEnvParams(env)
	if err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}

	if params.fieldsCache == nil {
		resp, err := params.client.GetFields(ctx, &seqapi.GetFieldsRequest{})
		if err != nil {
			wr.Error(err, http.StatusInternalServerError)
			return
		}

		res := getFieldsResponse{
			Fields:       fieldsFromProto(resp.GetFields()),
			SystemFields: params.systemFields,
			PinnedFields: params.pinnedFields,
		}
		wr.WriteJson(res)
		return
	}

	if cachedFields, ok := params.fieldsCache.get(ctx, params.fieldsCacheKey); ok {
		res := getFieldsResponse{
			Fields:       fieldsFromProto(cachedFields),
			SystemFields: params.systemFields,
			PinnedFields: params.pinnedFields,
		}
		wr.WriteJson(res)
		return
	}

	resp, err := params.client.GetFields(ctx, &seqapi.GetFieldsRequest{})
	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}

	params.fieldsCache.set(ctx, params.fieldsCacheKey, resp.GetFields(), params.options.Caches.Fields.TTL)
	res := getFieldsResponse{
		Fields:       fieldsFromProto(resp.GetFields()),
		SystemFields: params.systemFields,
		PinnedFields: params.pinnedFields,
	}

	wr.WriteJson(res)
}

// serveGetPinnedFields go doc.
//
//	@Router		/seqapi/v1/fields/pinned [get]
//	@ID			seqapi_v1_get_pinned_fields
//	@Tags		seqapi_v1
//	@Param		env		query		string				false	"Environment"
//	@Success	200		{object}	getFieldsResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error		"An unexpected error response"
func (a *API) serveGetPinnedFields(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_get_pinned_fields")
	defer span.End()

	wr := httputil.NewWriter(w)

	env := types.GetEnvKey(ctx)
	params, err := a.GetEnvParams(env)
	if err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}

	httputil.NewWriter(w).WriteJson(getFieldsResponse{
		Fields: params.pinnedFields,
	})
}

type field struct {
	Name string `json:"name"`
	Type string `json:"type" default:"unknown" enums:"unknown,keyword,text"`
} //	@name	seqapi.v1.Field

func fieldFromProto(proto *seqapi.Field) field {
	return field{
		Name: proto.GetName(),
		Type: seqdb.FieldTypeFromProto(proto.GetType()),
	}
}

type fields []field

func fieldsFromProto(proto []*seqapi.Field) fields {
	res := make(fields, len(proto))
	for i, f := range proto {
		res[i] = fieldFromProto(f)
	}
	return res
}

type getFieldsResponse struct {
	Fields       fields `json:"fields"`
	SystemFields fields `json:"system_fields,omitempty"`
	PinnedFields fields `json:"pinned_fields,omitempty"`
} //	@name	seqapi.v1.GetFieldsResponse
