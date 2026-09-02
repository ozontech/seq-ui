package config

import (
	"fmt"
	"time"
)

const (
	DefaultSeqDBClientID     = "default_seq_db"
	DefaultCHClientID        = "default_ch"
	DefaultInmemCacheID      = "default_inmem_cache"
	DefaultRedisID           = "default_redis"
	DefaultRedis2ID          = "default_redis_2"
	DefaultMassExportRedisID = "mass_export"

	SeqDBClientModeGRPC = "grpc"

	MaskModeMask    = "mask"
	MaskModeReplace = "replace"
	MaskModeCut     = "cut"

	FieldFilterConditionAnd = "and"
	FieldFilterConditionOr  = "or"
	FieldFilterConditionNot = "not"

	FieldFilterModeEqual    = "equal"
	FieldFilterModeContains = "contains"
	FieldFilterModePrefix   = "prefix"
	FieldFilterModeSuffix   = "suffix"

	minGRPCKeepaliveTime    = 10 * time.Second
	minGRPCKeepaliveTimeout = 1 * time.Second

	defaultAsyncSearchListQueryLengthLimit = 1000

	defaultMaxSearchTotal             = 1000000
	defaultMaxSearchOffset            = 1000000
	defaultMaxExportLimit             = 100000
	defaultMaxAggregationsPerRequest  = 1
	defaultMaxBucketsPerAggregationTs = 200
	defaultMaxParallelExportRequests  = 1

	defaultInmemCacheNumCounters = 10000000
	defaultInmemCacheMaxCost     = 1000000
	defaultInmemCacheBufferItems = 64

	defaultEventsCacheTTL = 24 * time.Hour

	defaultLogsLifespanCacheTTL = 10 * time.Minute

	defaultAdminCacheTTL = time.Hour

	defaultClickHouseDialTimeout = 5 * time.Second
	defaultClickHouseReadTimeout = 30 * time.Second
)

func Normalize(cfg *Config) error {
	if len(cfg.Clients.SeqDB) == 0 {
		return fmt.Errorf("clients.seq_db must contain at least one client")
	}

	seqDBIDs := make(map[string]struct{}, len(cfg.Clients.SeqDB))
	for i := range cfg.Clients.SeqDB {
		c := &cfg.Clients.SeqDB[i]
		if c.ID == "" {
			return fmt.Errorf("clients.seq_db[%d].id cannot be empty", i)
		}
		if _, ok := seqDBIDs[c.ID]; ok {
			return fmt.Errorf("duplicate clients.seq_db.id %q", c.ID)
		}

		seqDBIDs[c.ID] = struct{}{}

		if c.ClientMode == "" {
			c.ClientMode = SeqDBClientModeGRPC
		} else if c.ClientMode != SeqDBClientModeGRPC {
			return fmt.Errorf("invalid clients.seq_db[%q].client_mode: %q (allowed: %q)", c.ID, c.ClientMode, SeqDBClientModeGRPC)
		}

		if c.GRPCKeepaliveParams != nil {
			c.GRPCKeepaliveParams.Time = max(c.GRPCKeepaliveParams.Time, minGRPCKeepaliveTime)
			c.GRPCKeepaliveParams.Timeout = max(c.GRPCKeepaliveParams.Timeout, minGRPCKeepaliveTimeout)
		}
	}

	chIDs := make(map[string]struct{}, len(cfg.Clients.ClickHouse))
	for i := range cfg.Clients.ClickHouse {
		ch := &cfg.Clients.ClickHouse[i]
		if ch.ID == "" {
			return fmt.Errorf("clients.clickhouse[%d].id cannot be empty", i)
		}
		if _, ok := chIDs[ch.ID]; ok {
			return fmt.Errorf("duplicate clients.clickhouse.id %q", ch.ID)
		}

		chIDs[ch.ID] = struct{}{}

		if ch.DialTimeout <= 0 {
			ch.DialTimeout = defaultClickHouseDialTimeout
		}
		if ch.ReadTimeout <= 0 {
			ch.ReadTimeout = defaultClickHouseReadTimeout
		}
	}

	inmemIDs := make(map[string]struct{}, len(cfg.Cache.Inmemory))
	for i := range cfg.Cache.Inmemory {
		inm := &cfg.Cache.Inmemory[i]
		if inm.ID == "" {
			return fmt.Errorf("cache.inmemory[%d].id cannot be empty", i)
		}
		if _, ok := inmemIDs[inm.ID]; ok {
			return fmt.Errorf("duplicate cache.inmemory.id %q", inm.ID)
		}

		inmemIDs[inm.ID] = struct{}{}

		if inm.NumCounters <= 0 {
			inm.NumCounters = defaultInmemCacheNumCounters
		}
		if inm.MaxCost <= 0 {
			inm.MaxCost = defaultInmemCacheMaxCost
		}
		if inm.BufferItems <= 0 {
			inm.BufferItems = defaultInmemCacheBufferItems
		}
	}

	redisIDs := make(map[string]struct{}, len(cfg.Cache.Redis))
	for i := range cfg.Cache.Redis {
		r := &cfg.Cache.Redis[i]
		if r.ID == "" {
			return fmt.Errorf("cache.redis[%d].id cannot be empty", i)
		}
		if _, ok := redisIDs[r.ID]; ok {
			return fmt.Errorf("duplicate cache.redis.id %q", r.ID)
		}

		redisIDs[r.ID] = struct{}{}

		if r.WithInmemID != "" {
			if _, ok := inmemIDs[r.WithInmemID]; !ok {
				return fmt.Errorf("cache.redis[%q].with_inmem_id %q not found in cache.inmemory", r.ID, r.WithInmemID)
			}
		}
	}

	if cfg.DB != nil && cfg.DB.UsePreparedStatements == nil {
		cfg.DB.UsePreparedStatements = new(bool)
		*cfg.DB.UsePreparedStatements = true
	}

	setSeqAPIOptionsDefaults(&cfg.Handlers.SeqAPI.Options)

	if len(cfg.Handlers.SeqAPI.Envs) > 0 {
		if cfg.Handlers.SeqAPI.SeqDBID != "" {
			return fmt.Errorf("handlers.seq_api.seq_db_id must be empty when envs is used. Put seq_db_id inside each env")
		}
		if cfg.Handlers.SeqAPI.DefaultEnv == "" {
			return fmt.Errorf("handlers.seq_api.default_env must be specified when envs is used")
		}
		if _, exists := cfg.Handlers.SeqAPI.Envs[cfg.Handlers.SeqAPI.DefaultEnv]; !exists {
			return fmt.Errorf("handlers.seq_api.default_env %q not found in envs", cfg.Handlers.SeqAPI.DefaultEnv)
		}

		for envName, envConfig := range cfg.Handlers.SeqAPI.Envs {
			if envConfig.SeqDBID == "" {
				return fmt.Errorf("handlers.seq_api.envs[%q].seq_db_id cannot be empty", envName)
			}
			if _, ok := seqDBIDs[envConfig.SeqDBID]; !ok {
				return fmt.Errorf("unknown handlers.seq_api.envs[%q].seq_db_id %q", envName, envConfig.SeqDBID)
			}

			setSeqAPIOptionsDefaults(envConfig.Options)
		}
	} else {
		if cfg.Handlers.SeqAPI.SeqDBID == "" {
			return fmt.Errorf("handlers.seq_api.seq_db_id cannot be empty when envs is not used")
		}
		if _, ok := seqDBIDs[cfg.Handlers.SeqAPI.SeqDBID]; !ok {
			return fmt.Errorf("unknown handlers.seq_api.seq_db_id %q", cfg.Handlers.SeqAPI.SeqDBID)
		}
	}

	if cfg.Handlers.MassExport != nil {
		if cfg.Handlers.MassExport.SeqDBID == "" {
			return fmt.Errorf("handlers.mass_export.seq_db_id cannot be empty")
		}
		if _, ok := seqDBIDs[cfg.Handlers.MassExport.SeqDBID]; !ok {
			return fmt.Errorf("unknown handlers.mass_export.seq_db_id %q", cfg.Handlers.MassExport.SeqDBID)
		}

		if cfg.Handlers.MassExport.SessionStore != nil {
			if cfg.Handlers.MassExport.SessionStore.RedisID == "" {
				return fmt.Errorf("handlers.mass_export.session_store.redis_id cannot be empty")
			}
			if _, ok := redisIDs[cfg.Handlers.MassExport.SessionStore.RedisID]; !ok {
				return fmt.Errorf("unknown handlers.mass_export.session_store.redis_id %q", cfg.Handlers.MassExport.SessionStore.RedisID)
			}
		}
	}

	if cfg.Handlers.AsyncSearch != nil {
		as := cfg.Handlers.AsyncSearch

		if as.ListQueryLengthLimit <= 0 {
			as.ListQueryLengthLimit = defaultAsyncSearchListQueryLengthLimit
		}

		if len(as.Envs) > 0 {
			if as.SeqDBID != "" {
				return fmt.Errorf("handlers.async_search.seq_db_id must be empty when envs is used. Put seq_db_id inside each env")
			}
			if as.DefaultEnv == "" {
				return fmt.Errorf("handlers.async_search.default_env must be specified when using envs")
			}
			if _, ok := as.Envs[as.DefaultEnv]; !ok {
				return fmt.Errorf("handlers.async_search.default_env %q not found in envs", as.DefaultEnv)
			}

			for envName, envConfig := range as.Envs {
				if envConfig.SeqDBID == "" {
					return fmt.Errorf("handlers.async_search.envs[%q].seq_db_id cannot be empty", envName)
				}
				if _, ok := seqDBIDs[envConfig.SeqDBID]; !ok {
					return fmt.Errorf("unknown handlers.async_search.envs[%q].seq_db_id %q", envName, envConfig.SeqDBID)
				}
				if envConfig.ListQueryLengthLimit <= 0 {
					envConfig.ListQueryLengthLimit = as.ListQueryLengthLimit
				}

				as.Envs[envName] = envConfig
			}
		} else {
			if as.SeqDBID == "" {
				return fmt.Errorf("handlers.async_search.seq_db_id cannot be empty when envs is not used")
			}
			if _, ok := seqDBIDs[as.SeqDBID]; !ok {
				return fmt.Errorf("unknown handlers.async_search.seq_db_id %q", as.SeqDBID)
			}
		}
	}

	if cfg.Handlers.ErrorGroups != nil {
		eg := cfg.Handlers.ErrorGroups
		if len(eg.Envs) > 0 {
			if eg.CHID != "" {
				return fmt.Errorf("handlers.error_groups.ch_id must be empty when envs is used. Put ch_id inside each env")
			}
			if eg.DefaultEnv == "" {
				return fmt.Errorf("handlers.error_groups.default_env must be specified when using envs")
			}
			if _, ok := eg.Envs[eg.DefaultEnv]; !ok {
				return fmt.Errorf("handlers.error_groups.default_env %q not found in envs", eg.DefaultEnv)
			}

			for envName, envConfig := range eg.Envs {
				if envConfig.CHID == "" {
					return fmt.Errorf("handlers.error_groups.envs[%q].ch_id cannot be empty", envName)
				}
				if _, ok := chIDs[envConfig.CHID]; !ok {
					return fmt.Errorf("unknown handlers.error_groups.envs[%q].ch_id %q", envName, envConfig.CHID)
				}
			}
		} else {
			if cfg.Handlers.ErrorGroups.CHID == "" {
				return fmt.Errorf("handlers.error_groups.ch_id cannot be empty when envs is not used")
			}
			if _, ok := chIDs[cfg.Handlers.ErrorGroups.CHID]; !ok {
				return fmt.Errorf("unknown handlers.error_groups.ch_id %q", cfg.Handlers.ErrorGroups.CHID)
			}
		}
	}

	if cfg.Server.Auth != nil && cfg.Server.Auth.OIDC != nil {
		oidc := cfg.Server.Auth.OIDC
		hasCacheKey := oidc.CacheSecretKey != ""
		hasCacheID := oidc.CacheID != ""

		switch {
		case hasCacheKey && !hasCacheID:
			return fmt.Errorf("auth.oidc.cache_secret_key is set but auth.oidc.cache_id is empty")
		case !hasCacheKey && hasCacheID:
			return fmt.Errorf("auth.oidc.cache_id is set but auth.oidc.cache_secret_key is empty")
		case hasCacheID && hasCacheKey:
			if cfg.Cache.InmemByID(oidc.CacheID) == nil && cfg.Cache.RedisByID(oidc.CacheID) == nil {
				return fmt.Errorf("auth.oidc.cache_id %q not found", oidc.CacheID)
			}
		}
	}

	if cfg.Handlers.Admin != nil {
		if cfg.Handlers.Admin.CacheTTL <= 0 {
			cfg.Handlers.Admin.CacheTTL = defaultAdminCacheTTL
		}

		if cfg.Handlers.Admin.RedisID != "" {
			redisCfg := cfg.Cache.RedisByID(cfg.Handlers.Admin.RedisID)
			if redisCfg == nil {
				return fmt.Errorf("unknown handlers.admin.redis_id %q", cfg.Handlers.Admin.RedisID)
			}
			if redisCfg.WithInmemID != "" {
				return fmt.Errorf("handlers.admin.redis_id %q: with_inmem_id is not allowed", cfg.Handlers.Admin.RedisID)
			}
		}
	}

	return nil
}

func setSeqAPIOptionsDefaults(options *SeqAPIOptions) {
	if options.Limits.MaxAggregationsPerRequest <= 0 {
		options.Limits.MaxAggregationsPerRequest = defaultMaxAggregationsPerRequest
	}
	if options.Limits.MaxBucketsPerAggregationTs <= 0 {
		options.Limits.MaxBucketsPerAggregationTs = defaultMaxBucketsPerAggregationTs
	}
	if options.Limits.MaxParallelExportRequests <= 0 {
		options.Limits.MaxParallelExportRequests = defaultMaxParallelExportRequests
	}
	if options.Limits.MaxSearchTotal <= 0 {
		options.Limits.MaxSearchTotal = defaultMaxSearchTotal
	}
	if options.Limits.MaxSearchOffset <= 0 {
		options.Limits.MaxSearchOffset = defaultMaxSearchOffset
	}
	if options.Limits.MaxExportLimit <= 0 {
		options.Limits.MaxExportLimit = defaultMaxExportLimit
	}
	if options.Limits.MaxSearchLimit <= 0 {
		options.Limits.MaxSearchLimit = 1000 // Надо дефолт задать нам, до этого его не было
	}
	if options.Limits.SeqCLIMaxSearchLimit <= 0 {
		options.Limits.SeqCLIMaxSearchLimit = 1000 // Сюда тоже или в них не было небходимости вообще, чтобы они были по нулям
	}

	if options.Caches.TTL.Events <= 0 {
		options.Caches.TTL.Events = defaultEventsCacheTTL
	}
	if options.Caches.TTL.LogsLifespan <= 0 {
		options.Caches.TTL.LogsLifespan = defaultLogsLifespanCacheTTL
	}
}
