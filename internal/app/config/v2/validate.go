package config

import (
	"fmt"
	"time"
)

const (
	MaxHTTPHeaderBytes = 1 << 12 // 4 KiB

	DefaultSeqDBClientID = "default_seq_db"
	DefaultCHClientID    = "default_ch"
	DefaultInmemCacheID  = "default_inmem_cache"
	DefaultRedisID       = "default_redis"
	DefaultRedis2ID      = "default_redis_2"

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

	defaultGRPCConnectionTimeout = 120 * time.Second
	minGRPCKeepaliveTime         = 10 * time.Second
	minGRPCKeepaliveTimeout      = 1 * time.Second

	defaultAsyncSearchListQueryLengthLimit = 1000
	defaultCORSAllowedOrigins              = "*"

	defaultMaxSearchTotal             = 1000000
	defaultMaxSearchOffset            = 1000000
	defaultMaxExportLimit             = 100000
	defaultMaxSearchLimit             = 1000
	defaultSeqCLIMaxSearchLimit       = 1000
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

	defaultUsePreparedStatements = true
)

var defaultCORSAllowedMethods = []string{"HEAD", "GET", "POST", "PATCH", "DELETE"}

func (cfg *Config) Validate() error {
	// seqDBIDs, chIDs, err := validateClients(&cfg.Clients)
	// if err != nil {
	// 	return err
	// }
	// cacheIDs, err := validateCache(&cfg.Cache)
	// if err != nil {
	// 	return err
	// }

	validateDB(cfg.DB)

	if err := validateServer(&cfg.Server); err != nil {
		return err
	}
	// if err := validateSeqAPI(cfg.Handlers.SeqAPI, seqDBIDs, cacheIDs); err != nil {
	// 	return err
	// }
	// if err := validateAdmin(cfg.Handlers.Admin, &cfg.Cache); err != nil {
	// 	return err
	// }

	// 	if cfg.Handlers.MassExport != nil {
	// 		me := cfg.Handlers.MassExport
	//
	// 		if len(me.Envs) > 0 {
	// 			if me.SeqDBID != "" {
	// 				return fmt.Errorf("handlers.mass_export.seq_db_id must be empty when envs is used. Put seq_db_id inside each env")
	// 			}
	// 			if me.DefaultEnv == "" {
	// 				return fmt.Errorf("handlers.mass_export.default_env must be specified when envs is used")
	// 			}
	// 			if _, exists := me.Envs[me.DefaultEnv]; !exists {
	// 				return fmt.Errorf("handlers.mass_export.default_env %q not found in envs", me.DefaultEnv)
	// 			}
	//
	// 			for envName, envConfig := range me.Envs {
	// 				if envConfig.SeqDBID == "" {
	// 					return fmt.Errorf("handlers.mass_export.envs[%q].seq_db_id cannot be empty", envName)
	// 				}
	// 				if _, ok := seqDBIDs[envConfig.SeqDBID]; !ok {
	// 					return fmt.Errorf("unknown handlers.mass_export.envs[%q].seq_db_id %q", envName, envConfig.SeqDBID)
	// 				}
	//
	// 				if envConfig.Options != nil {
	// 					if envConfig.Options.SessionStore.RedisID == "" {
	// 						return fmt.Errorf("handlers.mass_export.envs[%q].options.session_store.redis_id cannot be empty", envName)
	// 					}
	// 					if _, ok := redisIDs[envConfig.Options.SessionStore.RedisID]; !ok {
	// 						return fmt.Errorf("unknown handlers.mass_export.envs[%q].options.session_store.redis_id %q", envName, envConfig.Options.SessionStore.RedisID)
	// 					}
	// 					if envConfig.Options.FileStore.S3 == nil {
	// 						return fmt.Errorf("handlers.mass_export.envs[%q].options.file_store.s3 cannot be empty", envName)
	// 					}
	// 				}
	// 			}
	// 		} else {
	// 			if me.SeqDBID == "" {
	// 				return fmt.Errorf("handlers.mass_export.seq_db_id cannot be empty when envs is not used")
	// 			}
	// 			if _, ok := seqDBIDs[me.SeqDBID]; !ok {
	// 				return fmt.Errorf("unknown handlers.mass_export.seq_db_id %q", me.SeqDBID)
	// 			}
	// 			if me.Options.SessionStore.RedisID == "" {
	// 				return fmt.Errorf("handlers.mass_export.options.session_store.redis_id cannot be empty")
	// 			}
	// 			if _, ok := redisIDs[me.Options.SessionStore.RedisID]; !ok {
	// 				return fmt.Errorf("unknown handlers.mass_export.options.session_store.redis_id %q", me.Options.SessionStore.RedisID)
	// 			}
	// 		}
	// 	}
	//
	// 	if cfg.Handlers.AsyncSearch != nil {
	// 		as := cfg.Handlers.AsyncSearch
	//
	// 		if as.Options.ListQueryLengthLimit <= 0 {
	// 			as.Options.ListQueryLengthLimit = defaultAsyncSearchListQueryLengthLimit
	// 		}
	//
	// 		if len(as.Envs) > 0 {
	// 			if as.SeqDBID != "" {
	// 				return fmt.Errorf("handlers.async_search.seq_db_id must be empty when envs is used. Put seq_db_id inside each env")
	// 			}
	// 			if as.DefaultEnv == "" {
	// 				return fmt.Errorf("handlers.async_search.default_env must be specified when envs is used")
	// 			}
	// 			if _, ok := as.Envs[as.DefaultEnv]; !ok {
	// 				return fmt.Errorf("handlers.async_search.default_env %q not found in envs", as.DefaultEnv)
	// 			}
	//
	// 			for envName, envConfig := range as.Envs {
	// 				if envConfig.SeqDBID == "" {
	// 					return fmt.Errorf("handlers.async_search.envs[%q].seq_db_id cannot be empty", envName)
	// 				}
	// 				if _, ok := seqDBIDs[envConfig.SeqDBID]; !ok {
	// 					return fmt.Errorf("unknown handlers.async_search.envs[%q].seq_db_id %q", envName, envConfig.SeqDBID)
	// 				}
	// 				if envConfig.Options.ListQueryLengthLimit <= 0 {
	// 					envConfig.Options.ListQueryLengthLimit = defaultAsyncSearchListQueryLengthLimit
	// 				}
	// 			}
	// 		} else {
	// 			if as.SeqDBID == "" {
	// 				return fmt.Errorf("handlers.async_search.seq_db_id cannot be empty when envs is not used")
	// 			}
	// 			if _, ok := seqDBIDs[as.SeqDBID]; !ok {
	// 				return fmt.Errorf("unknown handlers.async_search.seq_db_id %q", as.SeqDBID)
	// 			}
	// 		}
	// 	}
	//
	// 	if cfg.Handlers.ErrorGroups != nil {
	// 		eg := cfg.Handlers.ErrorGroups
	// 		if len(eg.Envs) > 0 {
	// 			if eg.CHID != "" {
	// 				return fmt.Errorf("handlers.error_groups.ch_id must be empty when envs is used. Put ch_id inside each env")
	// 			}
	// 			if eg.DefaultEnv == "" {
	// 				return fmt.Errorf("handlers.error_groups.default_env must be specified when envs is used")
	// 			}
	// 			if _, ok := eg.Envs[eg.DefaultEnv]; !ok {
	// 				return fmt.Errorf("handlers.error_groups.default_env %q not found in envs", eg.DefaultEnv)
	// 			}
	//
	// 			for envName, envConfig := range eg.Envs {
	// 				if envConfig.CHID == "" {
	// 					return fmt.Errorf("handlers.error_groups.envs[%q].ch_id cannot be empty", envName)
	// 				}
	// 				if _, ok := chIDs[envConfig.CHID]; !ok {
	// 					return fmt.Errorf("unknown handlers.error_groups.envs[%q].ch_id %q", envName, envConfig.CHID)
	// 				}
	// 			}
	// 		} else {
	// 			if cfg.Handlers.ErrorGroups.CHID == "" {
	// 				return fmt.Errorf("handlers.error_groups.ch_id cannot be empty when envs is not used")
	// 			}
	// 			if _, ok := chIDs[cfg.Handlers.ErrorGroups.CHID]; !ok {
	// 				return fmt.Errorf("unknown handlers.error_groups.ch_id %q", cfg.Handlers.ErrorGroups.CHID)
	// 			}
	// 		}
	// 	}

	return nil
}

func validateServer(server *Server) error {
	if server.HTTP == nil && server.GRPC == nil {
		return fmt.Errorf("server: at least one of http or grpc must be set")
	}

	if server.HTTP != nil && server.HTTP.CORS != nil {
		if len(server.HTTP.CORS.AllowedOrigins) == 0 {
			server.HTTP.CORS.AllowedOrigins = []string{defaultCORSAllowedOrigins}
		}
		if len(server.HTTP.CORS.AllowedMethods) == 0 {
			server.HTTP.CORS.AllowedMethods = defaultCORSAllowedMethods
		}
	}

	if server.GRPC != nil && server.GRPC.ConnectionTimeout <= 0 {
		server.GRPC.ConnectionTimeout = defaultGRPCConnectionTimeout
	}

	return nil
}

// func validateClients(clients *Clients) (seqDBIDs, chIDs map[string]struct{}, err error) {
// 	if len(clients.SeqDB) == 0 {
// 		return nil, nil, fmt.Errorf("clients.seq_db must contain at least one client")
// 	}
//
// 	seqDBIDs = make(map[string]struct{}, len(clients.SeqDB))
// 	for i := range clients.SeqDB {
// 		c := &clients.SeqDB[i]
// 		if c.ID == "" {
// 			return nil, nil, fmt.Errorf("clients.seq_db[%d].id cannot be empty", i)
// 		}
// 		if _, ok := seqDBIDs[c.ID]; ok {
// 			return nil, nil, fmt.Errorf("duplicate clients.seq_db.id %q", c.ID)
// 		}
//
// 		seqDBIDs[c.ID] = struct{}{}
//
// 		if len(c.Addrs) == 0 {
// 			return nil, nil, fmt.Errorf("clients.seq_db[%q].addrs cannot be empty", c.ID)
// 		}
// 		if c.Timeout <= 0 {
// 			return nil, nil, fmt.Errorf("clients.seq_db[%q].timeout must be > 0", c.ID)
// 		}
// 		if c.ClientMode == "" {
// 			c.ClientMode = SeqDBClientModeGRPC
// 		} else if c.ClientMode != SeqDBClientModeGRPC {
// 			return nil, nil, fmt.Errorf("invalid clients.seq_db[%q].client_mode: %q (allowed: %q)", c.ID, c.ClientMode, SeqDBClientModeGRPC)
// 		}
//
// 		if c.GRPCParams != nil {
// 			if c.GRPCParams.Keepalive != nil {
// 				c.GRPCParams.Keepalive.Time = max(c.GRPCParams.Keepalive.Time, minGRPCKeepaliveTime)
// 				c.GRPCParams.Keepalive.Timeout = max(c.GRPCParams.Keepalive.Timeout, minGRPCKeepaliveTimeout)
// 			}
// 		}
// 	}
//
// 	chIDs = make(map[string]struct{}, len(clients.ClickHouse))
// 	for i := range clients.ClickHouse {
// 		ch := &clients.ClickHouse[i]
// 		if ch.ID == "" {
// 			return nil, nil, fmt.Errorf("clients.clickhouse[%d].id cannot be empty", i)
// 		}
// 		if _, ok := chIDs[ch.ID]; ok {
// 			return nil, nil, fmt.Errorf("duplicate clients.clickhouse.id %q", ch.ID)
// 		}
//
// 		chIDs[ch.ID] = struct{}{}
//
// 		if ch.DialTimeout <= 0 {
// 			ch.DialTimeout = defaultClickHouseDialTimeout
// 		}
// 		if ch.ReadTimeout <= 0 {
// 			ch.ReadTimeout = defaultClickHouseReadTimeout
// 		}
// 	}
//
// 	return seqDBIDs, chIDs, nil
// }
//
// func validateCache(cache *Cache) (map[string]struct{}, error) {
// 	inmemIDs := make(map[string]struct{}, len(cache.Inmemory))
// 	for i := range cache.Inmemory {
// 		inm := &cache.Inmemory[i]
// 		if inm.ID == "" {
// 			return nil, fmt.Errorf("cache.inmemory[%d].id cannot be empty", i)
// 		}
// 		if _, ok := inmemIDs[inm.ID]; ok {
// 			return nil, fmt.Errorf("duplicate cache.inmemory.id %q", inm.ID)
// 		}
//
// 		inmemIDs[inm.ID] = struct{}{}
//
// 		if inm.NumCounters <= 0 {
// 			inm.NumCounters = defaultInmemCacheNumCounters
// 		}
// 		if inm.MaxCost <= 0 {
// 			inm.MaxCost = defaultInmemCacheMaxCost
// 		}
// 		if inm.BufferItems <= 0 {
// 			inm.BufferItems = defaultInmemCacheBufferItems
// 		}
// 	}
//
// 	cacheIDs := make(map[string]struct{}, len(cache.Inmemory)+len(cache.Redis))
// 	for i := range cache.Redis {
// 		r := &cache.Redis[i]
// 		if r.ID == "" {
// 			return nil, fmt.Errorf("cache.redis[%d].id cannot be empty", i)
// 		}
// 		if _, ok := cacheIDs[r.ID]; ok {
// 			return nil, fmt.Errorf("duplicate cache.redis.id %q", r.ID)
// 		}
//
// 		cacheIDs[r.ID] = struct{}{}
//
// 		if r.WithInmemID != "" {
// 			if _, ok := inmemIDs[r.WithInmemID]; !ok {
// 				return nil, fmt.Errorf("cache.redis[%q].with_inmem_id %q not found in cache.inmemory", r.ID, r.WithInmemID)
// 			}
// 		}
//
// 		for id := range inmemIDs {
// 			if !slices.ContainsFunc(cache.Redis, func(r Redis) bool {
// 				return r.WithInmemID == id
// 			}) {
// 				cacheIDs[id] = struct{}{}
// 			}
// 		}
// 	}
//
// 	return cacheIDs, nil
// }

func validateDB(db *DB) {
	if db != nil && db.UsePreparedStatements == nil {
		v := defaultUsePreparedStatements
		db.UsePreparedStatements = &v
	}
}

// func validateSeqAPI(seqAPI *SeqAPI, seqDBIDs, cacheIDs map[string]struct{}) error {
// 	setSeqAPIOptionsDefaults(&seqAPI.Options)
//
// 	_ = cacheIDs // Проверки потом нафигачу на кэши, пусть весит пока голенький
// 	if len(seqAPI.Envs) > 0 {
// 		if seqAPI.SeqDBID != "" {
// 			return fmt.Errorf("handlers.seq_api.seq_db_id must be empty when envs is used. Put seq_db_id inside each env")
// 		}
// 		if seqAPI.DefaultEnv == "" {
// 			return fmt.Errorf("handlers.seq_api.default_env must be specified when envs is used")
// 		}
// 		if _, exists := seqAPI.Envs[seqAPI.DefaultEnv]; !exists {
// 			return fmt.Errorf("handlers.seq_api.default_env %q not found in envs", seqAPI.DefaultEnv)
// 		}
//
// 		for envName, envConfig := range seqAPI.Envs {
// 			if envConfig.SeqDBID == "" {
// 				return fmt.Errorf("handlers.seq_api.envs[%q].seq_db_id cannot be empty", envName)
// 			}
// 			if _, ok := seqDBIDs[envConfig.SeqDBID]; !ok {
// 				return fmt.Errorf("unknown handlers.seq_api.envs[%q].seq_db_id %q", envName, envConfig.SeqDBID)
// 			}
// 			setSeqAPIOptionsDefaults(envConfig.Options)
// 		}
// 	} else {
// 		if seqAPI.SeqDBID == "" {
// 			return fmt.Errorf("handlers.seq_api.seq_db_id cannot be empty when envs is not used")
// 		}
// 		if _, ok := seqDBIDs[seqAPI.SeqDBID]; !ok {
// 			return fmt.Errorf("unknown handlers.seq_api.seq_db_id %q", seqAPI.SeqDBID)
// 		}
// 	}
//
// 	return nil
// }

// func validateAdmin(admin *Admin, cache *Cache) error {
// 	if admin == nil {
// 		return nil
// 	}
//
// 	if admin.Options.Cache != nil {
// 		if admin.Options.Cache.ID == "" {
// 			return fmt.Errorf("handlers.admin.options.cache.id cannot be empty")
// 		}
// 		redisCfg := cache.RedisByID(admin.Options.Cache.ID)
// 		if redisCfg == nil {
// 			return fmt.Errorf("unknown handlers.admin.options.cache.id %q", admin.Options.Cache.ID)
// 		}
// 		if redisCfg == nil {
// 			return fmt.Errorf("unknown handlers.admin.options.cache.id %q", admin.Options.Cache.ID)
// 		}
// 		if redisCfg.WithInmemID != "" {
// 			return fmt.Errorf("handlers.admin.options.cache.id %q: with_inmem_id is not allowed", admin.Options.Cache.ID)
// 		}
// 		if admin.Options.Cache.TTL <= 0 {
// 			admin.Options.Cache.TTL = defaultAdminCacheTTL
// 		}
// 	}
//
// 	if len(admin.Envs) > 0 {
//
// 	}
// }
//
// func setSeqAPIOptionsDefaults(options *SeqAPIOptions) {
// 	if options.Limits.MaxAggregationsPerRequest <= 0 {
// 		options.Limits.MaxAggregationsPerRequest = defaultMaxAggregationsPerRequest
// 	}
// 	if options.Limits.MaxBucketsPerAggregationTs <= 0 {
// 		options.Limits.MaxBucketsPerAggregationTs = defaultMaxBucketsPerAggregationTs
// 	}
// 	if options.Limits.MaxParallelExportRequests <= 0 {
// 		options.Limits.MaxParallelExportRequests = defaultMaxParallelExportRequests
// 	}
// 	if options.Limits.MaxSearchTotal <= 0 {
// 		options.Limits.MaxSearchTotal = defaultMaxSearchTotal
// 	}
// 	if options.Limits.MaxSearchOffset <= 0 {
// 		options.Limits.MaxSearchOffset = defaultMaxSearchOffset
// 	}
// 	if options.Limits.MaxExportLimit <= 0 {
// 		options.Limits.MaxExportLimit = defaultMaxExportLimit
// 	}
// 	if options.Limits.MaxSearchLimit <= 0 {
// 		options.Limits.MaxSearchLimit = defaultMaxSearchLimit
// 	}
// 	if options.Limits.SeqCLIMaxSearchLimit <= 0 {
// 		options.Limits.SeqCLIMaxSearchLimit = defaultSeqCLIMaxSearchLimit
// 	}
// 	if options.Caches.Events.TTL <= 0 {
// 		options.Caches.Events.TTL = defaultEventsCacheTTL
// 	}
// 	if options.Caches.LogsLifespan.TTL <= 0 {
// 		options.Caches.LogsLifespan.TTL = defaultLogsLifespanCacheTTL
// 	}
// }
