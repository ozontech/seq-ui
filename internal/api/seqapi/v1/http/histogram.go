package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// serveGetHistogram go doc.
//
//	@Router		/seqapi/v1/histogram [post]
//	@ID			seqapi_v1_getHistogram
//	@Tags		seqapi_v1
//	@Param		env		query		string					false	"Environment"
//	@Param		body	body		getHistogramRequest		true	"Request body"
//	@Success	200		{object}	getHistogramResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error			"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetHistogram(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_get_histogram")
	defer span.End()

	wr := httputil.NewWriter(w)

	env := getEnvFromContext(ctx)

	var httpReq getHistogramRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse getHistogram request: %w", err), http.StatusBadRequest)
		return
	}

	attributes := []attribute.KeyValue{
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
			Key:   "interval",
			Value: attribute.StringValue(httpReq.Interval),
		},
	}

	if env != "" {
		attributes = append(attributes, attribute.String("env", env))
	}

	span.SetAttributes(attributes...)

	params, err := a.GetEnvParams(env)
	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}

	resp, err := params.client.GetHistogram(ctx, httpReq.toProto())
	if err != nil {
		wr.Error(err, http.StatusInternalServerError)
		return
	}

	wr.WriteJson(getHistogramResponseFromProto(resp))
}

type getHistogramRequest struct {
	Query    string    `json:"query"`
	From     time.Time `json:"from" format:"date-time"`
	To       time.Time `json:"to" format:"date-time"`
	Interval string    `json:"interval"`
} //	@name	seqapi.v1.GetHistogramRequest

func (r getHistogramRequest) toProto() *seqapi.GetHistogramRequest {
	return &seqapi.GetHistogramRequest{
		Query:    r.Query,
		From:     timestamppb.New(r.From),
		To:       timestamppb.New(r.To),
		Interval: r.Interval,
	}
}

type histogramBucket struct {
	Key      string `json:"key" format:"uint64"`
	DocCount string `json:"docCount" format:"uint64"`
} //	@name	seqapi.v1.HistogramBucket

func histogramBucketFromProto(proto *seqapi.Histogram_Bucket) histogramBucket {
	return histogramBucket{
		Key:      strconv.FormatUint(proto.GetKey(), 10),
		DocCount: strconv.FormatUint(proto.GetDocCount(), 10),
	}
}

type histogramBuckets []histogramBucket

func histogramBucketsFromProto(proto []*seqapi.Histogram_Bucket) histogramBuckets {
	res := make(histogramBuckets, len(proto))
	for i, hb := range proto {
		res[i] = histogramBucketFromProto(hb)
	}
	return res
}

type histogram struct {
	Buckets histogramBuckets `json:"buckets"`
} //	@name	seqapi.v1.Histogram

func histogramFromProto(proto *seqapi.Histogram, alwaysCreate bool) *histogram {
	if proto == nil && !alwaysCreate {
		return nil
	}
	return &histogram{
		Buckets: histogramBucketsFromProto(proto.GetBuckets()),
	}
}

type getHistogramResponse struct {
	Histogram       histogram `json:"histogram"`
	Error           apiError  `json:"error"`
	PartialResponse bool      `json:"partialResponse"`
} //	@name	seqapi.v1.GetHistogramResponse

func getHistogramResponseFromProto(proto *seqapi.GetHistogramResponse) getHistogramResponse {
	return getHistogramResponse{
		Histogram:       *histogramFromProto(proto.GetHistogram(), true),
		Error:           apiErrorFromProto(proto.GetError()),
		PartialResponse: proto.GetPartialResponse(),
	}
}
