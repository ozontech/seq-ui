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

// serveGetServiceFilters go doc.
//
//	@Router		/errorgroups/v1/service_filters [post]
//	@ID			errorgroups_v1_get_service_filters
//	@Tags		errorgroups_v1
//	@Param		body	body		getServiceFiltersRequest	true	"Request body"
//	@Success	200		{object}	getServiceFiltersResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error				"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetServiceFilters(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "errorgroups_v1_get_service_filters")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq getServiceFiltersRequest
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

	req := types.GetServiceFiltersRequest{
		Service: httpReq.Service,
		Env:     httpReq.Env,
	}
	filters, err := a.service.GetServiceFilters(ctx, req)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(getServiceFiltersResponse{Filters: newServiceFilters(filters)})
}

type getServiceFiltersRequest struct {
	Service string  `json:"service"`
	Env     *string `json:"env,omitempty"`
} //	@name	errorgroups.v1.GetServiceFiltersRequest

type serviceFilter struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
} //	@name	errorgroups.v1.ServiceFilter

type getServiceFiltersResponse struct {
	Filters []serviceFilter `json:"filters"`
} //	@name	errorgroups.v1.GetServiceFiltersResponse

func newServiceFilters(source []types.ServiceFilter) []serviceFilter {
	filters := make([]serviceFilter, 0, len(source))

	for _, f := range source {
		filters = append(filters, serviceFilter{
			Key:    f.Key,
			Values: f.Values,
		})
	}

	return filters
}
