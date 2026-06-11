package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/tracing"
)

// serveGetTopGroups go doc.
//
//	@Router		/errorgroups/v1/top_groups [post]
//	@ID			errorgroups_v1_get_top_groups
//	@Tags		errorgroups_v1
//	@Param		body	body		getTopGroupsRequest		true	"Request body"
//	@Success	200		{object}	getTopGroupsResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error			"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetTopGroups(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "errorgroups_v1_get_top_groups")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq getTopGroupsRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse request: %w", err), http.StatusBadRequest)
		return
	}

	parsedDuration, err := parseDuration(httpReq.Duration)
	if err != nil {
		wr.Error(fmt.Errorf("failed to parse duration: %w", err), http.StatusBadRequest)
		return
	}

	attributes := []attribute.KeyValue{
		{Key: "limit", Value: attribute.IntValue(int(httpReq.Limit))},
		{Key: "offset", Value: attribute.IntValue(int(httpReq.Offset))},
		{Key: "with_total", Value: attribute.BoolValue(httpReq.WithTotal)},
	}
	if httpReq.Env != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "env", Value: attribute.StringValue(*httpReq.Env)})
	}
	if httpReq.Source != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "source", Value: attribute.StringValue(*httpReq.Source)})
	}
	if httpReq.Duration != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "duration", Value: attribute.StringValue(*httpReq.Duration)})
	}
	span.SetAttributes(attributes...)

	req := types.GetTopErrorGroupsRequest{
		Env:       httpReq.Env,
		Source:    httpReq.Source,
		Duration:  parsedDuration,
		Limit:     httpReq.Limit,
		Offset:    httpReq.Offset,
		WithTotal: httpReq.WithTotal,
	}

	groups, total, err := a.service.GetTopErrorGroups(ctx, req)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(getTopGroupsResponse{
		Total:  total,
		Groups: newTopGroups(groups),
	})
}

type getTopGroupsRequest struct {
	Env    *string `json:"env,omitempty"`
	Source *string `json:"source,omitempty"`
	// In go duration format. If not specified, then for the entire time.
	Duration  *string `json:"duration,omitempty" format:"duration" example:"1h"`
	Limit     uint32  `json:"limit"`
	Offset    uint32  `json:"offset"`
	WithTotal bool    `json:"with_total"`
} //	@name	errorgroups.v1.GetTopGroupsRequest

type getTopGroupsResponse struct {
	Total  uint64     `json:"total"`
	Groups []topGroup `json:"groups"`
} //	@name	errorgroups.v1.GetTopGroupsResponse

type topGroup struct {
	Hash      string `json:"hash" format:"uint64"`
	Message   string `json:"message"`
	Source    string `json:"source"`
	SeenTotal uint64 `json:"seen_total"`
} //	@name	errorgroups.v1.TopGroup

func newTopGroups(source []types.TopErrorGroup) []topGroup {
	groups := make([]topGroup, 0, len(source))

	for _, g := range source {
		groups = append(groups, topGroup{
			Hash:      strconv.FormatUint(g.Hash, 10),
			Message:   g.Message,
			Source:    g.Source,
			SeenTotal: g.Count,
		})
	}

	return groups
}
