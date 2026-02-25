package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// serveDiffByReleases go doc.
//
//	@Router		/errorgroups/v1/diff_by_releases [post]
//	@ID			errorgroups_v1_diff_by_releases
//	@Tags		errorgroups_v1
//	@Param		body	body		diffByReleasesRequest	true	"Request body"
//	@Success	200		{object}	diffByReleasesResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error			"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetDiffByReleases(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "errorgroups_v1_diff_by_releases")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq diffByReleasesRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse request: %w", err), http.StatusBadRequest)
		return
	}

	attributes := []attribute.KeyValue{
		{Key: "service", Value: attribute.StringValue(httpReq.Service)},
		{Key: "releases", Value: attribute.StringSliceValue(httpReq.Releases)},
		{Key: "limit", Value: attribute.IntValue(int(httpReq.Limit))},
		{Key: "offset", Value: attribute.IntValue(int(httpReq.Offset))},
		{Key: "order", Value: attribute.StringValue(string(httpReq.Order))},
		{Key: "with_total", Value: attribute.BoolValue(httpReq.WithTotal)},
	}
	if httpReq.Env != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "env", Value: attribute.StringValue(*httpReq.Env)})
	}
	if httpReq.Source != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "source", Value: attribute.StringValue(*httpReq.Source)})
	}
	span.SetAttributes(attributes...)

	groups, total, err := a.service.DiffByReleases(ctx, types.DiffByReleasesRequest{
		Service:   httpReq.Service,
		Releases:  httpReq.Releases,
		Env:       httpReq.Env,
		Source:    httpReq.Source,
		Limit:     httpReq.Limit,
		Offset:    httpReq.Offset,
		Order:     httpReq.Order.toDomain(),
		WithTotal: httpReq.WithTotal,
	})
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(diffByReleasesResponse{
		Total:  total,
		Groups: newDiffGroups(groups),
	})
}

type diffByReleasesRequest struct {
	Service   string   `json:"service"`
	Releases  []string `json:"releases"`
	Env       *string  `json:"env,omitempty"`
	Source    *string  `json:"source,omitempty"`
	Limit     uint32   `json:"limit"`
	Offset    uint32   `json:"offset"`
	Order     order    `json:"order"`
	WithTotal bool     `json:"with_total"`
} //	@name	errorgroups.v1.DiffByReleasesRequest

type diffByReleasesResponse struct {
	Total  uint64      `json:"total"`
	Groups []diffGroup `json:"groups"`
} //	@name	errorgroups.v1.DiffByReleasesResponse

type diffGroup struct {
	Hash        string    `json:"hash" format:"uint64"`
	Message     string    `json:"message"`
	FirstSeenAt time.Time `json:"first_seen_at" format:"date-time"`
	LastSeenAt  time.Time `json:"last_seen_at" format:"date-time"`
	Source      string    `json:"source"`

	ReleaseInfos map[string]diffReleaseInfo `json:"release_infos"`
} //	@name	errorgroups.v1.DiffGroup

type diffReleaseInfo struct {
	SeenTotal uint64 `json:"seen_total"`
} //	@name	errorgroups.v1.DiffReleaseInfo

func newDiffGroups(source []types.DiffGroup) []diffGroup {
	groups := make([]diffGroup, 0, len(source))

	for _, g := range source {
		releaseInfos := make(map[string]diffReleaseInfo)
		for release, info := range g.ReleaseInfos {
			releaseInfos[release] = diffReleaseInfo{
				SeenTotal: info.SeenTotal,
			}
		}
		groups = append(groups, diffGroup{
			Hash:         strconv.FormatUint(g.Hash, 10),
			Message:      g.Message,
			FirstSeenAt:  g.FirstSeenAt,
			LastSeenAt:   g.LastSeenAt,
			Source:       g.Source,
			ReleaseInfos: releaseInfos,
		})
	}

	return groups
}
