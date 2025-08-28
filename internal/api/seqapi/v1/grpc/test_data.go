package grpc

import (
	"context"

	"github.com/ozontech/seq-ui/internal/api/profiles"
	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/pkg/repository"
	"github.com/ozontech/seq-ui/internal/pkg/service"
	asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches"
)

func initTestAPI(data test.APITestData) *API {
	return New(data.Cfg, data.Mocks.SeqDB, data.Mocks.Cache, data.Mocks.Cache, nil, nil)
}

func initTestAPIWithAsyncSearches(data test.APITestData) *API {
	as := asyncsearches.New(context.Background(), data.Mocks.AsyncSearchesRepo, data.Mocks.SeqDB)
	s := service.New(&repository.Repository{
		UserProfiles: data.Mocks.ProfilesRepo,
	})
	p := profiles.New(s)
	return New(data.Cfg, data.Mocks.SeqDB, data.Mocks.Cache, data.Mocks.Cache, as, p)
}
