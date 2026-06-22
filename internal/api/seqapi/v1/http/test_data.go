package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches"
)

// Shared test data.
var (
	errSomethingWrong = errors.New("something happened wrong")
	testSearchID      = "69e4a4a6-0922-43bd-952d-060a86c2b622"
	testQuery         = "message:error"
	testFrom          = time.Date(2023, time.September, 25, 10, 20, 30, 0, time.UTC)
	testTo            = testFrom.Add(time.Second)
	testSomeMoment    = time.Date(2023, time.September, 25, 10, 20, 30, 0, time.UTC)
)

func setupTestAPI(data test.APITestData) *API {
	// when test cases don't explicitly provide configuration.
	if data.Cfg.SeqAPIOptions == nil {
		data.Cfg.SeqAPIOptions = &config.SeqAPIOptions{}
	}
	seqDBClients := make(map[string]seqdb.Client)
	seqDBClients[config.DefaultSeqDBClientID] = data.Mocks.SeqDB

	for _, envConfig := range data.Cfg.Envs {
		seqDBClients[envConfig.SeqDB] = data.Mocks.SeqDB
	}

	var asyncSvc asyncsearches.Service
	if data.Mocks.AsyncSearchesSvc != nil {
		asyncSvc = data.Mocks.AsyncSearchesSvc
	}

	return New(data.Cfg, seqDBClients, data.Mocks.Cache, data.Mocks.Cache, asyncSvc)
}

func withQueryParamID(h http.HandlerFunc, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rCtx := chi.NewRouteContext()
		rCtx.URLParams.Add("id", id)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rCtx))
		h(w, r)
	}
}
