package http

import (
	"maps"
	"slices"
	"sort"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"

	handlercache "github.com/ozontech/seq-ui/internal/api/seqapi/v1/handler_cache"
	"github.com/ozontech/seq-ui/internal/app/config/v2"
	"github.com/ozontech/seq-ui/internal/app/tokenlimiter"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	"github.com/ozontech/seq-ui/internal/pkg/mask"
	asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

type apiParams struct {
	client        seqdb.Client
	options       *config.SeqAPIOptions
	masker        *mask.Masker
	pinnedFields  fields
	systemFields  fields
	exportLimiter *tokenlimiter.Limiter
	caches        handlercache.Caches
}

type API struct {
	config             *config.SeqAPI
	paramsByEnv        map[string]apiParams
	nowFn              func() time.Time
	asyncSearchesByEnv map[string]asyncsearches.Service
	envsResponse       getEnvsResponse
}

func New(cfg *config.SeqAPI, seqDBСlients map[string]seqdb.Client, asyncSearchesByEnv map[string]asyncsearches.Service) *API {
	paramsByEnv := make(map[string]apiParams, max(1, len(cfg.Envs)))
	if len(cfg.Envs) > 0 {
		for envName, envConfig := range cfg.Envs {
			paramsByEnv[envName] = newAPIParams(envName, seqDBСlients[envConfig.SeqDBID], envConfig.Options)
		}
	} else {
		client := slices.Collect(maps.Values(seqDBСlients))[0]
		paramsByEnv[""] = newAPIParams("", client, &cfg.Options)
	}

	return &API{
		config:             cfg,
		paramsByEnv:        paramsByEnv,
		nowFn:              time.Now,
		asyncSearchesByEnv: asyncSearchesByEnv,
		envsResponse:       parseEnvs(cfg),
	}
}

func newAPIParams(env string, client seqdb.Client, options *config.SeqAPIOptions) apiParams {
	masker, err := mask.New(options.Masking)
	if err != nil {
		logger.Fatal("failed to init masking", zap.String("env", env), zap.Error(err))
	}

	// for export
	if masker != nil {
		client.WithMasking(masker)
	}

	return apiParams{
		client:        client,
		options:       options,
		masker:        masker,
		pinnedFields:  parseFields(options.PinnedFields),
		systemFields:  parseFields(options.SystemFields),
		exportLimiter: tokenlimiter.New(options.Limits.MaxParallelExportRequests),
		caches:        handlercache.New(env, options.Caches),
	}
}

func (a *API) Router() chi.Router {
	mux := chi.NewMux()

	// seqapi
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
	mux.Post("/async_search/export", a.serveExportAsyncSearch)

	return mux
}

func parseFields(fields []config.Field) []field {
	res := make([]field, len(fields))
	for i, f := range fields {
		res[i] = field{
			Name: f.Name,
			Type: f.Type,
		}
	}
	return res
}

func parseEnvs(cfg *config.SeqAPI) getEnvsResponse {
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
		envs = []envInfo{createEnvInfo("", &cfg.Options)}
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
		MaxSearchLimit:            uint32(opts.Limits.MaxSearchLimit),
		MaxExportLimit:            uint32(opts.Limits.MaxExportLimit),
		MaxParallelExportRequests: uint32(opts.Limits.MaxParallelExportRequests),
		MaxAggregationsPerRequest: uint32(opts.Limits.MaxAggregationsPerRequest),
		SeqCliMaxSearchLimit:      uint32(opts.Limits.SeqCLIMaxSearchLimit),
	}
}
