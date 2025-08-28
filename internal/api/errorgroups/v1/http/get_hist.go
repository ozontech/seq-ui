package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/tracing"
)

// serveGetHist go doc.
//
//	@Router		/errorgroups/v1/hist [post]
//	@ID			errorgroups_v1_get_hist
//	@Tags		errorgroups_v1
//	@Param		body	body		getHistRequest	true	"Request body"
//	@Success	200		{object}	getHistResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetHist(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "errorgroups_v1_get_hist")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq getHistRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse request: %w", err), http.StatusBadRequest)
		return
	}

	parsedGroupHash, err := parseGroupHash(httpReq.GroupHash)
	if err != nil {
		wr.Error(fmt.Errorf("failed to parse group_hash: %w", err), http.StatusBadRequest)
		return
	}

	parsedDuration, err := parseDuration(httpReq.Duration)
	if err != nil {
		wr.Error(fmt.Errorf("failed to parse duration: %w", err), http.StatusBadRequest)
		return
	}

	attributes := []attribute.KeyValue{
		{Key: "service", Value: attribute.StringValue(httpReq.Service)},
	}
	if httpReq.GroupHash != nil {
		attributes = append(attributes, attribute.KeyValue{
			Key: "group_hash", Value: attribute.StringValue(*httpReq.GroupHash),
		})
	}
	if httpReq.Env != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "env", Value: attribute.StringValue(*httpReq.Env)})
	}
	if httpReq.Release != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "release", Value: attribute.StringValue(*httpReq.Release)})
	}
	if httpReq.Duration != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "duration", Value: attribute.StringValue(*httpReq.Duration)})
	}
	if httpReq.Source != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "source", Value: attribute.StringValue(*httpReq.Source)})
	}
	span.SetAttributes(attributes...)

	req := types.GetErrorHistRequest{
		Service:   httpReq.Service,
		GroupHash: parsedGroupHash,
		Env:       httpReq.Env,
		Source:    httpReq.Source,
		Release:   httpReq.Release,
		Duration:  parsedDuration,
	}
	buckets, err := a.service.GetHist(ctx, req)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(getHistResponse{
		Buckets: newBuckets(buckets),
	})
}

type getHistRequest struct {
	Service   string  `json:"service"`
	GroupHash *string `json:"group_hash,omitempty" format:"uint64"`
	Env       *string `json:"env,omitempty"`
	Source    *string `json:"source,omitempty"`
	Release   *string `json:"release,omitempty"`
	// In go duration format. If not specified, `1h` is used.
	Duration *string `json:"duration,omitempty" format:"duration" example:"1h"`
} //	@name	errorgroups.v1.GetHistRequest

type getHistResponse struct {
	Buckets []bucket `json:"buckets"`
} //	@name	errorgroups.v1.GetHistResponse

type bucket struct {
	Time  time.Time `json:"time" format:"date-time"`
	Count uint64    `json:"count"`
} //	@name	errorgroups.v1.Bucket

func newBuckets(source []types.ErrorHistBucket) []bucket {
	buckets := make([]bucket, 0, len(source))

	for _, b := range source {
		buckets = append(buckets, bucket{
			Time:  b.Time,
			Count: b.Count,
		})
	}

	return buckets
}
