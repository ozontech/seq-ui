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
//	@Router		/errorgroups/v1/services [post]
//	@ID			errorgroups_v1_get_services
//	@Tags		errorgroups_v1
//	@Param		body	body		getServicesRequest	true	"Request body"
//	@Success	200		{object}	getServicesResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error		"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetServices(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "errorgroups_v1_get_releases")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq getServicesRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse request: %w", err), http.StatusBadRequest)
		return
	}

	attributes := []attribute.KeyValue{
		{Key: "query", Value: attribute.StringValue(httpReq.Query)},
	}
	if httpReq.Env != nil {
		attributes = append(attributes, attribute.KeyValue{Key: "env", Value: attribute.StringValue(*httpReq.Env)})
	}
	span.SetAttributes(attributes...)

	req := types.GetServicesRequest{
		Query:  httpReq.Query,
		Env:    httpReq.Env,
		Limit:  httpReq.Limit,
		Offset: httpReq.Offset,
	}
	services, err := a.service.GetServices(ctx, req)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(getServicesResponse{Services: services})
}

type getServicesRequest struct {
	Query  string  `json:"query"`
	Env    *string `json:"env,omitempty"`
	Limit  uint32  `json:"limit"`
	Offset uint32  `json:"offset"`
} //	@name	errorgroups.v1.GetServicesRequest

type getServicesResponse struct {
	Services []string `json:"services"`
} //	@name	errorgroups.v1.GetServicesResponse
