package grpc

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"

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

const defEnv = "default"

type envParam struct {
	client        seqdb.Client
	options       config.SeqAPIOptions
	fieldsCache   *fieldsCache
	masker        *mask.Masker
	pinnedFields  []*seqapi.Field
	exportLimiter *tokenlimiter.Limiter
}

type API struct {
	seqapi.UnimplementedSeqAPIServiceServer

	config              config.SeqAPI
	seqDBСlients        map[string]seqdb.Client
	envParams           map[string]envParam
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
	globalExportLimiter := tokenlimiter.New(cfg.MaxParallelExportRequests)

	envParams := make(map[string]envParam)
	if len(cfg.Envs) > 0 {
		for envName, envConfig := range cfg.Envs {
			client, exists := seqDBСlients[envConfig.SeqDB]
			if !exists {
				logger.Fatal("client not found for environment",
					zap.String("env", envName),
					zap.String("clientID", envConfig.SeqDB))
			}

			options := envConfig.Options
			if options == nil {
				options = cfg.SeqAPIOptions
			}

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

			envParams[envName] = envParam{
				client:        client,
				options:       *options,
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

		envParams["default"] = envParam{
			client:        client,
			options:       *cfg.SeqAPIOptions,
			fieldsCache:   globalfCache,
			masker:        globalMasker,
			pinnedFields:  globalPinnedFields,
			exportLimiter: globalExportLimiter,
		}
	}
	// for export
	for envName := range seqDBСlients {
		envParams := envParams[envName]
		if envParams.masker != nil {
			client := seqDBСlients[envName]
			client.WithMasking(envParams.masker)
			seqDBСlients[envName] = client
		}
	}

	return &API{
		config:              cfg,
		seqDBСlients:        seqDBСlients,
		envParams:           envParams,
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
		envs = make([]*seqapi.GetEnvsResponse_Env, 0, len(cfg.Envs))
		for envName, envConfig := range cfg.Envs {
			envs = append(envs, createEnvInfo(envName, envConfig.Options))
		}
	} else {
		envs = []*seqapi.GetEnvsResponse_Env{createEnvInfo("default", cfg.SeqAPIOptions)}
	}
	return &seqapi.GetEnvsResponse{Envs: envs}
}

func (a *API) GetEnvFromContext(ctx context.Context) (string, error) {
	md, _ := metadata.FromIncomingContext(ctx)

	envValues := md.Get("env")
	if len(envValues) == 0 {
		if len(a.config.Envs) == 0 {
			return defEnv, nil
		} else {
			return a.config.DefaultEnv, nil
		}
	}

	if len(a.config.Envs) == 0 {
		return defEnv, nil
	}

	if _, exists := a.config.Envs[envValues[0]]; !exists {
		return "", fmt.Errorf("env '%s' not found in configuration", envValues[0])
	}
	return envValues[0], nil
}

func (a *API) GetEnvParams(env string) (envParam, error) {
	if env == "" {
		if len(a.config.Envs) == 0 {
			env = defEnv
		} else {
			env = a.config.DefaultEnv
		}
	}

	params, exists := a.envParams[env]
	if !exists {
		return envParam{}, fmt.Errorf("env '%s' not found", env)
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

func checkEnv(env string) string {
	if env == "" {
		return "default"
	}
	return env
}
