package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/metric"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
)

// serveExportAsyncSearch go doc.
//
//	@Router		/seqapi/v1/async_search/export [post]
//	@ID			seqapi_v1_export_async_search
//	@Tags		seqapi_v1
//	@Param		env		query		string						false	"Environment"
//	@Param		body	body		exportAsyncSearchRequest	true	"Request body"
//	@Success	200		{object}	exportResponse				"A successful streaming responses"
//	@Failure	default	{object}	httputil.Error				"An unexpected error response"
//	@Security	bearer
func (a *API) serveExportAsyncSearch(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewWriter(w)

	if a.asyncSearches == nil {
		wr.Error(types.ErrAsyncSearchesDisabled, http.StatusBadRequest)
		return
	}

	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_export_async_search_http")
	defer span.End()

	env := getEnvFromContext(ctx)

	userStr := "_"
	if userName, err := types.GetUserKey(ctx); err == nil {
		userStr = userName
	}

	var httpReq exportAsyncSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse export async search request: %w", err), http.StatusBadRequest)
		return
	}

	if httpReq.Format == "" {
		httpReq.Format = efJSONL
	}

	attributes := []attribute.KeyValue{
		{
			Key:   "search_id",
			Value: attribute.StringValue(httpReq.SearchID),
		},
		{
			Key:   "limit",
			Value: attribute.IntValue(int(httpReq.Limit)),
		},
		{
			Key:   "offset",
			Value: attribute.IntValue(int(httpReq.Offset)),
		},
		{
			Key:   "format",
			Value: attribute.StringValue(string(httpReq.Format)),
		},
		{
			Key:   "fields",
			Value: attribute.StringSliceValue(httpReq.Fields),
		},
	}

	if env != "" {
		attributes = append(attributes, attribute.String("env", env))
	}

	span.SetAttributes(attributes...)

	if err := checkUUID(httpReq.SearchID); err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}

	if err := checkLimitOffset(httpReq.Limit, httpReq.Offset); err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}

	params, err := a.GetEnvParams(env)
	if err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}

	if params.exportLimiter.Limited(userStr) {
		metric.ServerExportRequestLimits.Inc()
		wr.Error(errors.New("parallel export limit exceeded"), http.StatusTooManyRequests)
		return
	}
	defer params.exportLimiter.Fill(userStr)

	if httpReq.Limit > params.options.MaxExportLimit {
		wr.Error(fmt.Errorf("too many events are requested: count=%d, max=%d",
			httpReq.Limit, params.options.MaxExportLimit),
			http.StatusBadRequest)
		return
	}

	if httpReq.Format == efCSV && len(httpReq.Fields) == 0 {
		wr.Error(errors.New("csv export required 'fields'"), http.StatusBadRequest)
		return
	}

	cw, err := httputil.NewChunkedWriter(wr.ResponseWriter)
	if err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}

	if err := a.asyncSearches.ExportAsyncSearch(ctx, httpReq.toProto(), cw); err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}

	wr.WriteHeader(http.StatusOK)
}

type exportAsyncSearchRequest struct {
	SearchID string       `json:"search_id" format:"uuid"`
	Limit    int32        `json:"limit" format:"int32"`
	Offset   int32        `json:"offset" format:"int32"`
	Format   exportFormat `json:"format" default:"jsonl"`
	Fields   []string     `json:"fields,omitempty"`
} //	@name	seqapi.v1.ExportAsyncSearchRequest

func (r exportAsyncSearchRequest) toProto() *seqapi.ExportAsyncSearchRequest {
	return &seqapi.ExportAsyncSearchRequest{
		SearchId: r.SearchID,
		Limit:    r.Limit,
		Offset:   r.Offset,
		Format:   r.Format.toProto(),
		Fields:   r.Fields,
	}
}
