package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"

	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
	"github.com/ozontech/seq-ui/tracing"
)

// serveDeleteAsyncSearch go doc.
//
//	@Router		/seqapi/v1/async_search/{id} [delete]
//	@ID			seqapi_v1_delete_async_search
//	@Tags		seqapi_v1
//	@Param		env		query		string			true	"Environment"
//	@Param		id		path		string			true	"search id"
//	@Success	200		{object}	nil				"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveDeleteAsyncSearch(w http.ResponseWriter, r *http.Request) {
	wr := httputil.NewWriter(w)

	if a.asyncSearches == nil {
		wr.Error(types.ErrAsyncSearchesDisabled, http.StatusBadRequest)
		return
	}

	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_delete_async_search")
	defer span.End()

	searchID := chi.URLParam(r, "id")

	if err := checkUUID(searchID); err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.KeyValue{
		Key:   "search_id",
		Value: attribute.StringValue(searchID),
	})

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	_, err = a.asyncSearches.DeleteAsyncSearch(ctx, profileID, &seqapi.DeleteAsyncSearchRequest{
		SearchId: searchID,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, types.ErrPermissionDenied) {
			status = http.StatusUnauthorized
		}
		wr.Error(err, status)
		return
	}

	w.WriteHeader(http.StatusOK)
}
