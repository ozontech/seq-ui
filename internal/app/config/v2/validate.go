package config

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/ozontech/seq-ui/logger"
)

const (
	MaxHTTPHeaderBytes = 1 << 12 // 4 KiB

	DefaultSeqDBClientID       = "default_seq_db"
	DefaultCHClientID          = "default_ch"
	DefaultInmemID             = "default_inmem"
	DefaultRedisWithInmemID    = "default_redis_with_inmem"
	DefaultRedisWithoutInmemID = "default_redis_without_inmem"
	DefaultMassExportRedisID   = "default_mass_export_redis"

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

	defaultClickHouseDialTimeout = 5 * time.Second
	defaultClickHouseReadTimeout = 30 * time.Second

	defaultUsePreparedStatements = true

	day                    = 24 * time.Hour
	defaultSessionLifetime = 7 * day
	minSessionLifetime     = 1 * day
	maxSessionLifetime     = 30 * day
)

var defaultCORSAllowedMethods = []string{"HEAD", "GET", "POST", "PATCH", "DELETE"}

func (cfg *Config) Validate() error {
	if err := validateClients(&cfg.Clients); err != nil {
		return err
	}
	cacheIDs, err := validateCache(&cfg.Cache)
	if err != nil {
		return err
	}

	validateDB(cfg.DB)

	if err := validateServer(&cfg.Server, cacheIDs); err != nil {
		return err
	}
	if err := validateSeqAPI(&cfg.Handlers.SeqAPI, cfg.Clients.SeqDB, cacheIDs); err != nil {
		return err
	}
	if err := validateAsyncSearch(cfg.Handlers.AsyncSearch, cfg.Clients.SeqDB); err != nil {
		return err
	}
	if err := validateMassExport(cfg.Handlers.MassExport, cfg.Clients.SeqDB, &cfg.Cache); err != nil {
		return err
	}
	if err := validateErrorGroups(cfg.Handlers.ErrorGroups, cfg.Clients.ClickHouse); err != nil {
		return err
	}
	if err := validateAdmin(cfg.Handlers.Admin, cacheIDs); err != nil {
		return err
	}

	return nil
}

func validateServer(server *Server, cacheIDs map[string]struct{}) error {
	if server.HTTP == nil && server.GRPC == nil {
		return fmt.Errorf("server: at least one of http or grpc must be set")
	}

	if server.HTTP != nil {
		setCORSDefaults(server.HTTP.CORS)
	}
	setCORSDefaults(server.Debug.CORS)

	if server.GRPC != nil && server.GRPC.ConnectionTimeout <= 0 {
		server.GRPC.ConnectionTimeout = defaultGRPCConnectionTimeout
	}

	if server.Auth != nil {
		if err := validateOIDC("server.auth.options.oidc", server.Auth.Options.OIDC, cacheIDs); err != nil {
			return err
		}
		for envName, envCfg := range server.Auth.Envs {
			if err := validateOIDC(fmt.Sprintf("server.auth.envs[%q].oidc", envName), envCfg.OIDC, cacheIDs); err != nil {
				return err
			}
		}
	}

	return nil
}

func setCORSDefaults(cors *CORS) {
	if cors == nil {
		return
	}

	if len(cors.AllowedOrigins) == 0 {
		cors.AllowedOrigins = []string{defaultCORSAllowedOrigins}
	}
	if len(cors.AllowedMethods) == 0 {
		cors.AllowedMethods = defaultCORSAllowedMethods
	}
}

func validateOIDC(path string, oidc *OIDC, cacheIDs map[string]struct{}) error {
	if oidc == nil {
		return nil
	}
	if len(oidc.AuthURLs) == 0 {
		return fmt.Errorf("%s.auth_urls cannot be empty", path)
	}
	if (oidc.CacheID == "") != (oidc.CacheSecretKey == "") {
		return fmt.Errorf("%s.cache_id and cache_secret_key must be set together", path)
	}
	if oidc.CacheID == "" {
		return nil
	}
	if _, ok := cacheIDs[oidc.CacheID]; !ok {
		return fmt.Errorf("%s.cache_id %q not found in cache", path, oidc.CacheID)
	}
	return nil
}

func validateClients(clients *Clients) error {
	if len(clients.SeqDB) == 0 {
		return fmt.Errorf("clients.seq_db must contain at least one client")
	}

	for id, c := range clients.SeqDB {
		if c == nil {
			return fmt.Errorf("clients.seq_db[%q] cannot be empty", id)
		}
		c.ID = id

		if len(c.Addrs) == 0 {
			return fmt.Errorf("clients.seq_db[%q].addrs cannot be empty", id)
		}
		if c.Timeout <= 0 {
			return fmt.Errorf("clients.seq_db[%q].timeout must be > 0", id)
		}
		if c.ClientMode == "" {
			c.ClientMode = SeqDBClientModeGRPC
		} else if c.ClientMode != SeqDBClientModeGRPC {
			return fmt.Errorf("invalid clients.seq_db[%q].client_mode: %q (allowed: %q)", id, c.ClientMode, SeqDBClientModeGRPC)
		}

		if c.GRPCParams != nil && c.GRPCParams.Keepalive != nil {
			c.GRPCParams.Keepalive.Time = max(c.GRPCParams.Keepalive.Time, minGRPCKeepaliveTime)
			c.GRPCParams.Keepalive.Timeout = max(c.GRPCParams.Keepalive.Timeout, minGRPCKeepaliveTimeout)
		}
	}

	for id, ch := range clients.ClickHouse {
		if ch == nil {
			return fmt.Errorf("clients.clickhouse[%q] cannot be empty", id)
		}
		ch.ID = id

		if len(ch.Addrs) == 0 {
			return fmt.Errorf("clients.clickhouse[%q].addrs cannot be empty", id)
		}
		if ch.DialTimeout <= 0 {
			ch.DialTimeout = defaultClickHouseDialTimeout
		}
		if ch.ReadTimeout <= 0 {
			ch.ReadTimeout = defaultClickHouseReadTimeout
		}
	}

	return nil
}

func validateCache(cache *Cache) (map[string]struct{}, error) {
	cacheIDs := make(map[string]struct{}, len(cache.Inmemory)+len(cache.Redis))

	for id, inm := range cache.Inmemory {
		if id == "" {
			return nil, fmt.Errorf("cache.inmemory: id cannot be empty")
		}
		if inm == nil {
			inm = &Inmemory{}
			cache.Inmemory[id] = inm
		}
		inm.ID = id
		cacheIDs[inm.ID] = struct{}{}

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

	for id, r := range cache.Redis {
		if id == "" {
			return nil, fmt.Errorf("cache.redis: id cannot be empty")
		}
		if _, ok := cacheIDs[id]; ok {
			return nil, fmt.Errorf("duplicate cache id %q in cache.inmemory and cache.redis", id)
		}
		if r == nil {
			return nil, fmt.Errorf("cache.redis[%q] cannot be empty", id)
		}
		r.ID = id

		if r.Addr == "" {
			return nil, fmt.Errorf("cache.redis[%q].addr cannot be empty", id)
		}
		if _, ok := cache.Inmemory[r.WithInmemID]; r.WithInmemID != "" && !ok {
			return nil, fmt.Errorf("cache.redis[%q].with_inmem_id %q not found in cache.inmemory", id, r.WithInmemID)
		}
		cacheIDs[id] = struct{}{}
	}

	return cacheIDs, nil
}

func validateDB(db *DB) {
	if db != nil && db.UsePreparedStatements == nil {
		v := defaultUsePreparedStatements
		db.UsePreparedStatements = &v
	}
}

func validateSeqAPI(seqAPI *SeqAPI, seqDBs map[string]*SeqDBClient, cacheIDs map[string]struct{}) error {
	if err := validateEnvs("handlers.seq_api", seqAPI.Envs, seqAPI.DefaultEnv); err != nil {
		return err
	}
	if len(seqAPI.Envs) == 0 {
		if err := validateClientID("handlers.seq_api.options.seq_db_id", seqAPI.Options.SeqDBID, seqDBs); err != nil {
			return err
		}
	}
	if err := validateSeqAPIOptions("handlers.seq_api.options", &seqAPI.Options, cacheIDs); err != nil {
		return err
	}

	for envName, envConfig := range seqAPI.Envs {
		// options are nil only if root options are not set, so env uses default options
		if envConfig.Options == nil {
			envConfig.Options = &SeqAPIOptions{}
			seqAPI.Envs[envName] = envConfig
		}

		path := fmt.Sprintf("handlers.seq_api.envs[%q].options", envName)
		if err := validateClientID(path+".seq_db_id", envConfig.Options.SeqDBID, seqDBs); err != nil {
			return err
		}
		if err := validateSeqAPIOptions(path, envConfig.Options, cacheIDs); err != nil {
			return err
		}
	}

	return nil
}

func validateAdmin(admin *Admin, cacheIDs map[string]struct{}) error {
	if admin == nil {
		return nil
	}
	if err := validateEnvs("handlers.admin", admin.Envs, admin.DefaultEnv); err != nil {
		return err
	}
	if err := validateHandlerCache("handlers.admin.options.cache", admin.Options.Cache, cacheIDs); err != nil {
		return err
	}

	for envName, envConfig := range admin.Envs {
		if envConfig.Options == nil {
			envConfig.Options = &AdminOptions{}
			admin.Envs[envName] = envConfig
		}

		if err := validateHandlerCache(fmt.Sprintf("handlers.admin.envs[%q].options.cache", envName), envConfig.Options.Cache, cacheIDs); err != nil {
			return err
		}
	}

	return nil
}

func validateClientID[T any](path, id string, clients map[string]T) error {
	if id == "" {
		return fmt.Errorf("%s cannot be empty", path)
	}
	if _, ok := clients[id]; !ok {
		return fmt.Errorf("unknown %s %q", path, id)
	}
	return nil
}

func validateEnvs[T any](path string, envs map[string]T, defaultEnv string) error {
	if len(envs) == 0 {
		return nil
	}
	if defaultEnv == "" {
		return fmt.Errorf("%s.default_env must be set when envs are used", path)
	}
	if _, ok := envs[defaultEnv]; !ok {
		return fmt.Errorf("%s.default_env %q not found in envs", path, defaultEnv)
	}
	return nil
}

func validateAsyncSearch(asyncSearch *AsyncSearch, seqDBs map[string]*SeqDBClient) error {
	if asyncSearch == nil {
		return nil
	}
	if err := validateEnvs("handlers.async_search", asyncSearch.Envs, asyncSearch.DefaultEnv); err != nil {
		return err
	}
	if len(asyncSearch.Envs) == 0 {
		if err := validateClientID("handlers.async_search.options.seq_db_id", asyncSearch.Options.SeqDBID, seqDBs); err != nil {
			return err
		}
	}
	if asyncSearch.Options.ListQueryLengthLimit <= 0 {
		asyncSearch.Options.ListQueryLengthLimit = defaultAsyncSearchListQueryLengthLimit
	}

	for envName, envConfig := range asyncSearch.Envs {
		if envConfig.Options == nil {
			envConfig.Options = &AsyncSearchOptions{}
			asyncSearch.Envs[envName] = envConfig
		}
		if err := validateClientID(
			fmt.Sprintf("handlers.async_search.envs[%q].options.seq_db_id", envName),
			envConfig.Options.SeqDBID,
			seqDBs); err != nil {
			return err
		}
		if envConfig.Options.ListQueryLengthLimit <= 0 {
			envConfig.Options.ListQueryLengthLimit = defaultAsyncSearchListQueryLengthLimit
		}
	}

	return nil
}

func validateSeqAPIOptions(path string, options *SeqAPIOptions, cacheIDs map[string]struct{}) error {
	limits := &options.Limits
	if limits.MaxAggregationsPerRequest <= 0 {
		limits.MaxAggregationsPerRequest = defaultMaxAggregationsPerRequest
	}
	if limits.MaxBucketsPerAggregationTs <= 0 {
		limits.MaxBucketsPerAggregationTs = defaultMaxBucketsPerAggregationTs
	}
	if limits.MaxParallelExportRequests <= 0 {
		limits.MaxParallelExportRequests = defaultMaxParallelExportRequests
	}
	if limits.MaxSearchTotal <= 0 {
		limits.MaxSearchTotal = defaultMaxSearchTotal
	}
	if limits.MaxSearchOffset <= 0 {
		limits.MaxSearchOffset = defaultMaxSearchOffset
	}
	if limits.MaxExportLimit <= 0 {
		limits.MaxExportLimit = defaultMaxExportLimit
	}
	if limits.MaxSearchLimit <= 0 {
		limits.MaxSearchLimit = defaultMaxSearchLimit
	}
	if limits.SeqCLIMaxSearchLimit <= 0 {
		limits.SeqCLIMaxSearchLimit = defaultSeqCLIMaxSearchLimit
	}

	caches := &options.Caches
	if err := validateHandlerCache(path+".caches.events", caches.Events, cacheIDs); err != nil {
		return err
	}
	if err := validateHandlerCache(path+".caches.logs_lifespan", caches.LogsLifespan, cacheIDs); err != nil {
		return err
	}

	return validateHandlerCache(path+".caches.fields", caches.Fields, cacheIDs)
}

func validateHandlerCache(path string, hc *HandlerCache, cacheIDs map[string]struct{}) error {
	if hc == nil {
		return nil
	}
	if _, ok := cacheIDs[hc.ID]; !ok {
		return fmt.Errorf("%s.id %q not found in cache", path, hc.ID)
	}
	if hc.TTL < 0 {
		return fmt.Errorf("%s.ttl must be >= 0", path)
	}

	return nil
}

func validateErrorGroups(errorGroups *ErrorGroups, chs map[string]*CHClient) error {
	if errorGroups == nil {
		return nil
	}
	if err := validateEnvs("handlers.error_groups", errorGroups.Envs, errorGroups.DefaultEnv); err != nil {
		return err
	}
	if len(errorGroups.Envs) == 0 {
		if err := validateClientID("handlers.error_groups.options.ch_id", errorGroups.Options.CHID, chs); err != nil {
			return err
		}
	}
	for envName, envConfig := range errorGroups.Envs {
		if envConfig.Options == nil {
			envConfig.Options = &ErrorGroupsOptions{}
			errorGroups.Envs[envName] = envConfig
		}
		if err := validateClientID(fmt.Sprintf("handlers.error_groups.envs[%q].options.ch_id", envName), envConfig.Options.CHID, chs); err != nil {
			return err
		}
	}

	return nil
}

func validateMassExport(massExport *MassExport, seqDBs map[string]*SeqDBClient, cache *Cache) error {
	if massExport == nil {
		return nil
	}
	if err := validateEnvs("handlers.mass_export", massExport.Envs, massExport.DefaultEnv); err != nil {
		return err
	}
	if len(massExport.Envs) == 0 {
		if err := validateClientID("handlers.mass_export.options.seq_db_id", massExport.Options.SeqDBID, seqDBs); err != nil {
			return err
		}
		return validateMassExportOptions("handlers.mass_export.options", &massExport.Options, cache, "")
	}

	for envName, envConfig := range massExport.Envs {
		if envConfig.Options == nil {
			envConfig.Options = &MassExportOptions{}
			massExport.Envs[envName] = envConfig
		}

		path := fmt.Sprintf("handlers.mass_export.envs[%q].options", envName)
		if err := validateClientID(path+".seq_db_id", envConfig.Options.SeqDBID, seqDBs); err != nil {
			return err
		}
		if err := validateMassExportOptions(path, envConfig.Options, cache, envName); err != nil {
			return err
		}
	}

	return nil
}

func validateMassExportOptions(path string, options *MassExportOptions, cache *Cache, env string) error {
	if options.FileStore.S3 == nil {
		return fmt.Errorf("%s.file_store.s3 cannot be empty", path)
	}

	redis := cache.Redis[options.SessionStore.RedisCacheID]
	if redis == nil || redis.WithInmemID != "" {
		return fmt.Errorf("%s.session_store.cache_id %q must be cache.redis without with_inmem_id", path, options.SessionStore.RedisCacheID)
	}

	setSessionStoreDefaults(&options.SessionStore, env)

	return nil
}

func setSessionStoreDefaults(sessionStore *SessionStore, env string) {
	switch {
	case sessionStore.Lifetime == 0:
		sessionStore.Lifetime = defaultSessionLifetime
	case sessionStore.Lifetime < minSessionLifetime:
		logger.Warn("export lifetime from config is too low; set min lifetime",
			zap.String("env", env),
			zap.String("export_lifetime", sessionStore.Lifetime.String()),
		)
		sessionStore.Lifetime = minSessionLifetime
	case sessionStore.Lifetime > maxSessionLifetime:
		logger.Warn("export lifetime from config is too high; set max lifetime",
			zap.String("env", env),
			zap.String("export_lifetime", sessionStore.Lifetime.String()),
		)
		sessionStore.Lifetime = maxSessionLifetime
	}
}
