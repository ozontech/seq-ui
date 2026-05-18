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

// serveGetDetails go doc.
//
//	@Router		/errorgroups/v1/details [post]
//	@ID			errorgroups_v1_get_details
//	@Tags		errorgroups_v1
//	@Param		body	body		getDetailsRequest	true	"Request body"
//	@Success	200		{object}	getDetailsResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error		"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetDetails(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "errorgroups_v1_get_details")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq getDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse request: %w", err), http.StatusBadRequest)
		return
	}

	parsedGroupHash, err := parseGroupHash(&httpReq.GroupHash)
	if err != nil {
		wr.Error(fmt.Errorf("failed to parse group_hash: %w", err), http.StatusBadRequest)
		return
	}

	attributes := []attribute.KeyValue{
		{Key: "group_hash", Value: attribute.StringValue(httpReq.GroupHash)},
	}
	if httpReq.Env != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "env", Value: attribute.StringValue(*httpReq.Env)})
	}
	if httpReq.Service != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "service", Value: attribute.StringValue(*httpReq.Service)})
	}
	if httpReq.Release != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "release", Value: attribute.StringValue(*httpReq.Release)})
	}
	if httpReq.Source != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "source", Value: attribute.StringValue(*httpReq.Source)})
	}
	span.SetAttributes(attributes...)

	request := types.GetErrorGroupDetailsRequest{
		GroupHash: *parsedGroupHash,
		Env:       httpReq.Env,
		Source:    httpReq.Source,
		Service:   httpReq.Service,
		Release:   httpReq.Release,
	}
	details, err := a.service.GetDetails(ctx, request)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(getDetailsResponse{
		GroupHash:     httpReq.GroupHash,
		Message:       details.Message,
		SeenTotal:     details.SeenTotal,
		FirstSeenAt:   details.FirstSeenAt,
		LastSeenAt:    details.LastSeenAt,
		Source:        details.Source,
		LogTags:       details.LogTags,
		Distributions: newDistributions(details.Distributions),
	})
}

type getDetailsRequest struct {
	GroupHash string  `json:"group_hash" format:"uint64"`
	Env       *string `json:"env,omitempty"`
	Service   *string `json:"service,omitempty"`
	Release   *string `json:"release,omitempty"`
	Source    *string `json:"source,omitempty"`
} //	@name	errorgroups.v1.GetDetailsRequest

type getDetailsResponse struct {
	GroupHash     string            `json:"group_hash" format:"uint64"`
	Message       string            `json:"message"`
	SeenTotal     uint64            `json:"seen_total"`
	FirstSeenAt   time.Time         `json:"first_seen_at" format:"date-time"`
	LastSeenAt    time.Time         `json:"last_seen_at" format:"date-time"`
	Source        string            `json:"source"`
	LogTags       map[string]string `json:"log_tags,omitempty"`
	Distributions distributions     `json:"distributions"`
} //	@name	errorgroups.v1.GetDetailsResponse

type distribution struct {
	Value   string `json:"value"`
	Percent uint64 `json:"percent"`
} //	@name	errorgroups.v1.Distribution

type distributions struct {
	ByEnv     []distribution `json:"by_env"`
	BySource  []distribution `json:"by_source"`
	ByService []distribution `json:"by_service"`
	ByRelease []distribution `json:"by_release"`
} //	@name	errorgroups.v1.Distributions

func newDistributions(source types.ErrorGroupDistributions) distributions {
	newDistr := func(ds []types.ErrorGroupDistribution) []distribution {
		if len(ds) == 0 {
			return nil
		}

		res := make([]distribution, 0, len(ds))
		for _, d := range ds {
			res = append(res, distribution{
				Value:   d.Value,
				Percent: d.Percent,
			})
		}

		return res
	}

	return distributions{
		ByEnv:     newDistr(source.ByEnv),
		BySource:  newDistr(source.BySource),
		ByService: newDistr(source.ByService),
		ByRelease: newDistr(source.ByRelease),
	}
}
