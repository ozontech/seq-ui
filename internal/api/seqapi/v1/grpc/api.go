package grpc

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"

	"github.com/ozontech/seq-ui/internal/api/profiles"
	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	"github.com/ozontech/seq-ui/internal/pkg/mask"
	asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

type apiParams struct {
	client       seqdb.Client
	options      *config.SeqAPIOptions
	fieldsCache  *fieldsCache
	masker       *mask.Masker
	pinnedFields []*seqapi.Field
}

type API struct {
	seqapi.UnimplementedSeqAPIServiceServer

	config              config.SeqAPI
	apiParams           apiParams
	apiParamsByEnv      map[string]apiParams
	inmemWithRedisCache cache.Cache
	redisCache          cache.Cache
	nowFn               func() time.Time
	asyncSearches       *asyncsearches.Service
	profiles            *profiles.Profiles
	envsResponse        *seqapi.GetEnvsResponse
}

func New(
	cfg config.SeqAPI,
	seqDBСlients map[string]seqdb.Client,
	inmemWithRedisCache cache.Cache,
	redisCache cache.Cache,
	asyncSearches *asyncsearches.Service,
	p *profiles.Profiles,
) *API {
	var globalfCache *fieldsCache
	if cfg.FieldsCacheTTL > 0 {
		globalfCache = newFieldsCache(cfg.FieldsCacheTTL)
	}

	globalMasker, err := mask.New(cfg.Masking)
	if err != nil {
		logger.Fatal("failed to init masking", zap.Error(err))
	}

	globalPinnedFields := parsePinnedFields(cfg.PinnedFields)

	var params apiParams
	var paramsByEnv map[string]apiParams

	if len(cfg.Envs) > 0 {
		paramsByEnv = make(map[string]apiParams)
		for envName, envConfig := range cfg.Envs {
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

			paramsByEnv[envName] = apiParams{
				client:       client,
				options:      options,
				fieldsCache:  envfCache,
				masker:       envMasker,
				pinnedFields: envPinnedFields,
			}
		}
	} else {
		client, exists := seqDBСlients[config.DefaultSeqDBClientID]
		if !exists {
			logger.Fatal("default client not found",
				zap.String("clientID", config.DefaultSeqDBClientID))
		}

		params = apiParams{
			client:       client,
			options:      cfg.SeqAPIOptions,
			fieldsCache:  globalfCache,
			masker:       globalMasker,
			pinnedFields: globalPinnedFields,
		}
	}

	return &API{
		config:              cfg,
		apiParams:           params,
		apiParamsByEnv:      paramsByEnv,
		inmemWithRedisCache: inmemWithRedisCache,
		redisCache:          redisCache,
		nowFn:               time.Now,
		asyncSearches:       asyncSearches,
		profiles:            p,
		envsResponse:        parseEnvs(cfg),
	}
}

func parsePinnedFields(fields []config.PinnedField) []*seqapi.Field {
	res := make([]*seqapi.Field, len(fields))
	for i, f := range fields {
		res[i] = &seqapi.Field{
			Name: f.Name,
			Type: seqdb.FieldTypeToProto(f.Type),
		}
	}
	return res
}

func parseEnvs(cfg config.SeqAPI) *seqapi.GetEnvsResponse {
	var envs []*seqapi.GetEnvsResponse_Env
	if len(cfg.Envs) > 0 {
		// sort environment names to ensure deterministic output
		names := make([]string, 0, len(cfg.Envs))
		for name := range cfg.Envs {
			names = append(names, name)
		}
		sort.Strings(names)

		envs = make([]*seqapi.GetEnvsResponse_Env, 0, len(cfg.Envs))
		for _, name := range names {
			envConfig := cfg.Envs[name]
			envs = append(envs, createEnvInfo(name, envConfig.Options))
		}
	} else {
		envs = []*seqapi.GetEnvsResponse_Env{createEnvInfo("", cfg.SeqAPIOptions)}
	}
	return &seqapi.GetEnvsResponse{Envs: envs}
}

func (a *API) GetEnvFromContext(ctx context.Context) string {
	md, _ := metadata.FromIncomingContext(ctx)
	envValues := md.Get("env")
	if len(envValues) == 0 {
		return ""
	}
	return envValues[0]
}

func (a *API) GetParams(env string) (apiParams, error) {
	if len(a.config.Envs) == 0 {
		return a.apiParams, nil
	}

	if env == "" {
		env = a.config.DefaultEnv
	}

	params, exists := a.apiParamsByEnv[env]
	if !exists {
		return apiParams{}, fmt.Errorf("env '%s' not found", env)
	}

	return params, nil
}

type fieldsCache struct {
	ttl time.Duration

	ts     time.Time
	fields []*seqapi.Field
}

func newFieldsCache(ttl time.Duration) *fieldsCache {
	return &fieldsCache{ttl: ttl}
}

func (c *fieldsCache) getFields() ([]*seqapi.Field, bool, bool) {
	return c.fields, !c.ts.IsZero(), time.Since(c.ts) < c.ttl
}

func (c *fieldsCache) setFields(fields []*seqapi.Field) {
	c.fields = fields
	c.ts = time.Now()
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

func createEnvInfo(envName string, opts *config.SeqAPIOptions) *seqapi.GetEnvsResponse_Env {
	return &seqapi.GetEnvsResponse_Env{
		Env:                       envName,
		MaxSearchLimit:            uint32(opts.MaxSearchLimit),
		MaxExportLimit:            uint32(opts.MaxExportLimit),
		MaxParallelExportRequests: uint32(opts.MaxParallelExportRequests),
		MaxAggregationsPerRequest: uint32(opts.MaxAggregationsPerRequest),
		SeqCliMaxSearchLimit:      uint32(opts.SeqCLIMaxSearchLimit),
	}
}
