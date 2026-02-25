package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/metric"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// serveExport go doc.
//
//	@Router		/seqapi/v1/export [post]
//	@ID			seqapi_v1_export
//	@Tags		seqapi_v1
//	@Param		env		query		string			false	"Environment"
//	@Param		body	body		exportRequest	true	"Request body"
//	@Success	200		{object}	exportResponse	"A successful streaming responses"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveExport(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_export_http")
	defer span.End()

	wr := httputil.NewWriter(w)

	env := getEnvFromContext(ctx)
	params, err := a.GetEnvParams(env)
	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}

	userStr := "_"
	if userName, err := types.GetUserKey(ctx); err == nil {
		userStr = userName
	}

	var httpReq exportRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse export request: %w", err), http.StatusBadRequest)
		return
	}

	if httpReq.Format == "" {
		httpReq.Format = efJSONL
	}

	attributes := []attribute.KeyValue{
		attribute.KeyValue{
			Key:   "query",
			Value: attribute.StringValue(httpReq.Query),
		},
		attribute.KeyValue{
			Key:   "from",
			Value: attribute.StringValue(httpReq.From.Format(time.DateTime)),
		},
		attribute.KeyValue{
			Key:   "to",
			Value: attribute.StringValue(httpReq.To.Format(time.DateTime)),
		},
		attribute.KeyValue{
			Key:   "limit",
			Value: attribute.IntValue(int(httpReq.Limit)),
		},
		attribute.KeyValue{
			Key:   "offset",
			Value: attribute.IntValue(int(httpReq.Offset)),
		},
		attribute.KeyValue{
			Key:   "format",
			Value: attribute.StringValue(string(httpReq.Format)),
		},
		attribute.KeyValue{
			Key:   "fields",
			Value: attribute.StringSliceValue(httpReq.Fields),
		},
	}

	if env != "" {
		attributes = append(attributes, attribute.String("env", env))
	}

	span.SetAttributes(attributes...)

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

	err = params.client.Export(ctx, httpReq.toProto(), cw)
	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}

	wr.WriteHeader(http.StatusOK)
}

type exportFormat string //	@name	seqapi.v1.ExportFormat

const (
	efJSONL exportFormat = "jsonl"
	efCSV   exportFormat = "csv"
)

func (f exportFormat) toProto() seqapi.ExportFormat {
	switch f {
	case efCSV:
		return seqapi.ExportFormat_EXPORT_FORMAT_CSV
	default:
		return seqapi.ExportFormat_EXPORT_FORMAT_JSONL
	}
}

type exportRequest struct {
	Query  string       `json:"query"`
	From   time.Time    `json:"from" format:"date-time"`
	To     time.Time    `json:"to" format:"date-time"`
	Limit  int32        `json:"limit" format:"int32"`
	Offset int32        `json:"offset" format:"int32"`
	Format exportFormat `json:"format" default:"jsonl"`
	Fields []string     `json:"fields,omitempty"`
} //	@name	seqapi.v1.ExportRequest

func (r exportRequest) toProto() *seqapi.ExportRequest {
	return &seqapi.ExportRequest{
		Query:  r.Query,
		From:   timestamppb.New(r.From),
		To:     timestamppb.New(r.To),
		Limit:  r.Limit,
		Offset: r.Offset,
		Format: r.Format.toProto(),
		Fields: r.Fields,
	}
}

// nolint:unused
//
//	@Description	Export response in one of the following formats:<br>
//	@Description	- JSONL: {"id":"some-id","data":{"field1":"value1","field2":"value2"},"time":"2024-12-31T10:20:30.0004Z"}<br>
//	@Description	- CSV: value1,value2,value3
type exportResponse struct {
} //	@name	seqapi.v1.ExportResponse
