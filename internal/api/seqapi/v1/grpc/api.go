package grpc

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"

	"github.com/ozontech/seq-ui/internal/app/config/v2"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	"github.com/ozontech/seq-ui/internal/pkg/mask"
	asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

type apiParams struct {
	client  seqdb.Client
	options *config.SeqAPIOptions
}

type apiGlobalParams struct {
	fieldsCache          *fieldsCache
	masker               *mask.Masker
	pinnedFields         []*seqapi.Field
	systemFields         []*seqapi.Field
	eventsCacheTTL       time.Duration
	logsLifespanCacheTTL time.Duration
}

type API struct {
	seqapi.UnimplementedSeqAPIServiceServer

	config              *config.SeqAPI
	globalParams        apiGlobalParams
	paramsByEnv         map[string]apiParams
	inmemWithRedisCache cache.Cache
	redisCache          cache.Cache
	nowFn               func() time.Time
	asyncSearches       asyncsearches.Service
	envsResponse        *seqapi.GetEnvsResponse
}

func New(
	cfg *config.SeqAPI,
	seqDBСlients map[string]seqdb.Client,
	inmemWithRedisCache cache.Cache,
	redisCache cache.Cache,
	asyncSearches asyncsearches.Service,
) *API {
	globalParams := apiGlobalParams{
		pinnedFields:         parseFields(cfg.GlobalOptions.PinnedFields),
		systemFields:         parseFields(cfg.GlobalOptions.SystemFields),
		eventsCacheTTL:       cfg.GlobalOptions.EventsCacheTTL,
		logsLifespanCacheTTL: cfg.GlobalOptions.LogsLifespanCacheTTL,
	}

	if cfg.GlobalOptions.FieldsCacheTTL > 0 {
		globalParams.fieldsCache = newFieldsCache(cfg.GlobalOptions.FieldsCacheTTL)
	}

	globalMasker, err := mask.New(cfg.GlobalOptions.Masking)
	if err != nil {
		logger.Fatal("failed to init masking", zap.Error(err))
	}
	globalParams.masker = globalMasker

	paramsByEnv := make(map[string]apiParams, max(1, len(cfg.Envs)))
	if len(cfg.Envs) > 0 {
		for envName, envConfig := range cfg.Envs {
			paramsByEnv[envName] = apiParams{
				client:  seqDBСlients[envConfig.SeqDBID],
				options: envConfig.Options,
			}
		}
	} else {
		paramsByEnv[""] = apiParams{
			client:  seqDBСlients[cfg.SeqDBID],
			options: &cfg.Options,
		}
	}

	return &API{
		config:              cfg,
		globalParams:        globalParams,
		paramsByEnv:         paramsByEnv,
		inmemWithRedisCache: inmemWithRedisCache,
		redisCache:          redisCache,
		nowFn:               time.Now,
		asyncSearches:       asyncSearches,
		envsResponse:        parseEnvs(cfg),
	}
}

func parseFields(fields []config.Field) []*seqapi.Field {
	res := make([]*seqapi.Field, len(fields))
	for i, f := range fields {
		res[i] = &seqapi.Field{
			Name: f.Name,
			Type: seqdb.FieldTypeToProto(f.Type),
		}
	}
	return res
}

func parseEnvs(cfg *config.SeqAPI) *seqapi.GetEnvsResponse {
	var envs []*seqapi.GetEnvsResponse_Env
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

		envs = make([]*seqapi.GetEnvsResponse_Env, 0, len(cfg.Envs))
		for _, name := range names {
			envConfig := cfg.Envs[name]
			envs = append(envs, createEnvInfo(name, envConfig.Options))
		}
	} else {
		envs = []*seqapi.GetEnvsResponse_Env{createEnvInfo("", &cfg.Options)}
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
		return a.paramsByEnv[""], nil
	}

	if env == "" {
		env = a.config.DefaultEnv
	}

	params, exists := a.paramsByEnv[env]
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
