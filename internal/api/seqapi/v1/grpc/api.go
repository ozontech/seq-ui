package grpc

import (
	"context"
	"fmt"
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

type API struct {
	seqapi.UnimplementedSeqAPIServiceServer

	config              config.SeqAPI
	seqDBСlients        map[string]seqdb.Client
	defaultEnv          string
	inmemWithRedisCache cache.Cache
	redisCache          cache.Cache
	nowFn               func() time.Time
	fieldsCache         *fieldsCache
	pinnedFields        []*seqapi.Field
	asyncSearches       *asyncsearches.Service
	profiles            *profiles.Profiles
	masker              *mask.Masker
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
	var fCache *fieldsCache
	if cfg.FieldsCacheTTL > 0 {
		fCache = newFieldsCache(cfg.FieldsCacheTTL)
	}

	masker, err := mask.New(cfg.Masking)
	if err != nil {
		logger.Fatal("failed to init masking", zap.Error(err))
	}

	return &API{
		config:              cfg,
		seqDBСlients:        seqDBСlients,
		defaultEnv:          cfg.DefaultEnv,
		inmemWithRedisCache: inmemWithRedisCache,
		redisCache:          redisCache,
		nowFn:               time.Now,
		fieldsCache:         fCache,
		pinnedFields:        parsePinnedFields(cfg.PinnedFields),
		asyncSearches:       asyncSearches,
		profiles:            p,
		masker:              masker,
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
	envs := make([]*seqapi.GetEnvsResponse_Env, 0, len(cfg.Envs))
	for envName, envConfig := range cfg.Envs {
		env := &seqapi.GetEnvsResponse_Env{
			Env:                       envName,
			MaxSearchLimit:            uint32(envConfig.Options.MaxSearchLimit),
			MaxExportLimit:            uint32(envConfig.Options.MaxExportLimit),
			MaxParallelExportRequests: uint32(envConfig.Options.MaxParallelExportRequests),
			MaxAggregationsPerRequest: uint32(envConfig.Options.MaxAggregationsPerRequest),
			SeqCliMaxSearchLimit:      uint32(envConfig.Options.SeqCLIMaxSearchLimit),
		}
		envs = append(envs, env)
	}
	return &seqapi.GetEnvsResponse{
		Envs: envs,
	}
}

func (a *API) GetEnvFromContext(ctx context.Context) (string, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	envValues := md.Get("env")
	if _, exists := a.config.Envs[envValues[0]]; !exists {
		return "", fmt.Errorf("env '%s' not found in configuration", envValues[0])
	}
	return envValues[0], nil
}

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
