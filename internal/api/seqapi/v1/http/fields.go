package http

import (
	"encoding/json"
	"net/http"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.uber.org/zap"
)

// serveGetFields go doc.
//
//	@Router		/seqapi/v1/fields [get]
//	@ID			seqapi_v1_getFields
//	@Tags		seqapi_v1
//	@Success	200		{object}	getFieldsResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error		"An unexpected error response"
func (a *API) serveGetFields(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_get_fields")
	defer span.End()

	wr := httputil.NewWriter(w)

	if a.fieldsCache == nil {
		resp, err := a.seqDB.GetFields(ctx, &seqapi.GetFieldsRequest{})
		if err != nil {
			wr.Error(err, http.StatusInternalServerError)
			return
		}

		wr.WriteJson(getFieldsResponseFromProto(resp))
		return
	}

	rawFields, cached, isActual := a.fieldsCache.getFields()
	if cached && isActual {
		_, _ = wr.Write(rawFields)
		return
	}

	resp, err := a.seqDB.GetFields(ctx, &seqapi.GetFieldsRequest{})
	if err != nil {
		if cached {
			logger.Error("can't get fields; use cached fields", zap.Error(err))
			_, _ = wr.Write(rawFields)
			return
		}

		wr.Error(err, http.StatusInternalServerError)
		return
	}

	res := getFieldsResponseFromProto(resp)
	resData, err := json.Marshal(res)
	if err != nil {
		if cached {
			logger.Error("can't marshal fields; use cached fields", zap.Error(err))
			_, _ = wr.Write(rawFields)
			return
		}

		wr.Error(err, http.StatusInternalServerError)
		return
	}

	a.fieldsCache.setFields(resData)
	_, _ = wr.Write(resData)
}

// serveGetPinnedFields go doc.
//
//	@Router		/seqapi/v1/fields/pinned [get]
//	@ID			seqapi_v1_getPinnedFields
//	@Tags		seqapi_v1
//	@Success	200		{object}	getFieldsResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error		"An unexpected error response"
func (a *API) serveGetPinnedFields(w http.ResponseWriter, _ *http.Request) {
	httputil.NewWriter(w).WriteJson(getFieldsResponse{
		Fields: a.pinnedFields,
	})
}

type field struct {
	Name string `json:"name"`
	Type string `json:"type" default:"unknown" enums:"unknown,keyword,text"`
} // @name seqapi.v1.Field

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
	Fields fields `json:"fields"`
} // @name seqapi.v1.GetFieldsResponse

func getFieldsResponseFromProto(proto *seqapi.GetFieldsResponse) getFieldsResponse {
	return getFieldsResponse{
		Fields: fieldsFromProto(proto.GetFields()),
	}
}
