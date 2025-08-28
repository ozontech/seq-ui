package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/attribute"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/tracing"
)

// serveGetReleases go doc.
//
//	@Router		/errorgroups/v1/releases [post]
//	@ID			errorgroups_v1_get_releases
//	@Tags		errorgroups_v1
//	@Param		body	body		getReleasesRequest	true	"Request body"
//	@Success	200		{object}	getReleasesResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error		"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetReleases(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "errorgroups_v1_get_releases")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq getReleasesRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse request: %w", err), http.StatusBadRequest)
		return
	}

	attributes := []attribute.KeyValue{
		{Key: "service", Value: attribute.StringValue(httpReq.Service)},
	}
	if httpReq.Env != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "env", Value: attribute.StringValue(*httpReq.Env)})
	}
	span.SetAttributes(attributes...)

	req := types.GetErrorGroupReleasesRequest{
		Service: httpReq.Service,
		Env:     httpReq.Env,
	}
	releases, err := a.service.GetReleases(ctx, req)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(getReleasesResponse{Releases: releases})
}

type getReleasesRequest struct {
	Service string  `json:"service"`
	Env     *string `json:"env,omitempty"`
} //	@name	errorgroups.v1.GetReleasesRequest

type getReleasesResponse struct {
	Releases []string `json:"releases"`
} //	@name	errorgroups.v1.GetReleasesResponse
