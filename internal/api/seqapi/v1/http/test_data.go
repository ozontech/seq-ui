package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config/v2"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches"
)

// Shared test data.
var (
	errSomethingWrong = errors.New("something happened wrong")
	testSearchID      = "69e4a4a6-0922-43bd-952d-060a86c2b622"
	testQuery         = "message:error"
	testTimestamp     = time.Date(2023, time.September, 25, 10, 20, 30, 0, time.UTC)
)

func setupTestAPI(data test.APITestData) *API {
	seqDBClients := make(map[string]seqdb.Client, max(1, len(data.Cfg.Envs)))
	asyncSearhes := make(map[string]asyncsearches.Service, max(1, len(data.Cfg.Envs)))

	if len(data.Cfg.Envs) == 0 {
		seqDBClients[config.DefaultSeqDBClientID] = data.Mocks.SeqDB
		asyncSearhes[""] = data.Mocks.AsyncSearchesSvc
	}
	for envName, envConfig := range data.Cfg.Envs {
		seqDBClients[envConfig.SeqDBID] = data.Mocks.SeqDB
		asyncSearhes[envName] = data.Mocks.AsyncSearchesSvc
	}

	if data.Mocks.AsyncSearchesSvc == nil {
		asyncSearhes = nil
	}

	api := New(&data.Cfg, seqDBClients, asyncSearhes)

	if data.Mocks.Cache != nil {
		for env, params := range api.paramsByEnv {
			params.eventsCache = data.Mocks.Cache
			params.logsLifespanCache = data.Mocks.Cache
			params.fieldsCache = newFieldsCache(data.Mocks.Cache)
			api.paramsByEnv[env] = params
		}
	}

	return api
}

func withQueryParamID(h http.HandlerFunc, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rCtx := chi.NewRouteContext()
		rCtx.URLParams.Add("id", id)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rCtx))
		h(w, r)
	}
}
