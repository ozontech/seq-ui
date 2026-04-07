package http

import (
	"fmt"
	"sort"
	"strconv"
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

type apiParams struct {
	client        seqdb.Client
	options       *config.SeqAPIOptions
	fieldsCache   *fieldsCache
	masker        *mask.Masker
	pinnedFields  fields
	exportLimiter *tokenlimiter.Limiter
}

type API struct {
	config              config.Handlers
	params              apiParams
	paramsByEnv         map[string]apiParams
	inmemWithRedisCache cache.Cache
	redisCache          cache.Cache
	nowFn               func() time.Time
	asyncSearches       *asyncsearches.Service
	profiles            *profiles.Profiles
	envsResponse        getEnvsResponse
}

func New(
	cfg config.Handlers,
	seqDBСlients map[string]seqdb.Client,
	inmemWithRedisCache cache.Cache,
	redisCache cache.Cache,
	asyncSearches *asyncsearches.Service,
	p *profiles.Profiles,
) *API {
	var globalfCache *fieldsCache
	if cfg.SeqAPI.FieldsCacheTTL > 0 {
		globalfCache = newFieldsCache(cfg.SeqAPI.FieldsCacheTTL)
	}

	globalMasker, err := mask.New(cfg.SeqAPI.Masking)
	if err != nil {
		logger.Fatal("failed to init masking", zap.Error(err))
	}

	globalPinnedFields := parsePinnedFields(cfg.SeqAPI.PinnedFields)
	globalExportLimiter := tokenlimiter.New(cfg.SeqAPI.MaxParallelExportRequests)

	var params apiParams
	var paramsByEnv map[string]apiParams

	if len(cfg.SeqAPI.Envs) > 0 {
		paramsByEnv = make(map[string]apiParams)
		for envName, envConfig := range cfg.SeqAPI.Envs {
			client := seqDBСlients[envConfig.SeqDB]
			options := envConfig.Options

			var envfCache *fieldsCache
			if options.FieldsCacheTTL > 0 {
				envfCache = newFieldsCache(options.FieldsCacheTTL)
			}

			var envMasker, err = mask.New(options.Masking)
			if err != nil {
				logger.Fatal("failed to init env masking", zap.Error(err))
			}

			envPinnedFields := parsePinnedFields(options.PinnedFields)
			envExportLimiter := tokenlimiter.New(options.MaxParallelExportRequests)

			paramsByEnv[envName] = apiParams{
				client:        client,
				options:       options,
				fieldsCache:   envfCache,
				masker:        envMasker,
				pinnedFields:  envPinnedFields,
				exportLimiter: envExportLimiter,
			}
		}
	} else {
		client, exists := seqDBСlients[config.DefaultSeqDBClientID]
		if !exists {
			logger.Fatal("default client not found",
				zap.String("clientID", config.DefaultSeqDBClientID))
		}

		params = apiParams{
			client:        client,
			options:       cfg.SeqAPI.SeqAPIOptions,
			fieldsCache:   globalfCache,
			masker:        globalMasker,
			pinnedFields:  globalPinnedFields,
			exportLimiter: globalExportLimiter,
		}
	}
	// for export
	if len(cfg.SeqAPI.Envs) > 0 {
		for _, param := range paramsByEnv {
			if param.masker != nil {
				param.client.WithMasking(param.masker)
			}
		}
	} else if params.masker != nil {
		params.client.WithMasking(params.masker)
	}

	return &API{
		config:              cfg,
		params:              params,
		paramsByEnv:         paramsByEnv,
		inmemWithRedisCache: inmemWithRedisCache,
		redisCache:          redisCache,
		nowFn:               time.Now,
		asyncSearches:       asyncSearches,
		profiles:            p,
		envsResponse:        parseEnvs(cfg.SeqAPI),
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

func parseEnvs(cfg config.SeqAPI) getEnvsResponse {
	var envs []envInfo
	if len(cfg.Envs) > 0 {
		// sort environment names to ensure deterministic output
		names := make([]string, 0, len(cfg.Envs))
		for name := range cfg.Envs {
			names = append(names, name)
		}
		sort.Slice(names, func(i, j int) bool {
			a, b := names[i], names[j]

			var aPrefix, bPrefix string
			var aNum, bNum int

			k := 0
			for k < len(a) && (a[k] < '0' || a[k] > '9') {
				k++
			}
			aPrefix = a[:k]
			if k < len(a) {
				aNum, _ = strconv.Atoi(a[k:])
			}

			k = 0
			for k < len(b) && (b[k] < '0' || b[k] > '9') {
				k++
			}
			bPrefix = b[:k]
			if k < len(b) {
				bNum, _ = strconv.Atoi(b[k:])
			}

			if aPrefix != bPrefix {
				return aPrefix < bPrefix
			}
			return aNum < bNum
		})

		envs = make([]envInfo, 0, len(cfg.Envs))
		for _, name := range names {
			envConfig := cfg.Envs[name]
			envs = append(envs, createEnvInfo(name, envConfig.Options))
		}
	} else {
		envs = []envInfo{createEnvInfo("", cfg.SeqAPIOptions)}
	}
	return getEnvsResponse{Envs: envs}
}

type apiErrorCode string //	@name	seqapi.v1.ErrorCode

const (
	aecNo                  apiErrorCode = "ERROR_CODE_NO"
	aecUnspecified         apiErrorCode = "ERROR_CODE_UNSPECIFIED"
	aecPartialResponse     apiErrorCode = "ERROR_CODE_PARTIAL_RESPONSE"
	aecQueryTooHeavy       apiErrorCode = "ERROR_CODE_QUERY_TOO_HEAVY"
	aecTooManyFractionsHit apiErrorCode = "ERROR_CODE_TOO_MANY_FRACTIONS_HIT"
)

func (a *API) GetEnvParams(env string) (apiParams, error) {
	if len(a.config.SeqAPI.Envs) == 0 {
		return a.params, nil
	}

	if env == "" {
		env = a.config.SeqAPI.DefaultEnv
	}

	params, exists := a.paramsByEnv[env]
	if !exists {
		return apiParams{}, fmt.Errorf("env '%s' not found", env)
	}

	return params, nil
}

func apiErrorCodeFromProto(proto seqapi.ErrorCode) apiErrorCode {
	switch proto {
	case seqapi.ErrorCode_ERROR_CODE_UNSPECIFIED:
		return aecUnspecified
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

func createEnvInfo(envName string, opts *config.SeqAPIOptions) envInfo {
	return envInfo{
		Env:                       envName,
		MaxSearchLimit:            uint32(opts.MaxSearchLimit),
		MaxExportLimit:            uint32(opts.MaxExportLimit),
		MaxParallelExportRequests: uint32(opts.MaxParallelExportRequests),
		MaxAggregationsPerRequest: uint32(opts.MaxAggregationsPerRequest),
		SeqCliMaxSearchLimit:      uint32(opts.SeqCLIMaxSearchLimit),
	}
}
