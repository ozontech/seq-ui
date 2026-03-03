package http

import (
	"context"

	"github.com/ozontech/seq-ui/internal/api/profiles"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	"github.com/ozontech/seq-ui/internal/pkg/repository"
	"github.com/ozontech/seq-ui/internal/pkg/service"
	asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches"
)

func initTestAPI(data test.APITestData) *API {
	// when test cases don't explicitly provide configuration.
	if data.Cfg.SeqAPIOptions == nil {
		data.Cfg.SeqAPIOptions = &config.SeqAPIOptions{}
	}
	seqDBClients := make(map[string]seqdb.Client)
	seqDBClients[config.DefaultSeqDBClientID] = data.Mocks.SeqDB

	for _, envConfig := range data.Cfg.Envs {
		seqDBClients[envConfig.SeqDB] = data.Mocks.SeqDB
	}

	return New(data.Cfg, seqDBClients, data.Mocks.Cache, data.Mocks.Cache, nil, nil)
}

func initTestAPIWithAsyncSearches(data test.APITestData) *API {
	if data.Cfg.SeqAPIOptions == nil {
		data.Cfg.SeqAPIOptions = &config.SeqAPIOptions{}
	}
	seqDBClients := map[string]seqdb.Client{
		config.DefaultSeqDBClientID: data.Mocks.SeqDB,
	}
	as := asyncsearches.New(context.Background(), data.Mocks.AsyncSearchesRepo, data.Mocks.SeqDB, []string{})
	s := service.New(&repository.Repository{
		UserProfiles: data.Mocks.ProfilesRepo,
	})
	p := profiles.New(s)
	return New(data.Cfg, seqDBClients, data.Mocks.Cache, data.Mocks.Cache, as, p)
}
