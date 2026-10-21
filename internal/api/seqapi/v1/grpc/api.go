package grpc

import (
	"maps"
	"slices"
	"sort"
	"strconv"
	"time"

	"go.uber.org/zap"

	handlercache "github.com/ozontech/seq-ui/internal/api/seqapi/v1/handler_cache"
	"github.com/ozontech/seq-ui/internal/app/config/v2"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	"github.com/ozontech/seq-ui/internal/pkg/mask"
	asyncsearches "github.com/ozontech/seq-ui/internal/pkg/service/async_searches"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

type apiParams struct {
	client       seqdb.Client
	options      *config.SeqAPIOptions
	masker       *mask.Masker
	pinnedFields []*seqapi.Field
	systemFields []*seqapi.Field
	caches       handlercache.Caches
}

type API struct {
	seqapi.UnimplementedSeqAPIServiceServer

	config             *config.SeqAPI
	paramsByEnv        map[string]apiParams
	nowFn              func() time.Time
	asyncSearchesByEnv map[string]asyncsearches.Service
	envsResponse       *seqapi.GetEnvsResponse
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

	return apiParams{
		client:       client,
		options:      options,
		masker:       masker,
		pinnedFields: parseFields(options.PinnedFields),
		systemFields: parseFields(options.SystemFields),
		caches:       handlercache.New(env, options.Caches),
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
		MaxSearchLimit:            uint32(opts.Limits.MaxSearchLimit),
		MaxExportLimit:            uint32(opts.Limits.MaxExportLimit),
		MaxParallelExportRequests: uint32(opts.Limits.MaxParallelExportRequests),
		MaxAggregationsPerRequest: uint32(opts.Limits.MaxAggregationsPerRequest),
		SeqCliMaxSearchLimit:      uint32(opts.Limits.SeqCLIMaxSearchLimit),
	}
}
