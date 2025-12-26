package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/tracing"
)

// serveGetGroups go doc.
//
//	@Router		/errorgroups/v1/groups [post]
//	@ID			errorgroups_v1_get_groups
//	@Tags		errorgroups_v1
//	@Param		body	body		getGroupsRequest	true	"Request body"
//	@Success	200		{object}	getGroupsResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error		"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetGroups(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "errorgroups_v1_get_groups")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq getGroupsRequest
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
		{Key: "service", Value: attribute.StringValue(httpReq.Service)},
		{Key: "limit", Value: attribute.IntValue(int(httpReq.Limit))},
		{Key: "offset", Value: attribute.IntValue(int(httpReq.Offset))},
		{Key: "order", Value: attribute.StringValue(string(httpReq.Order))},
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

	req := types.GetErrorGroupsRequest{
		Service:   httpReq.Service,
		Env:       httpReq.Env,
		Source:    httpReq.Source,
		Release:   httpReq.Release,
		Duration:  parsedDuration,
		Limit:     httpReq.Limit,
		Offset:    httpReq.Offset,
		Order:     httpReq.Order.toDomain(),
		WithTotal: httpReq.WithTotal,
	}
	groups, total, err := a.service.GetErrorGroups(ctx, req)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(getGroupsResponse{
		Total:  total,
		Groups: newGroups(groups),
	})
}

type order string //	@name	errorgroups.v1.Order

const (
	OrderFrequent order = "frequent"
	OrderLatest   order = "latest"
	OrderOldest   order = "oldest"
)

func (o order) toDomain() types.ErrorGroupsOrder {
	switch o {
	case OrderFrequent:
		return types.OrderFrequent
	case OrderLatest:
		return types.OrderLatest
	case OrderOldest:
		return types.OrderOldest
	default:
		return types.OrderFrequent
	}
}

type getGroupsRequest struct {
	Service string  `json:"service"`
	Env     *string `json:"env,omitempty"`
	Source  *string `json:"source,omitempty"`
	Release *string `json:"release,omitempty"`
	// In go duration format. If not specified, then for the entire time.
	Duration  *string `json:"duration,omitempty" format:"duration" example:"1h"`
	Limit     uint32  `json:"limit"`
	Offset    uint32  `json:"offset"`
	Order     order   `json:"order"`
	WithTotal bool    `json:"with_total"`
} //	@name	errorgroups.v1.GetGroupsRequest

type getGroupsResponse struct {
	Total     uint64   `json:"total"`
	Groups    []group  `json:"groups"`
	NewGroups []uint64 `json:"new_groups"`
} //	@name	errorgroups.v1.GetGroupsResponse

type group struct {
	Hash        string    `json:"hash" format:"uint64"`
	Message     string    `json:"message"`
	SeenTotal   uint64    `json:"seen_total"`
	FirstSeenAt time.Time `json:"first_seen_at" format:"date-time"`
	LastSeenAt  time.Time `json:"last_seen_at" format:"date-time"`
	Source      string    `json:"source"`
} //	@name	errorgroups.v1.Group

func newGroups(source []types.ErrorGroup) []group {
	groups := make([]group, 0, len(source))

	for _, g := range source {
		groups = append(groups, group{
			Hash:        strconv.FormatUint(g.Hash, 10),
			Message:     g.Message,
			SeenTotal:   g.SeenTotal,
			FirstSeenAt: g.FirstSeenAt,
			LastSeenAt:  g.LastSeenAt,
			Source:      g.Source,
		})
	}

	return groups
}
