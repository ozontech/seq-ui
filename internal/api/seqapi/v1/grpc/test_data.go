package grpc

import (
	"errors"
	"time"

	handlercache "github.com/ozontech/seq-ui/internal/api/seqapi/v1/handler_cache"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config/v2"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches"
)

// Shared test data.
var (
	errSomethingWrong = errors.New("something happened wrong")
	testSearchID      = "69e4a4a6-0922-43bd-952d-060a86c2b622"
	testTimestamp     = time.Date(2023, time.September, 25, 10, 20, 30, 0, time.UTC)
)

func setupTestAPI(data test.APITestData) *API {
	seqDBClients := make(map[string]seqdb.Client, max(1, len(data.Cfg.Envs)))
	asyncSearhesByEnv := make(map[string]asyncsearches.Service, max(1, len(data.Cfg.Envs)))

	if len(data.Cfg.Envs) == 0 {
		seqDBClients[config.DefaultSeqDBClientID] = data.Mocks.SeqDB
		asyncSearhesByEnv[""] = data.Mocks.AsyncSearchesSvc
	}
	for envName, envConfig := range data.Cfg.Envs {
		seqDBClients[envConfig.SeqDBID] = data.Mocks.SeqDB
		asyncSearhesByEnv[envName] = data.Mocks.AsyncSearchesSvc
	}

	if data.Mocks.AsyncSearchesSvc == nil {
		asyncSearhesByEnv = nil
	}

	api := New(&data.Cfg, seqDBClients, asyncSearhesByEnv)

	if data.Mocks.Cache != nil {
		caches := data.Cfg.Options.Caches
		for env, params := range api.paramsByEnv {
			if caches.Events != nil {
				params.caches.Events = &handlercache.Cache{Cache: data.Mocks.Cache, KeyPrefix: "seqapi_event_", TTL: caches.Events.TTL}
			}
			if caches.Fields != nil {
				params.caches.Fields = &handlercache.Cache{Cache: data.Mocks.Cache, KeyPrefix: "seqapi_fields", TTL: caches.Fields.TTL}
			}
			if caches.LogsLifespan != nil {
				params.caches.LogsLifespan = &handlercache.Cache{Cache: data.Mocks.Cache, KeyPrefix: "seqapi_logs_lifespan", TTL: caches.LogsLifespan.TTL}
			}
			api.paramsByEnv[env] = params
		}
	}

	return api
}
