package grpc

import (
	"errors"
	"time"

	"github.com/ozontech/seq-ui/internal/api/seqapi/v1/test"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
)

var (
	errSomethingWrong       = errors.New("something happened wrong")
	errCache                = errors.New("test error")
	errorMsg                = "some err"
	mockSearchID            = "69e4a4a6-0922-43bd-952d-060a86c2b622"
	mockSearchID2           = "9e4c068e-d4f4-4a5d-be27-a6524a70d70d"
	mockUserName1           = "some_user_1"
	mockUserName2           = "some_user_2"
	id1                     = "test1"
	id2                     = "test2"
	id3                     = "test3"
	id4                     = "test4"
	query                   = "message:error"
	cacheKey                = "logs_lifespan"
	targetBucketRate        = "3s"
	interval                = "2s"
	resultStr               = "36000" // 10(h) * 60(min/h) * 60(sec/min)
	meta                    = `{"some":"meta"}`
	someMoment              = time.Now()
	from                    = time.Now()
	to                      = from.Add(time.Second)
	cacheTTL                = time.Minute
	ttl                     = 10 * time.Millisecond
	result                  = 10 * time.Hour
	limit             int32 = 3
)

func setupAPI(data test.APITestData) *API {
	// when test cases don't explicitly provide configuration
	if data.Cfg.SeqAPIOptions == nil {
		data.Cfg.SeqAPIOptions = &config.SeqAPIOptions{}
	}
	seqDBClients := make(map[string]seqdb.Client)
	seqDBClients[config.DefaultSeqDBClientID] = data.Mocks.SeqDB

	for _, envConfig := range data.Cfg.Envs {
		seqDBClients[envConfig.SeqDB] = data.Mocks.SeqDB
	}

	return New(data.Cfg, seqDBClients, data.Mocks.Cache, data.Mocks.Cache, nil)
}

func setupAPIWithAsyncSearches(data test.APITestData) *API {
	if data.Cfg.SeqAPIOptions == nil {
		data.Cfg.SeqAPIOptions = &config.SeqAPIOptions{}
	}
	seqDBClients := map[string]seqdb.Client{
		config.DefaultSeqDBClientID: data.Mocks.SeqDB,
	}

	return New(data.Cfg, seqDBClients, data.Mocks.Cache, data.Mocks.Cache, data.Mocks.AsyncSearchesSvc)
}
