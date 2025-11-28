package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/api_error"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// serveSearch go doc.
//
//	@Router		/seqapi/v1/search [post]
//	@ID			seqapi_v1_search
//	@Tags		seqapi_v1
//	@Param		body	body		searchRequest	true	"Request body"
//	@Success	200		{object}	searchResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveSearch(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_search")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq searchRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse search request: %w", err), http.StatusBadRequest)
		return
	}

	if httpReq.Order == "" {
		httpReq.Order = oDESC
	}

	spanAttributes := []attribute.KeyValue{
		{
			Key:   "query",
			Value: attribute.StringValue(httpReq.Query),
		},
		{
			Key:   "from",
			Value: attribute.StringValue(httpReq.From.Format(time.DateTime)),
		},
		{
			Key:   "to",
			Value: attribute.StringValue(httpReq.To.Format(time.DateTime)),
		},
		{
			Key:   "with_total",
			Value: attribute.BoolValue(httpReq.WithTotal),
		},
		{
			Key:   "offset",
			Value: attribute.IntValue(int(httpReq.Offset)),
		},
		{
			Key:   "limit",
			Value: attribute.IntValue(int(httpReq.Limit)),
		},
		{
			Key:   "order",
			Value: attribute.StringValue(string(httpReq.Order)),
		},
	}
	if httpReq.Histogram.Interval != "" {
		spanAttributes = append(spanAttributes, attribute.KeyValue{
			Key:   "histogram_interval",
			Value: attribute.StringValue(httpReq.Histogram.Interval),
		})
	}
	if len(httpReq.Aggregations) > 0 {
		aggsRaw, _ := json.Marshal(httpReq.Aggregations)
		spanAttributes = append(spanAttributes, attribute.KeyValue{
			Key:   "aggregations",
			Value: attribute.StringValue(string(aggsRaw)),
		})
	}

	span.SetAttributes(spanAttributes...)

	if err := api_error.CheckSearchLimit(httpReq.Limit, a.config.MaxSearchLimit); err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}
	if err := api_error.CheckAggregationsCount(len(httpReq.Aggregations), a.config.MaxAggregationsPerRequest); err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}
	if err := api_error.CheckSearchOffsetLimit(httpReq.Offset, a.config.MaxSearchOffsetLimit); err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}

	resp, err := a.seqDB.Search(ctx, httpReq.toProto())
	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}

	if resp.Total > a.config.MaxSearchTotalLimit {
		resp.Error = &seqapi.Error{
			Code:    seqapi.ErrorCode_ERROR_CODE_QUERY_TOO_HEAVY,
			Message: api_error.ErrQueryTooHeavy.Error(),
		}
	}

	searchResp := searchResponseFromProto(resp, httpReq.WithTotal)
	if a.masker != nil {
		for i := range searchResp.Events {
			a.masker.Mask(searchResp.Events[i].Data)
		}
	}

	wr.WriteJson(searchResp)
}

type order string // @name seqapi.v1.Order

const (
	oDESC order = "desc"
	oASC  order = "asc"
)

func (o order) toProto() seqapi.Order {
	switch o {
	case oASC:
		return seqapi.Order_ORDER_ASC
	default:
		return seqapi.Order_ORDER_DESC
	}
}

type searchRequest struct {
	Query        string             `json:"query"`
	From         time.Time          `json:"from" format:"date-time"`
	To           time.Time          `json:"to" format:"date-time"`
	Limit        int32              `json:"limit" format:"int32"`
	Offset       int32              `json:"offset" format:"int32"`
	WithTotal    bool               `json:"withTotal"`
	Aggregations aggregationQueries `json:"aggregations"`
	Histogram    struct {
		Interval string `json:"interval"`
	} `json:"histogram"`
	Order order `json:"order" default:"desc"`
} // @name seqapi.v1.SearchRequest

func (r searchRequest) toProto() *seqapi.SearchRequest {
	req := &seqapi.SearchRequest{
		Query:        r.Query,
		From:         timestamppb.New(r.From),
		To:           timestamppb.New(r.To),
		Limit:        r.Limit,
		Offset:       r.Offset,
		WithTotal:    r.WithTotal,
		Aggregations: r.Aggregations.toProto(),
		Order:        r.Order.toProto(),
	}
	if r.Histogram.Interval != "" {
		req.Histogram = &seqapi.SearchRequest_Histogram{
			Interval: r.Histogram.Interval,
		}
	}
	return req
}

type searchResponse struct {
	Events          events       `json:"events"`
	Histogram       *histogram   `json:"histogram,omitempty"`
	Aggregations    aggregations `json:"aggregations,omitempty"`
	Total           string       `json:"total,omitempty" format:"int64"`
	Error           apiError     `json:"error"`
	PartialResponse bool         `json:"partialResponse"`
} // @name seqapi.v1.SearchResponse

func searchResponseFromProto(proto *seqapi.SearchResponse, withTotal bool) searchResponse {
	sr := searchResponse{
		Events:          eventsFromProto(proto.GetEvents()),
		Histogram:       histogramFromProto(proto.GetHistogram(), false),
		Aggregations:    aggregationsFromProto(proto.GetAggregations(), false),
		Error:           apiErrorFromProto(proto.GetError()),
		PartialResponse: proto.GetPartialResponse(),
	}
	if withTotal {
		sr.Total = strconv.FormatInt(proto.GetTotal(), 10)
	}
	return sr
}
