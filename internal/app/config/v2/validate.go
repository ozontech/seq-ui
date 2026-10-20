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

	defaultEventsCacheTTL = 24 * time.Hour

	defaultLogsLifespanCacheTTL = 10 * time.Minute

	defaultAdminCacheTTL = time.Hour

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
	seqDBIDs, chIDs, err := validateClients(&cfg.Clients)
	if err != nil {
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
	if err := validateSeqAPI(&cfg.Handlers.SeqAPI, seqDBIDs, cacheIDs); err != nil {
		return err
	}
	if err := validateAsyncSearch(cfg.Handlers.AsyncSearch, seqDBIDs); err != nil {
		return err
	}
	if err := validateMassExport(cfg.Handlers.MassExport, seqDBIDs, &cfg.Cache); err != nil {
		return err
	}
	if err := validateErrorGroups(cfg.Handlers.ErrorGroups, chIDs); err != nil {
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

func validateClients(clients *Clients) (seqDBIDs, chIDs map[string]struct{}, err error) {
	if len(clients.SeqDB) == 0 {
		return nil, nil, fmt.Errorf("clients.seq_db must contain at least one client")
	}

	seqDBIDs = make(map[string]struct{}, len(clients.SeqDB))
	for i := range clients.SeqDB {
		c := &clients.SeqDB[i]
		if c.ID == "" {
			return nil, nil, fmt.Errorf("clients.seq_db[%d].id cannot be empty", i)
		}
		if _, ok := seqDBIDs[c.ID]; ok {
			return nil, nil, fmt.Errorf("duplicate clients.seq_db.id %q", c.ID)
		}
		seqDBIDs[c.ID] = struct{}{}

		if len(c.Addrs) == 0 {
			return nil, nil, fmt.Errorf("clients.seq_db[%q].addrs cannot be empty", c.ID)
		}
		if c.Timeout <= 0 {
			return nil, nil, fmt.Errorf("clients.seq_db[%q].timeout must be > 0", c.ID)
		}
		if c.ClientMode == "" {
			c.ClientMode = SeqDBClientModeGRPC
		} else if c.ClientMode != SeqDBClientModeGRPC {
			return nil, nil, fmt.Errorf("invalid clients.seq_db[%q].client_mode: %q (allowed: %q)", c.ID, c.ClientMode, SeqDBClientModeGRPC)
		}

		if c.GRPCParams != nil && c.GRPCParams.Keepalive != nil {
			c.GRPCParams.Keepalive.Time = max(c.GRPCParams.Keepalive.Time, minGRPCKeepaliveTime)
			c.GRPCParams.Keepalive.Timeout = max(c.GRPCParams.Keepalive.Timeout, minGRPCKeepaliveTimeout)
		}
	}

	chIDs = make(map[string]struct{}, len(clients.ClickHouse))
	for i := range clients.ClickHouse {
		ch := &clients.ClickHouse[i]
		if ch.ID == "" {
			return nil, nil, fmt.Errorf("clients.clickhouse[%d].id cannot be empty", i)
		}
		if _, ok := chIDs[ch.ID]; ok {
			return nil, nil, fmt.Errorf("duplicate clients.clickhouse.id %q", ch.ID)
		}
		chIDs[ch.ID] = struct{}{}

		if len(ch.Addrs) == 0 {
			return nil, nil, fmt.Errorf("clients.clickhouse[%q].addrs cannot be empty", ch.ID)
		}
		if ch.DialTimeout <= 0 {
			ch.DialTimeout = defaultClickHouseDialTimeout
		}
		if ch.ReadTimeout <= 0 {
			ch.ReadTimeout = defaultClickHouseReadTimeout
		}
	}

	return seqDBIDs, chIDs, nil
}

func validateCache(cache *Cache) (map[string]struct{}, error) {
	cacheIDs := make(map[string]struct{}, len(cache.Inmemory)+len(cache.Redis))

	for i := range cache.Inmemory {
		inm := &cache.Inmemory[i]
		if inm.ID == "" {
			return nil, fmt.Errorf("cache.inmemory[%d].id cannot be empty", i)
		}
		if _, ok := cacheIDs[inm.ID]; ok {
			return nil, fmt.Errorf("duplicate cache id %q", inm.ID)
		}
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

	for i := range cache.Redis {
		r := &cache.Redis[i]
		if r.ID == "" {
			return nil, fmt.Errorf("cache.redis[%d].id cannot be empty", i)
		}
		if _, ok := cacheIDs[r.ID]; ok {
			return nil, fmt.Errorf("duplicate cache id %q", r.ID)
		}
		if r.Addr == "" {
			return nil, fmt.Errorf("cache.redis[%q].addr cannot be empty", r.ID)
		}
		if r.WithInmemID != "" && cache.InmemByID(r.WithInmemID) == nil {
			return nil, fmt.Errorf("cache.redis[%q].with_inmem_id %q not found in cache.inmemory", r.ID, r.WithInmemID)
		}
		cacheIDs[r.ID] = struct{}{}
	}

	return cacheIDs, nil
}

func validateDB(db *DB) {
	if db != nil && db.UsePreparedStatements == nil {
		v := defaultUsePreparedStatements
		db.UsePreparedStatements = &v
	}
}

func validateSeqAPI(seqAPI *SeqAPI, seqDBIDs, cacheIDs map[string]struct{}) error {
	if err := validateEnvs("handlers.seq_api", seqAPI.Envs, seqAPI.DefaultEnv); err != nil {
		return err
	}
	if len(seqAPI.Envs) == 0 && len(seqDBIDs) != 1 {
		return fmt.Errorf("handlers.seq_api: clients.seq_db must contain one client when envs are not used")
	}
	if err := validateSeqAPIOptions("handlers.seq_api.options", &seqAPI.Options, cacheIDs); err != nil {
		return err
	}

	for envName, envConfig := range seqAPI.Envs {
		if envConfig.SeqDBID == "" {
			return fmt.Errorf("handlers.seq_api.envs[%q].seq_db_id cannot be empty", envName)
		}
		if _, ok := seqDBIDs[envConfig.SeqDBID]; !ok {
			return fmt.Errorf("unknown handlers.seq_api.envs[%q].seq_db_id %q", envName, envConfig.SeqDBID)
		}

		// options are nil only if root options are not set, so env uses default options
		if envConfig.Options == nil {
			envConfig.Options = &SeqAPIOptions{}
			seqAPI.Envs[envName] = envConfig
		}
		if err := validateSeqAPIOptions(fmt.Sprintf("handlers.seq_api.envs[%q].options", envName), envConfig.Options, cacheIDs); err != nil {
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
	if err := validateHandlerCache("handlers.admin.options.cache", admin.Options.Cache, cacheIDs, defaultAdminCacheTTL); err != nil {
		return err
	}

	for envName, envConfig := range admin.Envs {
		if envConfig.Options == nil {
			envConfig.Options = &AdminOptions{}
			admin.Envs[envName] = envConfig
		}

		if err := validateHandlerCache(fmt.Sprintf("handlers.admin.envs[%q].options.cache", envName), envConfig.Options.Cache, cacheIDs, defaultAdminCacheTTL); err != nil {
			return err
		}
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

func validateAsyncSearch(asyncSearch *AsyncSearch, seqDBIDs map[string]struct{}) error {
	if asyncSearch == nil {
		return nil
	}
	if err := validateEnvs("handlers.async_search", asyncSearch.Envs, asyncSearch.DefaultEnv); err != nil {
		return err
	}
	if asyncSearch.Options.ListQueryLengthLimit <= 0 {
		asyncSearch.Options.ListQueryLengthLimit = defaultAsyncSearchListQueryLengthLimit
	}

	for envName, envConfig := range asyncSearch.Envs {
		if envConfig.SeqDBID == "" {
			return fmt.Errorf("handlers.async_search.envs[%q].seq_db_id cannot be empty", envName)
		}
		if _, ok := seqDBIDs[envConfig.SeqDBID]; !ok {
			return fmt.Errorf("unknown handlers.async_search.envs[%q].seq_db_id %q", envName, envConfig.SeqDBID)
		}
		if envConfig.Options == nil {
			envConfig.Options = &AsyncSearchOptions{}
			asyncSearch.Envs[envName] = envConfig
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
	if err := validateHandlerCache(path+".caches.events", caches.Events, cacheIDs, defaultEventsCacheTTL); err != nil {
		return err
	}
	if err := validateHandlerCache(path+".caches.logs_lifespan", caches.LogsLifespan, cacheIDs, defaultLogsLifespanCacheTTL); err != nil {
		return err
	}

	return validateHandlerCache(path+".caches.fields", caches.Fields, cacheIDs, 0)
}

func validateHandlerCache(path string, hc *HandlerCache, cacheIDs map[string]struct{}, defaultTTL time.Duration) error {
	if hc == nil {
		return nil
	}
	if _, ok := cacheIDs[hc.ID]; !ok {
		return fmt.Errorf("%s.id %q not found in cache", path, hc.ID)
	}
	if hc.TTL == nil {
		hc.TTL = new(defaultTTL)
	} else if *hc.TTL < 0 {
		return fmt.Errorf("%s.ttl must be >= 0", path)
	}

	return nil
}

func validateErrorGroups(errorGroups *ErrorGroups, chIDs map[string]struct{}) error {
	if errorGroups == nil {
		return nil
	}
	if err := validateEnvs("handlers.error_groups", errorGroups.Envs, errorGroups.DefaultEnv); err != nil {
		return err
	}
	if len(errorGroups.Envs) == 0 && len(chIDs) != 1 {
		return fmt.Errorf("handlers.error_groups: clients must contain one client when envs are not used")
	}
	for envName, envConfig := range errorGroups.Envs {
		if envConfig.CHID == "" {
			return fmt.Errorf("handlers.error_groups.envs[%q].ch_id cannot be empty", envName)
		}
		if _, ok := chIDs[envConfig.CHID]; !ok {
			return fmt.Errorf("unknown handlers.error_groups.envs[%q].ch_id %q", envName, envConfig.CHID)
		}
		if envConfig.Options == nil {
			envConfig.Options = &ErrorGroupsOptions{}
			errorGroups.Envs[envName] = envConfig
		}
	}

	return nil
}

func validateMassExport(massExport *MassExport, seqDBIDs map[string]struct{}, cache *Cache) error {
	if massExport == nil {
		return nil
	}
	if err := validateEnvs("handlers.mass_export", massExport.Envs, massExport.DefaultEnv); err != nil {
		return err
	}
	if len(massExport.Envs) == 0 {
		return validateMassExportOptions("handlers.mass_export.options", &massExport.Options, cache, "")
	}

	for envName, envConfig := range massExport.Envs {
		if envConfig.SeqDBID == "" {
			return fmt.Errorf("handlers.mass_export.envs[%q].seq_db_id cannot be empty", envName)
		}
		if _, ok := seqDBIDs[envConfig.SeqDBID]; !ok {
			return fmt.Errorf("unknown handlers.mass_export.envs[%q].seq_db_id %q", envName, envConfig.SeqDBID)
		}
		if envConfig.Options == nil {
			envConfig.Options = &MassExportOptions{}
			massExport.Envs[envName] = envConfig
		}
		if err := validateMassExportOptions(fmt.Sprintf("handlers.mass_export.envs[%q].options", envName), envConfig.Options, cache, envName); err != nil {
			return err
		}
	}

	return nil
}

func validateMassExportOptions(path string, options *MassExportOptions, cache *Cache, env string) error {
	if options.FileStore.S3 == nil {
		return fmt.Errorf("%s.file_store.s3 cannot be empty", path)
	}

	redis := cache.RedisByID(options.SessionStore.CacheID)
	if redis == nil || redis.WithInmemID != "" {
		return fmt.Errorf("%s.session_store.cache_id %q must be cache.redis without with_inmem_id", path, options.SessionStore.CacheID)
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
