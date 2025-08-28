package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// serveGetMy go doc.
//
//	@Router		/dashboards/v1/my [post]
//	@ID			dashboards_v1_getMy
//	@Tags		dashboards_v1
//	@Param		body	body		getMyRequest	true	"Request body"
//	@Success	200		{object}	getMyResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetMy(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "dashboards_v1_get_my")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq getMyRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse request: %w", err), http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "limit",
			Value: attribute.IntValue(httpReq.Limit),
		},
		attribute.KeyValue{
			Key:   "offset",
			Value: attribute.IntValue(httpReq.Offset),
		},
	)

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	req := types.GetUserDashboardsRequest{
		ProfileID: profileID,
		Limit:     httpReq.Limit,
		Offset:    httpReq.Offset,
	}
	dis, err := a.service.GetMyDashboards(ctx, req)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(getMyResponse{
		Dashboards: newInfos(dis),
	})
}

type getMyRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
} // @name dashboards.v1.GetMyRequest

type getMyResponse struct {
	Dashboards infos `json:"dashboards"`
} // @name dashboards.v1.GetMyResponse
