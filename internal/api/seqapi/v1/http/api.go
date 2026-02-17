package http

import (
	"fmt"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"

	"github.com/ozontech/seq-ui/internal/api/profiles"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/tokenlimiter"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	"github.com/ozontech/seq-ui/internal/pkg/mask"
	asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

type API struct {
	config              config.SeqAPI
	seqDBСlients        map[string]seqdb.Client
	inmemWithRedisCache cache.Cache
	redisCache          cache.Cache
	nowFn               func() time.Time
	fieldsCache         *fieldsCache
	pinnedFields        fields
	exportLimiter       *tokenlimiter.Limiter
	asyncSearches       *asyncsearches.Service
	profiles            *profiles.Profiles
	masker              *mask.Masker
}

func New(
	cfg config.SeqAPI,
	seqDBСlients map[string]seqdb.Client,
	inmemWithRedisCache cache.Cache,
	redisCache cache.Cache,
	asyncSearches *asyncsearches.Service,
	p *profiles.Profiles,
) *API {
	var fCache *fieldsCache
	if cfg.FieldsCacheTTL > 0 {
		fCache = newFieldsCache(cfg.FieldsCacheTTL)
	}

	masker, err := mask.New(cfg.Masking)
	if err != nil {
		logger.Fatal("failed to init masking", zap.Error(err))
	}
	// for export
	if masker != nil {
		for env, client := range seqDBСlients {
			client.WithMasking(masker)
			seqDBСlients[env] = client
		}
	}

	return &API{
		config:              cfg,
		seqDBСlients:        seqDBСlients,
		inmemWithRedisCache: inmemWithRedisCache,
		redisCache:          redisCache,
		nowFn:               time.Now,
		fieldsCache:         fCache,
		pinnedFields:        parsePinnedFields(cfg.PinnedFields),
		exportLimiter:       tokenlimiter.New(cfg.MaxParallelExportRequests),
		asyncSearches:       asyncSearches,
		profiles:            p,
		masker:              masker,
	}
}

func (a *API) Router() chi.Router {
	mux := chi.NewMux()

	mux.Use(a.envInterceptor)

	mux.Post("/aggregation", a.serveGetAggregation)
	mux.Post("/aggregation_ts", a.serveGetAggregationTs)
	mux.Get("/events/{id}", a.serveGetEvent)
	mux.Post("/export", a.serveExport)
	mux.Get("/fields", a.serveGetFields)
	mux.Get("/fields/pinned", a.serveGetPinnedFields)
	mux.Post("/histogram", a.serveGetHistogram)
	mux.Get("/limits", a.serveGetLimits)
	mux.Post("/search", a.serveSearch)
	mux.Get("/status", a.serveStatus)
	mux.Get("/logs_lifespan", a.serveGetLogsLifespan)
	mux.Get("/envs", a.serveGetEnvs)

	// async searches
	mux.Post("/async_search/start", a.serveStartAsyncSearch)
	mux.Post("/async_search/fetch", a.serveFetchAsyncSearchResult)
	mux.Post("/async_search/list", a.serveGetAsyncSearchesList)
	mux.Post("/async_search/{id}/cancel", a.serveCancelAsyncSearch)
	mux.Delete("/async_search/{id}", a.serveDeleteAsyncSearch)

	return mux
}

func parsePinnedFields(fields []config.PinnedField) []field {
	res := make([]field, len(fields))
	for i, f := range fields {
		res[i] = field{
			Name: f.Name,
			Type: f.Type,
		}
	}
	return res
}

type apiErrorCode string //	@name	seqapi.v1.ErrorCode

const (
	aecNo                  apiErrorCode = "ERROR_CODE_NO"
	aecPartialResponse     apiErrorCode = "ERROR_CODE_PARTIAL_RESPONSE"
	aecQueryTooHeavy       apiErrorCode = "ERROR_CODE_QUERY_TOO_HEAVY"
	aecTooManyFractionsHit apiErrorCode = "ERROR_CODE_TOO_MANY_FRACTIONS_HIT"
)

func (a *API) GetClientFromEnv(env string) (seqdb.Client, *config.SeqAPIOptions, error) {
	envConfig, exists := a.config.Envs[env]
	if !exists {
		return nil, nil, fmt.Errorf("env '%s' not found in configuration", env)
	}

	client, exists := a.seqDBСlients[envConfig.SeqDB]
	if !exists {
		return nil, nil, fmt.Errorf("seqdb client '%s' not found for env '%s'", envConfig.SeqDB, env)
	}

	options := envConfig.Options

	return client, options, nil
}

func apiErrorCodeFromProto(proto seqapi.ErrorCode) apiErrorCode {
	switch proto {
	case seqapi.ErrorCode_ERROR_CODE_PARTIAL_RESPONSE:
		return aecPartialResponse
	case seqapi.ErrorCode_ERROR_CODE_QUERY_TOO_HEAVY:
		return aecQueryTooHeavy
	case seqapi.ErrorCode_ERROR_CODE_TOO_MANY_FRACTIONS_HIT:
		return aecTooManyFractionsHit
	default:
		return aecNo
	}
}

type apiError struct {
	Code    apiErrorCode `json:"code" default:"ERROR_CODE_NO"`
	Message string       `json:"message,omitempty"`
} //	@name	seqapi.v1.Error

func apiErrorFromProto(proto *seqapi.Error) apiError {
	return apiError{
		Code:    apiErrorCodeFromProto(proto.GetCode()),
		Message: proto.GetMessage(),
	}
}

type fieldsCache struct {
	ttl time.Duration

	ts        time.Time
	rawFields []byte
}

func newFieldsCache(ttl time.Duration) *fieldsCache {
	return &fieldsCache{ttl: ttl}
}

func (c *fieldsCache) getFields() ([]byte, bool, bool) {
	return c.rawFields, !c.ts.IsZero(), time.Since(c.ts) < c.ttl
}

func (c *fieldsCache) setFields(rawFields []byte) {
	c.rawFields = rawFields
	c.ts = time.Now()
}

type asyncSearchStatus string //	@name	seqapi.v1.AsyncSearchStatus

const (
	AsyncSearchStatusInProgress asyncSearchStatus = "in_progress"
	AsyncSearchStatusDone       asyncSearchStatus = "done"
	AsyncSearchStatusCanceled   asyncSearchStatus = "canceled"
	AsyncSearchStatusError      asyncSearchStatus = "error"
)

func asyncSearchStatusToProto(s asyncSearchStatus) (seqapi.AsyncSearchStatus, error) {
	switch s {
	case AsyncSearchStatusInProgress:
		return seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_IN_PROGRESS, nil
	case AsyncSearchStatusDone:
		return seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE, nil
	case AsyncSearchStatusError:
		return seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_ERROR, nil
	case AsyncSearchStatusCanceled:
		return seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_CANCELED, nil
	default:
		return seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_UNSPECIFIED, types.NewErrInvalidRequestField("unknown async search status")
	}
}

func asyncSearchStatusFromProto(proto seqapi.AsyncSearchStatus) asyncSearchStatus {
	switch proto {
	case seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_DONE:
		return AsyncSearchStatusDone
	case seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_IN_PROGRESS:
		return AsyncSearchStatusInProgress
	case seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_ERROR:
		return AsyncSearchStatusError
	case seqapi.AsyncSearchStatus_ASYNC_SEARCH_STATUS_CANCELED:
		return AsyncSearchStatusCanceled
	default:
		return AsyncSearchStatusInProgress
	}
}

func checkUUID(v string) error {
	if _, err := uuid.FromString(v); err != nil {
		return types.NewErrInvalidRequestField("invalid uuid")
	}
	return nil
}

func checkLimitOffset(limit, offset int32) error {
	if limit < 0 {
		return types.NewErrInvalidRequestField("'limit' must be non-negative")
	}
	if offset < 0 {
		return types.NewErrInvalidRequestField("'offset' must be non-negative")
	}
	return nil
}
