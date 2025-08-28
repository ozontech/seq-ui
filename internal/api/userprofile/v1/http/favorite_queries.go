package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ozontech/seq-ui/internal/api/httputil"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// serveGetFavoriteQueries go doc.
//
//	@Router		/userprofile/v1/queries/favorite [get]
//	@ID			userprofile_v1_getFavoriteQueries
//	@Tags		userprofile_v1
//	@Success	200		{object}	getFavoriteQueriesResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error				"An unexpected error response"
//	@Security	bearer
func (a *API) serveGetFavoriteQueries(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "userprofile_v1_get_favorite_queries")
	defer span.End()

	wr := httputil.NewWriter(w)

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	req := types.GetFavoriteQueriesRequest{
		ProfileID: profileID,
	}
	fqs, err := a.service.GetFavoriteQueries(ctx, req)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(getFavoriteQueriesResponse{
		Queries: newFavoriteQueries(fqs),
	})
}

// serveCreateFavoriteQuery go doc.
//
//	@Router		/userprofile/v1/queries/favorite [post]
//	@ID			userprofile_v1_createFavoriteQuery
//	@Tags		userprofile_v1
//	@Param		body	body		createFavoriteQueryRequest	true	"Request body"
//	@Success	200		{object}	createFavoriteQueryResponse	"A successful response"
//	@Failure	default	{object}	httputil.Error				"An unexpected error response"
//	@Security	bearer
func (a *API) serveCreateFavoriteQuery(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "userprofile_v1_create_favorite_query")
	defer span.End()

	wr := httputil.NewWriter(w)

	var httpReq createFavoriteQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&httpReq); err != nil {
		wr.Error(fmt.Errorf("failed to parse request: %w", err), http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.KeyValue{
			Key:   "query",
			Value: attribute.StringValue(httpReq.Query),
		},
		attribute.KeyValue{
			Key:   "name",
			Value: attribute.StringValue(httpReq.GetName()),
		},
		attribute.KeyValue{
			Key:   "relative_from",
			Value: attribute.StringValue(httpReq.GetRelativeFrom()),
		},
	)

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	req := types.GetOrCreateFavoriteQueryRequest{
		ProfileID: profileID,
		Query:     httpReq.Query,
	}
	if httpReq.Name != nil {
		req.Name = *httpReq.Name
	}
	if httpReq.RelativeFrom != nil {
		req.RelativeFrom, err = strconv.ParseUint(*httpReq.RelativeFrom, 10, 64)
		if err != nil {
			wr.Error(errors.New("incorrect favorite query 'relativeFrom' format"), http.StatusBadRequest)
			return
		}
	}
	fqID, err := a.service.GetOrCreateFavoriteQuery(ctx, req)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	wr.WriteJson(createFavoriteQueryResponse{
		ID: strconv.FormatInt(fqID, 10),
	})
}

// serveDeleteFavoriteQuery go doc.
//
//	@Router		/userprofile/v1/queries/favorite/{id} [delete]
//	@ID			userprofile_v1_deleteFavoriteQuery
//	@Tags		userprofile_v1
//	@Param		id		path		string			true	"Favorite Query ID"	Format(int64)
//	@Success	200		{object}	nil				"A successful response"
//	@Failure	default	{object}	httputil.Error	"An unexpected error response"
//	@Security	bearer
func (a *API) serveDeleteFavoriteQuery(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "userprofile_v1_delete_favorite_query")
	defer span.End()

	wr := httputil.NewWriter(w)

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		wr.Error(errors.New("incorrect 'id' format"), http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.KeyValue{
		Key:   "id",
		Value: attribute.Int64Value(id),
	})

	profileID, err := a.profiles.GeIDFromContext(ctx)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	req := types.DeleteFavoriteQueryRequest{
		ID:        id,
		ProfileID: profileID,
	}
	err = a.service.DeleteFavoriteQuery(ctx, req)
	if err != nil {
		httputil.ProcessError(wr, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type favoriteQuery struct {
	ID           string `json:"id" format:"int64"`
	Query        string `json:"query"`
	Name         string `json:"name,omitempty"`
	RelativeFrom string `json:"relativeFrom,omitempty" format:"uint64"`
} // @name userprofile.v1.FavoriteQuery

func newFavoriteQuery(t types.FavoriteQuery) favoriteQuery {
	fq := favoriteQuery{
		ID:    strconv.FormatInt(t.ID, 10),
		Query: t.Query,
		Name:  t.Name,
	}
	if t.RelativeFrom != 0 {
		fq.RelativeFrom = strconv.FormatUint(t.RelativeFrom, 10)
	}
	return fq
}

type favoriteQueries []favoriteQuery

func newFavoriteQueries(t types.FavoriteQueries) favoriteQueries {
	res := make(favoriteQueries, len(t))
	for i, fq := range t {
		res[i] = newFavoriteQuery(fq)
	}
	return res
}

type getFavoriteQueriesResponse struct {
	Queries favoriteQueries `json:"queries"`
} // @name userprofile.v1.GetFavoriteQueriesResponse

type createFavoriteQueryRequest struct {
	Query        string  `json:"query"`
	Name         *string `json:"name"`
	RelativeFrom *string `json:"relativeFrom" format:"uint64"`
} // @name userprofile.v1.CreateFavoriteQueryRequest

func (r createFavoriteQueryRequest) GetName() string {
	if r.Name != nil {
		return *r.Name
	}
	return ""
}

func (r createFavoriteQueryRequest) GetRelativeFrom() string {
	if r.RelativeFrom != nil {
		return *r.RelativeFrom
	}
	return ""
}

type createFavoriteQueryResponse struct {
	ID string `json:"id" format:"int64"`
} // @name userprofile.v1.CreateFavoriteQueryResponse
