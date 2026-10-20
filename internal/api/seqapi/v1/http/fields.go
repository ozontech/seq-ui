package http

import (
	"net/http"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
)

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
	params := a.paramsByEnv[types.GetEnvKey(ctx)]

	if cached, ok := params.caches.Fields.Get(ctx, ""); ok {
		var resp seqapi.GetFieldsResponse
		err := proto.Unmarshal([]byte(cached), &resp)
		if err == nil {
			wr.WriteJson(getFieldsResponse{
				Fields:       fieldsFromProto(resp.GetFields()),
				SystemFields: params.systemFields,
				PinnedFields: params.pinnedFields,
			})
			return
		}
		logger.Error("failed to unmarshal cached fields", zap.Error(err))
	}

	resp, err := params.client.GetFields(ctx, &seqapi.GetFieldsRequest{})
	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}

	if params.caches.Fields != nil {
		if data, err := proto.Marshal(&seqapi.GetFieldsResponse{Fields: resp.GetFields()}); err == nil {
			params.caches.Fields.Set(ctx, "", string(data))
		} else {
			logger.Error("failed to marshal fields proto for caching", zap.Error(err))
		}
	}

	wr.WriteJson(getFieldsResponse{
		Fields:       fieldsFromProto(resp.GetFields()),
		SystemFields: params.systemFields,
		PinnedFields: params.pinnedFields,
	})
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
	params := a.paramsByEnv[types.GetEnvKey(ctx)]

	wr.WriteJson(getFieldsResponse{
		Fields: params.pinnedFields,
	})
}

type field struct {
	Name string `json:"name"`
	Type string `json:"type" default:"unknown" enums:"unknown,keyword,text"`
} //	@name	seqapi.v1.Field

func fieldFromProto(f *seqapi.Field) field {
	return field{
		Name: f.GetName(),
		Type: seqdb.FieldTypeFromProto(f.GetType()),
	}
}

type fields []field

func fieldsFromProto(fs []*seqapi.Field) fields {
	res := make(fields, len(fs))
	for i, f := range fs {
		res[i] = fieldFromProto(f)
	}
	return res
}

type getFieldsResponse struct {
	Fields       fields `json:"fields"`
	SystemFields fields `json:"system_fields,omitempty"`
	PinnedFields fields `json:"pinned_fields,omitempty"`
} //	@name	seqapi.v1.GetFieldsResponse
