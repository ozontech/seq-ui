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

// serveCancelAsyncSearch go doc.
//
//	@Router		/seqapi/v1/async_search/{id}/cancel [post]
//	@ID			seqapi_v1_cancel_async_search
//	@Tags		seqapi_v1
//	@Param		id		path		string			true	"search id"
//	@Success	200		{object}	nil				"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveCancelAsyncSearch(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "seqapi_v1_cancel_async_search")
	defer span.End()

	wr := httputil.NewWriter(w)
	env := types.GetEnvKey(ctx)

	asyncSearches, ok := a.asyncSearches[env]
	if !ok {
		wr.Error(types.ErrAsyncSearchesDisabled, http.StatusBadRequest)
		return
	}

	searchID := chi.URLParam(r, "id")
	// я бы в service вынес проверочку и попробовал перейти на стандартную либу uuid(она вроде с этой версии появилось)
	// https://go.dev/doc/go1.27#uuid такая вот фигня
	if err := checkUUID(searchID); err != nil {
		wr.Error(err, http.StatusBadRequest)
		return
	}

	attributes := []attribute.KeyValue{
		{
			Key:   "search_id",
			Value: attribute.StringValue(searchID),
		},
	}

	if env != "" {
		attributes = append(attributes, attribute.String("env", env))
	}

	span.SetAttributes(attributes...)

	_, err := asyncSearches.CancelAsyncSearch(ctx, &seqapi.CancelAsyncSearchRequest{SearchId: searchID})
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
